package agentcontext

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"sort"
)

type hashPayload struct {
	AgentID        string                 `json:"agent_id"`
	AgentRole      string                 `json:"agent_role"`
	AgentVersion   string                 `json:"agent_version"`
	AgentHash      string                 `json:"agent_hash"`
	Feature        string                 `json:"feature"`
	TaskID         string                 `json:"task_id"`
	RunID          string                 `json:"run_id"`
	Phase          string                 `json:"phase"`
	Mode           string                 `json:"mode"`
	SystemPrompt   string                 `json:"system_prompt"`
	Instructions   []string               `json:"instructions,omitempty"`
	Constraints    []string               `json:"constraints,omitempty"`
	OutputContract agentspecOutputForHash `json:"output_contract"`
	UserTaskPrompt string                 `json:"user_task_prompt"`
	ContextFiles   []string               `json:"context_files,omitempty"`
	References     []string               `json:"references,omitempty"`
	AllowedPaths   []string               `json:"allowed_paths,omitempty"`
	ForbiddenPaths []string               `json:"forbidden_paths,omitempty"`
}

type agentspecOutputForHash struct {
	Format         string   `json:"format"`
	RequiredFields []string `json:"required_fields,omitempty"`
}

// ContextHash returns a stable SHA-256 hex digest of the execution context payload.
func ContextHash(ctx ExecutionContext) string {
	payload := hashPayload{
		AgentID:      ctx.AgentID,
		AgentRole:    ctx.AgentRole,
		AgentVersion: ctx.AgentVersion,
		AgentHash:    ctx.AgentHash,
		Feature:      ctx.Feature,
		TaskID:       ctx.TaskID,
		RunID:        ctx.RunID,
		Phase:        ctx.Phase,
		Mode:         ctx.Mode,
		SystemPrompt: ctx.SystemPrompt,
		Instructions: copyStrings(ctx.Instructions),
		Constraints:  copyStrings(ctx.Constraints),
		OutputContract: agentspecOutputForHash{
			Format:         ctx.OutputContract.Format,
			RequiredFields: copyStrings(ctx.OutputContract.RequiredFields),
		},
		UserTaskPrompt: ctx.UserTaskPrompt,
		ContextFiles:   copyStrings(ctx.ContextFiles),
		References:     copyStrings(ctx.References),
		AllowedPaths:   copyStrings(ctx.AllowedPaths),
		ForbiddenPaths: copyStrings(ctx.ForbiddenPaths),
	}
	sort.Strings(payload.Instructions)
	sort.Strings(payload.Constraints)
	sort.Strings(payload.OutputContract.RequiredFields)
	sort.Strings(payload.ContextFiles)
	sort.Strings(payload.References)
	sort.Strings(payload.AllowedPaths)
	sort.Strings(payload.ForbiddenPaths)

	body, err := json.Marshal(payload)
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func copyStrings(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	return append([]string(nil), in...)
}
