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
	ID         string          `yaml:"id" json:"id"`
	Title      string          `yaml:"title" json:"title"`
	Feature    string          `yaml:"feature" json:"feature"`
	Status     string          `yaml:"status" json:"status"`
	Risk       string          `yaml:"risk,omitempty" json:"risk,omitempty"`
	Type       string          `yaml:"type,omitempty" json:"type,omitempty"`
	Source     TaskSource      `yaml:"source,omitempty" json:"source,omitempty"`
	Scope      TaskScope       `yaml:"scope,omitempty" json:"scope,omitempty"`
	Acceptance []string        `yaml:"acceptance,omitempty" json:"acceptance,omitempty"`
	Validation TaskValidation  `yaml:"validation,omitempty" json:"validation,omitempty"`
	Agents     TaskAgents      `yaml:"agents,omitempty" json:"agents,omitempty"`
	Gates      *TaskGates      `yaml:"gates,omitempty" json:"gates,omitempty"`
	Governance *TaskGovernance `yaml:"governance,omitempty" json:"governance,omitempty"`
	Metadata   TaskMetadata    `yaml:"metadata,omitempty" json:"metadata,omitempty"`
}

// TaskGates stores work-gate evaluation history on a task payload (ADR-032 Phase 2).
type TaskGates struct {
	History []GateHistoryEntry `yaml:"history,omitempty" json:"history,omitempty"`
}

// GateHistoryEntry is one work-gate evaluation persisted on the task (GateEvaluation alias in docs).
type GateHistoryEntry struct {
	Gate       string        `yaml:"gate,omitempty" json:"gate,omitempty"`
	At         string        `yaml:"at" json:"at"`
	Status     string        `yaml:"status" json:"status"`
	Confidence float64       `yaml:"confidence" json:"confidence"`
	Notes      []string      `yaml:"notes,omitempty" json:"notes,omitempty"`
	Findings   []GateFinding `yaml:"findings,omitempty" json:"findings,omitempty"`
	Retry      int           `yaml:"retry,omitempty" json:"retry,omitempty"`
	DryRun     bool          `yaml:"dry_run,omitempty" json:"dry_run,omitempty"`
	ParseError string        `yaml:"parse_error,omitempty" json:"parse_error,omitempty"`
}

// GateFinding is a structured finding from a work gate validator.
type GateFinding struct {
	Code     string   `yaml:"code" json:"code"`
	Severity string   `yaml:"severity" json:"severity"`
	Message  string   `yaml:"message" json:"message"`
	Actions  []string `yaml:"actions,omitempty" json:"actions,omitempty"`
}

// GovernanceFinding is kept for backward-compatible governance.history payloads.
type GovernanceFinding = GateFinding

// TaskGovernance stores governance gate history on a task payload.
type TaskGovernance struct {
	History []GovernanceRecord `yaml:"history,omitempty" json:"history,omitempty"`
	Retries int                `yaml:"retries,omitempty" json:"retries,omitempty"`
}

// GovernanceRecord is one governance gate evaluation persisted on the task.
type GovernanceRecord struct {
	At         string              `yaml:"at" json:"at"`
	Status     string              `yaml:"status" json:"status"`
	Confidence float64             `yaml:"confidence" json:"confidence"`
	Notes      []string            `yaml:"notes,omitempty" json:"notes,omitempty"`
	Findings   []GovernanceFinding `yaml:"findings,omitempty" json:"findings,omitempty"`
	Retry      int                 `yaml:"retry,omitempty" json:"retry,omitempty"`
	DryRun     bool                `yaml:"dry_run,omitempty" json:"dry_run,omitempty"`
	ParseError string              `yaml:"parse_error,omitempty" json:"parse_error,omitempty"`
}

type TaskSource struct {
	Spec              string `yaml:"spec,omitempty" json:"spec,omitempty"`
	Section           string `yaml:"section,omitempty" json:"section,omitempty"`
	Product           string `yaml:"product,omitempty" json:"product,omitempty"`
	Flow              string `yaml:"flow,omitempty" json:"flow,omitempty"`
	Step              string `yaml:"step,omitempty" json:"step,omitempty"`
	Action            string `yaml:"action,omitempty" json:"action,omitempty"`
	BusinessObjective string `yaml:"business_objective,omitempty" json:"business_objective,omitempty"`
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
	Reviewer    string `yaml:"reviewer,omitempty" json:"reviewer,omitempty"`
	Enricher    string `yaml:"enricher,omitempty" json:"enricher,omitempty"`
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
