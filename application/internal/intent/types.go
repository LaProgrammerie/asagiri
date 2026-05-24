package intent

import (
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

// IntentAction classifies user intent (specv2 §5).
type IntentAction string

const (
	IntentDevelop  IntentAction = "develop"
	IntentResume   IntentAction = "resume"
	IntentContinue IntentAction = "continue"
	IntentVerify   IntentAction = "verify"
	IntentReview   IntentAction = "review"
	IntentFix      IntentAction = "fix"
	IntentImport   IntentAction = "import"
	IntentSync     IntentAction = "sync"
	IntentStatus   IntentAction = "status"
	IntentUnknown  IntentAction = "unknown"
)

// IntentInput is resolver input (specv2 §5.1).
type IntentInput struct {
	RawInstruction string
	WorkingDir     string
	Config         *config.Config
	StateSnapshot  StateSnapshot
	Interactive    bool
}

// IntentConstraints carries execution limits from resolution.
type IntentConstraints struct {
	MaxTasks   int
	StopAfter  string
	NoReview   bool
	Agent      string
	Reviewer   string
	SourceHint string
}

// ResolvedIntent is resolver output (specv2 §5.1).
type ResolvedIntent struct {
	Action        IntentAction
	Feature       string
	TaskID        string
	RunID         string
	Source        string
	SourceRef     string
	Confidence    float64
	RequiresSync  bool
	RequiresPlan  bool
	RequiresHuman bool
	Constraints   IntentConstraints
	Reason        string
}

// StateSnapshot aggregates repo state for resolution and planning.
type StateSnapshot struct {
	Features      []FeatureState
	Runs          []RunState
	ActiveFeature string
}

// FeatureState describes one feature on disk / in store.
type FeatureState struct {
	Name           string
	SpecPath       string
	HasLocalSpec   bool
	HasTasks       bool
	TaskCount      int
	NextTaskID     string
	NextTaskStatus string
	SourceType     string
	SourceRef      string
	ModifiedAt     string
	Status         string // ready, draft, active
}

// RunState is a lightweight run view.
type RunState struct {
	ID        string
	Feature   string
	Status    string
	UpdatedAt string
	Resumable bool
}

// PlanStep is one primitive CLI step (specv2 §6).
type PlanStep struct {
	Command   string
	Args      []string
	Condition string
	Reason    string
}

// ExecutionPlan is the high-level plan (specv2 §6).
type ExecutionPlan struct {
	Intent ResolvedIntent
	Steps  []PlanStep
}

// WorkOptions configures work/continue execution.
type WorkOptions struct {
	PlanOnly    bool
	Yes         bool
	DryRun      bool
	MaxTasks    int
	StopAfter   string
	NoReview    bool
	Agent       string
	Reviewer    string
	Source      string
	Interactive bool
}

// NextRecommendation is output for agentflow next.
type NextRecommendation struct {
	Feature   string
	TaskID    string
	Action    string
	Reason    string
	Primitive string
}
