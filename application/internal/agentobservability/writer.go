package agentobservability

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontract"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

// Mode controls observability behavior for agent ledger/logs/contract writes.
type Mode string

const (
	ModeBestEffort Mode = config.AgentObservabilityModeBestEffort
	ModeWarn       Mode = config.AgentObservabilityModeWarn
	ModeStrict     Mode = config.AgentObservabilityModeStrict
)

// Writer centralizes agent observability writes with configurable failure handling.
type Writer struct {
	RepoRoot string
	TaskID   string
	AgentID  string
	Mode     Mode
}

// New builds a writer for one task/agent execution.
func New(repoRoot, taskID, agentID string, cfg *config.Config) Writer {
	return Writer{
		RepoRoot: strings.TrimSpace(repoRoot),
		TaskID:   strings.TrimSpace(taskID),
		AgentID:  strings.TrimSpace(agentID),
		Mode:     ModeFromConfig(cfg),
	}
}

// ModeFromConfig reads work.agent_observability.mode (default best_effort).
func ModeFromConfig(cfg *config.Config) Mode {
	if cfg == nil {
		return ModeBestEffort
	}
	mode := strings.TrimSpace(cfg.Work.AgentObservability.Mode)
	if mode == "" {
		return ModeBestEffort
	}
	switch mode {
	case string(ModeWarn):
		return ModeWarn
	case string(ModeStrict):
		return ModeStrict
	default:
		return ModeBestEffort
	}
}

// Run executes fn and applies observability policy on error.
func (w Writer) Run(operation string, fn func() error) error {
	return w.handle(operation, fn())
}

// WriteLegacyLogs persists enrich-style logs under .asagiri/logs/<task-id>/.
func (w Writer) WriteLegacyLogs(ctx asagiri.AgentContext, res asagiri.AgentResult) error {
	return w.Run("agent_logs", func() error {
		return agent.WriteLogs(w.RepoRoot, w.TaskID, ctx, res)
	})
}

// WriteOrchestratedLogs persists context.json and prompt.md under the agent log directory.
func (w Writer) WriteOrchestratedLogs(ctx agentcontext.ExecutionContext, prompt string) error {
	return w.Run("agent_logs", func() error {
		return agentcontext.WriteLogs(w.RepoRoot, ctx, prompt)
	})
}

// WriteContract persists contract.json for orchestrated dev runs.
func (w Writer) WriteContract(result agentcontract.ContractValidationResult) error {
	return w.Run("contract_logs", func() error {
		return agentcontract.WriteLog(w.RepoRoot, w.TaskID, w.AgentID, result)
	})
}

// RecordLedger appends one ledger entry.
func (w Writer) RecordLedger(p agentledger.Params) error {
	return w.Run("ledger", func() error {
		return agentledger.Record(w.RepoRoot, p)
	})
}

func (w Writer) handle(operation string, err error) error {
	if err == nil {
		return nil
	}
	op := strings.TrimSpace(operation)
	if op == "" {
		op = "write"
	}
	switch w.Mode {
	case ModeWarn:
		_ = w.appendWarn(op, err)
		return nil
	case ModeStrict:
		return fmt.Errorf("agent observability %s: %w", op, err)
	default:
		return nil
	}
}

func (w Writer) appendWarn(operation string, err error) error {
	if w.RepoRoot == "" || w.TaskID == "" {
		return fmt.Errorf("agentobservability: repo_root et task_id requis pour observability.warn")
	}
	agentID := strings.TrimSpace(w.AgentID)
	var path string
	if agentID != "" {
		path = filepath.Join(agentcontext.AgentLogDir(w.RepoRoot, w.TaskID, agentID), "observability.warn")
	} else {
		path = filepath.Join(w.RepoRoot, ".asagiri", "logs", w.TaskID, "observability.warn")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	line := fmt.Sprintf("%s operation=%s error=%q\n", time.Now().UTC().Format(time.RFC3339Nano), operation, err.Error())
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	_, err = f.WriteString(line)
	return err
}
