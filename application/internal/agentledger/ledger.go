package agentledger

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

const (
	ReportVersion = "agents-runs-v1"
	ledgerRel     = ".asagiri/logs/agents/ledger.jsonl"
)

// Entry is one append-only agent execution record.
type Entry struct {
	TaskID        string `json:"task_id"`
	RunID         string `json:"run_id"`
	Feature       string `json:"feature"`
	AgentID       string `json:"agent_id"`
	Role          string `json:"role,omitempty"`
	Provider      string `json:"provider,omitempty"`
	Phase         string `json:"phase,omitempty"`
	StartedAt     string `json:"started_at"`
	EndedAt       string `json:"ended_at"`
	DurationMS    int64  `json:"duration_ms"`
	ExitCode      int    `json:"exit_code"`
	PromptHash    string `json:"prompt_hash"`
	ContextHash   string `json:"context_hash,omitempty"`
	OutputHash    string `json:"output_hash"`
	ContractValid *bool  `json:"contract_valid,omitempty"`
	LogDir        string `json:"log_dir"`
	DryRun        bool   `json:"dry_run,omitempty"`
}

// Params carries inputs to record one agent run.
type Params struct {
	TaskID        string
	RunID         string
	Feature       string
	AgentKey      string
	AgentID       string
	Role          string
	Provider      string
	Phase         string
	Prompt        string
	ContextHash   string
	ContractValid *bool
	LogDir        string
	Result        agent.RunResult
}

// Report is the read model for `asa agents runs`.
type Report struct {
	ReportVersion string  `json:"report_version"`
	LedgerPath    string  `json:"ledger_path"`
	Count         int     `json:"count"`
	Entries       []Entry `json:"entries"`
}

// ListOptions filters ledger reads.
type ListOptions struct {
	TaskID string
}

// LedgerPath returns the repo-relative ledger path.
func LedgerPath() string {
	return ledgerRel
}

// Path joins repo root with the ledger file path.
func Path(repoRoot string) string {
	return filepath.Join(strings.TrimSpace(repoRoot), filepath.FromSlash(ledgerRel))
}

// HashText returns a stable SHA-256 hex digest of text.
func HashText(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// ProviderFromConfig resolves the configured provider id for an agent key.
func ProviderFromConfig(cfg *config.Config, agentKey string) string {
	if cfg == nil {
		return ""
	}
	key := strings.TrimSpace(agentKey)
	if key == "" {
		return ""
	}
	if a, ok := cfg.Agents[key]; ok {
		return strings.TrimSpace(a.Provider)
	}
	return ""
}

// Record appends one ledger entry (best-effort; returns error only on I/O failure).
func Record(repoRoot string, p Params) error {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return fmt.Errorf("agentledger: repo_root requis")
	}
	entry := EntryFromParams(p)
	return Append(repoRoot, entry)
}

// EntryFromParams builds a ledger entry from run metadata.
func EntryFromParams(p Params) Entry {
	agentID := strings.TrimSpace(p.AgentID)
	if agentID == "" {
		agentID = strings.TrimSpace(p.AgentKey)
	}
	return Entry{
		TaskID:        strings.TrimSpace(p.TaskID),
		RunID:         strings.TrimSpace(p.RunID),
		Feature:       strings.TrimSpace(p.Feature),
		AgentID:       agentID,
		Role:          strings.TrimSpace(p.Role),
		Provider:      strings.TrimSpace(p.Provider),
		Phase:         strings.TrimSpace(p.Phase),
		StartedAt:     strings.TrimSpace(p.Result.StartedAt),
		EndedAt:       strings.TrimSpace(p.Result.EndedAt),
		DurationMS:    durationMS(p.Result.StartedAt, p.Result.EndedAt),
		ExitCode:      p.Result.ExitCode,
		PromptHash:    HashText(p.Prompt),
		ContextHash:   strings.TrimSpace(p.ContextHash),
		OutputHash:    HashText(p.Result.Stdout),
		ContractValid: p.ContractValid,
		LogDir:        filepath.ToSlash(strings.TrimSpace(p.LogDir)),
		DryRun:        p.Result.DryRun,
	}
}

// Append writes one JSON line to the ledger.
func Append(repoRoot string, entry Entry) error {
	path := Path(repoRoot)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("agentledger: mkdir: %w", err)
	}
	body, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("agentledger: marshal: %w", err)
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("agentledger: open: %w", err)
	}
	defer func() { _ = f.Close() }()
	if _, err := f.Write(append(body, '\n')); err != nil {
		return fmt.Errorf("agentledger: write: %w", err)
	}
	return nil
}

// List reads the ledger and returns matching entries (newest first).
func List(repoRoot string, opts ListOptions) (Report, error) {
	repoRoot = strings.TrimSpace(repoRoot)
	report := Report{
		ReportVersion: ReportVersion,
		LedgerPath:    ledgerRel,
		Entries:       []Entry{},
	}
	path := Path(repoRoot)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return report, nil
		}
		return Report{}, fmt.Errorf("agentledger: read: %w", err)
	}
	taskFilter := strings.TrimSpace(opts.TaskID)
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry Entry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			continue
		}
		if taskFilter != "" && entry.TaskID != taskFilter {
			continue
		}
		report.Entries = append(report.Entries, entry)
	}
	if err := scanner.Err(); err != nil {
		return Report{}, fmt.Errorf("agentledger: scan: %w", err)
	}
	sort.Slice(report.Entries, func(i, j int) bool {
		return report.Entries[i].StartedAt > report.Entries[j].StartedAt
	})
	report.Count = len(report.Entries)
	return report, nil
}

func durationMS(started, ended string) int64 {
	if strings.TrimSpace(started) == "" || strings.TrimSpace(ended) == "" {
		return 0
	}
	st, err1 := time.Parse(time.RFC3339Nano, started)
	en, err2 := time.Parse(time.RFC3339Nano, ended)
	if err1 != nil || err2 != nil {
		return 0
	}
	return en.Sub(st).Milliseconds()
}
