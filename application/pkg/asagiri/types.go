package asagiri

import "time"

// Task status values (spec §8.2).
const (
	StatusPending      = "pending"
	StatusPlanned      = "planned"
	StatusEnriched     = "enriched"
	StatusRunning      = "running"
	StatusImplemented  = "implemented"
	StatusVerifyFailed = "verify_failed"
	StatusVerified     = "verified"
	StatusReviewFailed = "review_failed"
	StatusReviewed     = "reviewed"
	StatusReadyForPR   = "ready_for_pr"
	StatusMerged       = "merged"
	StatusAborted      = "aborted"
	StatusFailed       = "failed"
)

// Task is the canonical task model (spec §8.1).
type Task struct {
	ID         string            `yaml:"id" json:"id"`
	Title      string            `yaml:"title" json:"title"`
	Feature    string            `yaml:"feature" json:"feature"`
	Status     string            `yaml:"status" json:"status"`
	Risk       string            `yaml:"risk,omitempty" json:"risk,omitempty"`
	Type       string            `yaml:"type,omitempty" json:"type,omitempty"`
	Source     TaskSource        `yaml:"source,omitempty" json:"source,omitempty"`
	Scope      TaskScope         `yaml:"scope,omitempty" json:"scope,omitempty"`
	Acceptance []string          `yaml:"acceptance,omitempty" json:"acceptance,omitempty"`
	Validation TaskValidation    `yaml:"validation,omitempty" json:"validation,omitempty"`
	Agents     TaskAgents        `yaml:"agents,omitempty" json:"agents,omitempty"`
	Metadata   TaskMetadata      `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

type TaskSource struct {
	Spec    string `yaml:"spec,omitempty" json:"spec,omitempty"`
	Section string `yaml:"section,omitempty" json:"section,omitempty"`
}

type TaskScope struct {
	AllowedPaths   []string `yaml:"allowed_paths,omitempty" json:"allowed_paths,omitempty"`
	ForbiddenPaths []string `yaml:"forbidden_paths,omitempty" json:"forbidden_paths,omitempty"`
}

type TaskValidation struct {
	Commands []string `yaml:"commands,omitempty" json:"commands,omitempty"`
}

type TaskAgents struct {
	Implementer string `yaml:"implementer,omitempty" json:"implementer,omitempty"`
	Reviewer      string `yaml:"reviewer,omitempty" json:"reviewer,omitempty"`
	Enricher      string `yaml:"enricher,omitempty" json:"enricher,omitempty"`
}

type TaskMetadata struct {
	CreatedAt string `yaml:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt string `yaml:"updated_at,omitempty" json:"updated_at,omitempty"`
}

// TouchMetadata sets RFC3339 timestamps if empty.
func (t *Task) TouchMetadata(now time.Time) {
	ts := now.UTC().Format(time.RFC3339)
	if t.Metadata.CreatedAt == "" {
		t.Metadata.CreatedAt = ts
	}
	t.Metadata.UpdatedAt = ts
}

// AgentContext is the standardized agent input (spec §9.1).
type AgentContext struct {
	RunID              string   `json:"run_id"`
	TaskID             string   `json:"task_id"`
	Objective          string   `json:"objective"`
	AllowedPaths       []string `json:"allowed_paths"`
	ForbiddenPaths     []string `json:"forbidden_paths"`
	AcceptanceCriteria []string `json:"acceptance_criteria"`
	ValidationCommands []string `json:"validation_commands"`
	ContextFiles       []string `json:"context_files"`
	OutputFormat       string   `json:"output_format"`
}

// CommandRun records one validation command outcome in agent output.
type CommandRun struct {
	Command  string `json:"command"`
	ExitCode int    `json:"exit_code"`
}

// AgentResult is the standardized agent output (spec §9.2).
type AgentResult struct {
	Status              string       `json:"status"`
	Summary             string       `json:"summary"`
	ChangedFiles        []string     `json:"changed_files,omitempty"`
	CommandsRun         []CommandRun `json:"commands_run,omitempty"`
	Risks               []string     `json:"risks,omitempty"`
	RequiresHumanReview bool         `json:"requires_human_review"`
}
