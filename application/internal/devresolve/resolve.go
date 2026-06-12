package devresolve

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentresolve"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Params carries dev prompt resolution inputs (no subprocess).
type Params struct {
	RepoRoot     string
	Config       *config.Config
	AgentKey     string
	RunID        string
	Feature      string
	TaskID       string
	ContextFiles []string
}

// Result is the resolved dev prompt and orchestration metadata.
type Result = agentresolve.Result

// LegacyDevPrompt is the historical devOneTask prompt (unchanged fallback).
func LegacyDevPrompt(taskID string) string {
	return "Implémente la task " + strings.TrimSpace(taskID)
}

// Resolve loads AgentSpec from disk, builds ExecutionContext, renders via agentadapter.
// On missing/invalid spec, returns LegacyDevPrompt without failing the workflow.
func Resolve(params Params) (Result, error) {
	return agentresolve.Resolve(agentresolve.Params{
		RepoRoot:     params.RepoRoot,
		Config:       params.Config,
		Phase:        agentresolve.PhaseDev,
		AgentKey:     params.AgentKey,
		RunID:        params.RunID,
		Feature:      params.Feature,
		TaskID:       params.TaskID,
		LegacyPrompt: LegacyDevPrompt(params.TaskID),
		ContextFiles: params.ContextFiles,
	})
}
