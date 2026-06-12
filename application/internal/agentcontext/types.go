package agentcontext

import (
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
)

// ModeOrchestrated is the standard Asagiri agent execution mode label.
const ModeOrchestrated = "orchestrated"

// ExecutionContext is the standardized agent run context (AgentSpec + run metadata).
type ExecutionContext struct {
	AgentID        string                   `json:"agent_id"`
	AgentRole      string                   `json:"agent_role"`
	AgentVersion   string                   `json:"agent_version"`
	AgentHash      string                   `json:"agent_hash"`
	Feature        string                   `json:"feature"`
	TaskID         string                   `json:"task_id"`
	RunID          string                   `json:"run_id"`
	Phase          string                   `json:"phase"`
	Mode           string                   `json:"mode"`
	SystemPrompt   string                   `json:"system_prompt"`
	Instructions   []string                 `json:"instructions,omitempty"`
	Constraints    []string                 `json:"constraints,omitempty"`
	OutputContract agentspec.OutputContract `json:"output_contract"`
	UserTaskPrompt string                   `json:"user_task_prompt"`
	ContextFiles   []string                 `json:"context_files,omitempty"`
	References     []string                 `json:"references,omitempty"`
	AllowedPaths   []string                 `json:"allowed_paths,omitempty"`
	ForbiddenPaths []string                 `json:"forbidden_paths,omitempty"`
}
