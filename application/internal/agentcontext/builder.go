package agentcontext

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
)

// Input carries AgentSpec and run metadata for building an ExecutionContext.
// No agent subprocess is started.
type Input struct {
	Spec               agentspec.Spec
	Feature            string
	TaskID             string
	RunID              string
	Phase              string
	UserTaskPrompt     string
	ContextFiles       []string
	References         []string
	AllowedPaths       []string
	ForbiddenPaths     []string
	ValidationCommands []string
}

// Build constructs an ExecutionContext from AgentSpec and run metadata.
func Build(in Input) ExecutionContext {
	spec := in.Spec
	ctx := ExecutionContext{
		AgentID:        strings.TrimSpace(spec.ID),
		AgentRole:      strings.TrimSpace(spec.Role),
		AgentVersion:   strings.TrimSpace(spec.Version),
		AgentHash:      strings.TrimSpace(spec.ContentHash),
		Feature:        strings.TrimSpace(in.Feature),
		TaskID:         strings.TrimSpace(in.TaskID),
		RunID:          strings.TrimSpace(in.RunID),
		Phase:          strings.TrimSpace(in.Phase),
		Mode:           ModeOrchestrated,
		SystemPrompt:   strings.TrimSpace(spec.SystemPrompt),
		Instructions:   append([]string(nil), spec.Instructions...),
		Constraints:    append([]string(nil), spec.Constraints...),
		OutputContract: spec.OutputContract,
		UserTaskPrompt: strings.TrimSpace(in.UserTaskPrompt),
		ContextFiles:   append([]string(nil), in.ContextFiles...),
		References:     append([]string(nil), in.References...),
		AllowedPaths:   append([]string(nil), in.AllowedPaths...),
		ForbiddenPaths: append([]string(nil), in.ForbiddenPaths...),
	}
	if ctx.AgentHash == "" {
		ctx.AgentHash = agentspec.SemanticHash(spec)
	}
	return ctx
}

// AgentLogDir returns `.asagiri/logs/<task-id>/agents/<agent-id>/`.
func AgentLogDir(repoRoot, taskID, agentID string) string {
	return filepath.Join(
		strings.TrimSpace(repoRoot),
		".asagiri", "logs",
		strings.TrimSpace(taskID),
		"agents",
		strings.TrimSpace(agentID),
	)
}

// WriteLogs persists context.json and prompt.md under the agent log directory.
func WriteLogs(repoRoot string, ctx ExecutionContext, prompt string) error {
	return WritePhaseLogs(repoRoot, "dev", ctx, prompt)
}

// WritePhaseLogs persists orchestration logs under the phase-specific agent directory.
func WritePhaseLogs(repoRoot, phase string, ctx ExecutionContext, prompt string) error {
	if strings.TrimSpace(ctx.TaskID) == "" {
		return fmt.Errorf("agentcontext: task_id requis pour les logs")
	}
	if strings.TrimSpace(ctx.AgentID) == "" {
		return fmt.Errorf("agentcontext: agent_id requis pour les logs")
	}
	dir := AgentLogDirForPhase(repoRoot, ctx.TaskID, ctx.AgentID, phase)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("agentcontext: création répertoire logs: %w", err)
	}
	if err := writeJSON(filepath.Join(dir, "context.json"), ctx); err != nil {
		return err
	}
	promptPath := filepath.Join(dir, "prompt.md")
	if err := os.WriteFile(promptPath, []byte(prompt), 0o644); err != nil {
		return fmt.Errorf("agentcontext: écriture %s: %w", promptPath, err)
	}
	return nil
}

// WriteFromInput builds, renders, and persists logs in one call.
func WriteFromInput(repoRoot string, in Input) (ExecutionContext, string, error) {
	ctx := Build(in)
	prompt := RenderPrompt(ctx)
	if err := WriteLogs(repoRoot, ctx, prompt); err != nil {
		return ctx, prompt, err
	}
	return ctx, prompt, nil
}

func writeJSON(path string, v any) error {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("agentcontext: marshal JSON: %w", err)
	}
	body = append(body, '\n')
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("agentcontext: écriture %s: %w", path, err)
	}
	return nil
}
