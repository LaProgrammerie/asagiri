package agentresolve

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentadapter"
	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentobservability"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Params carries prompt resolution inputs (no subprocess).
type Params struct {
	RepoRoot     string
	Config       *config.Config
	Phase        Phase
	AgentKey     string
	RunID        string
	Feature      string
	TaskID       string
	LegacyPrompt string
	ContextFiles []string
}

// Result is the resolved prompt and orchestration metadata.
type Result struct {
	Prompt       string
	Orchestrated bool
	Warning      string
	Spec         agentspec.Spec
	AgentID      string
	Role         string
	ContextHash  string
	LogDir       string
	Phase        Phase
}

type resolveMeta struct {
	Phase        string `json:"phase,omitempty"`
	Orchestrated bool   `json:"orchestrated"`
	Warning      string `json:"warning,omitempty"`
}

// Resolve loads AgentSpec from disk, builds ExecutionContext, renders via agentadapter.
// When AgentSpec is absent, returns LegacyPrompt unchanged.
func Resolve(params Params) (Result, error) {
	legacy := params.LegacyPrompt
	phase := params.Phase
	if phase == "" {
		phase = PhaseDev
	}
	agentKey := strings.TrimSpace(params.AgentKey)
	if agentKey == "" {
		agentKey = "dev"
	}
	obs := agentobservability.New(params.RepoRoot, params.TaskID, agentKey, params.Config)

	loader := agentspec.NewLoader(params.RepoRoot)
	spec, err := loader.LoadDiskOnly(agentKey)
	if err != nil {
		out := Result{
			Prompt:       legacy,
			Orchestrated: false,
			Warning:      err.Error(),
			AgentID:      agentKey,
			LogDir:       relativeLogDir(params.RepoRoot, params.TaskID, agentKey, phase),
			Phase:        phase,
		}
		if err := obs.Run("resolve_meta", func() error {
			return writeResolveMeta(params.RepoRoot, params.TaskID, agentKey, phase, out)
		}); err != nil {
			return Result{}, err
		}
		return out, nil
	}

	in := agentcontext.Input{
		Spec:           spec,
		Feature:        params.Feature,
		TaskID:         params.TaskID,
		RunID:          params.RunID,
		Phase:          string(phase),
		UserTaskPrompt: legacy,
		ContextFiles:   params.ContextFiles,
	}
	ctx := agentcontext.Build(in)
	prompt := agentcontext.RenderPrompt(ctx)
	obs.AgentID = spec.ID
	if err := obs.Run("agent_logs", func() error {
		return agentcontext.WritePhaseLogs(params.RepoRoot, string(phase), ctx, prompt)
	}); err != nil {
		return Result{}, err
	}

	inv, invErr := agentadapter.RenderFromConfig(params.Config, agentKey, spec, ctx)
	finalPrompt := prompt
	if invErr == nil && strings.TrimSpace(inv.StdinPrompt) != "" {
		finalPrompt = inv.StdinPrompt
	}
	if err := obs.Run("invocation_logs", func() error {
		return writeInvocation(params.RepoRoot, phase, ctx, inv, invErr)
	}); err != nil {
		return Result{}, err
	}

	out := Result{
		Prompt:       finalPrompt,
		Orchestrated: true,
		Spec:         spec,
		AgentID:      spec.ID,
		Role:         strings.TrimSpace(spec.Role),
		ContextHash:  agentcontext.ContextHash(ctx),
		LogDir:       relativeLogDir(params.RepoRoot, params.TaskID, spec.ID, phase),
		Phase:        phase,
	}
	if invErr != nil {
		out.Warning = invErr.Error()
	}
	if err := obs.Run("resolve_meta", func() error {
		return writeResolveMeta(params.RepoRoot, params.TaskID, agentKey, phase, out)
	}); err != nil {
		return Result{}, err
	}
	return out, nil
}

func relativeLogDir(repoRoot, taskID, agentID string, phase Phase) string {
	abs := agentcontext.AgentLogDirForPhase(repoRoot, taskID, agentID, string(phase))
	rel, err := filepath.Rel(strings.TrimSpace(repoRoot), abs)
	if err != nil {
		return filepath.ToSlash(abs)
	}
	return filepath.ToSlash(rel)
}

func writeResolveMeta(repoRoot, taskID, agentKey string, phase Phase, res Result) error {
	if strings.TrimSpace(taskID) == "" || strings.TrimSpace(agentKey) == "" {
		return nil
	}
	dir := agentcontext.AgentLogDirForPhase(repoRoot, taskID, agentKey, string(phase))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("agentresolve: création répertoire resolve: %w", err)
	}
	body, err := json.MarshalIndent(resolveMeta{
		Phase:        string(phase),
		Orchestrated: res.Orchestrated,
		Warning:      strings.TrimSpace(res.Warning),
	}, "", "  ")
	if err != nil {
		return fmt.Errorf("agentresolve: marshal resolve.json: %w", err)
	}
	body = append(body, '\n')
	path := filepath.Join(dir, "resolve.json")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("agentresolve: écriture %s: %w", path, err)
	}
	return nil
}

func writeInvocation(repoRoot string, phase Phase, ctx agentcontext.ExecutionContext, inv agentadapter.RenderedInvocation, invErr error) error {
	if strings.TrimSpace(ctx.TaskID) == "" || strings.TrimSpace(ctx.AgentID) == "" {
		return nil
	}
	dir := agentcontext.AgentLogDirForPhase(repoRoot, ctx.TaskID, ctx.AgentID, string(phase))
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("agentresolve: création répertoire invocation: %w", err)
	}
	payload := inv
	if invErr != nil {
		if payload.Warnings == nil {
			payload.Warnings = []string{}
		}
		payload.Warnings = append(payload.Warnings, invErr.Error())
	}
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return fmt.Errorf("agentresolve: marshal invocation.json: %w", err)
	}
	body = append(body, '\n')
	path := filepath.Join(dir, "invocation.json")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("agentresolve: écriture %s: %w", path, err)
	}
	return nil
}
