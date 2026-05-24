package agent

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
)

const OutputFormatV1 = "agentflow-v1"

// BuildContext constructs a spec §9.1 context from task fields.
func BuildContext(runID string, task *agentflow.Task, contextFiles []string) agentflow.AgentContext {
	allowed := task.Scope.AllowedPaths
	if len(allowed) == 0 {
		allowed = []string{"application/**"}
	}
	valCmds := task.Validation.Commands
	if len(valCmds) == 0 {
		valCmds = []string{"go test ./..."}
	}
	return agentflow.AgentContext{
		RunID:              runID,
		TaskID:             task.ID,
		Objective:          task.Title,
		AllowedPaths:       allowed,
		ForbiddenPaths:     task.Scope.ForbiddenPaths,
		AcceptanceCriteria: task.Acceptance,
		ValidationCommands: valCmds,
		ContextFiles:       contextFiles,
		OutputFormat:       OutputFormatV1,
	}
}

// DryRunResult returns a fixture result for dry-run mode.
func DryRunResult(summary string) agentflow.AgentResult {
	return agentflow.AgentResult{
		Status:              "completed",
		Summary:             summary,
		ChangedFiles:        []string{},
		CommandsRun:         []agentflow.CommandRun{},
		Risks:               []string{},
		RequiresHumanReview: true,
	}
}

// ParseResult attempts to parse agent stdout as AgentResult JSON.
func ParseResult(stdout string) (agentflow.AgentResult, bool) {
	stdout = trimJSON(stdout)
	if stdout == "" {
		return agentflow.AgentResult{}, false
	}
	var res agentflow.AgentResult
	if err := json.Unmarshal([]byte(stdout), &res); err != nil {
		return agentflow.AgentResult{}, false
	}
	if res.Status == "" {
		return agentflow.AgentResult{}, false
	}
	return res, true
}

func trimJSON(s string) string {
	s = trimSpace(s)
	if len(s) > 0 && s[0] == '{' {
		return s
	}
	return ""
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\n') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\n') {
		s = s[:len(s)-1]
	}
	return s
}

// WriteLogs persists context.json and result.json under .agentflow/logs/<task-id>/.
func WriteLogs(repoRoot, taskID string, ctx agentflow.AgentContext, res agentflow.AgentResult) error {
	logDir := filepath.Join(repoRoot, ".agentflow", "logs", taskID)
	if err := os.MkdirAll(logDir, 0o755); err != nil {
		return fmt.Errorf("create log dir: %w", err)
	}
	if err := writeJSON(filepath.Join(logDir, "context.json"), ctx); err != nil {
		return err
	}
	return writeJSON(filepath.Join(logDir, "result.json"), res)
}

func writeJSON(path string, v any) error {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
