package workflow

import (
	"context"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/store/sqlite"
	"github.com/LaProgrammerie/hyper-fast-builder/application/pkg/agentflow"
)

// WorkflowRequest groups inputs for a full workflow run (spec §11.2).
type WorkflowRequest struct {
	Feature string
	TaskID  string
	Force   bool
}

// WorkflowResult summarizes a workflow execution.
type WorkflowResult struct {
	RunID  string
	Status string
}

// WorkflowEngine orchestrates feature workflows.
type WorkflowEngine interface {
	Run(ctx context.Context, req WorkflowRequest) (WorkflowResult, error)
	Resume(ctx context.Context, runID string, force bool) (WorkflowResult, error)
}

// TaskStore persists runs and tasks.
type TaskStore interface {
	SaveRun(ctx context.Context, run sqlite.Run) error
	GetRun(ctx context.Context, id string) (sqlite.Run, error)
	SaveTask(ctx context.Context, task sqlite.Task) error
	ListTasks(ctx context.Context, feature string) ([]sqlite.Task, error)
}

// WorktreeInfo describes an isolated git worktree.
type WorktreeInfo struct {
	Path       string
	BranchName string
}

// WorktreeManager creates and cleans worktrees per task.
type WorktreeManager interface {
	Create(ctx context.Context, feature, taskID string) (WorktreeInfo, error)
	Cleanup(ctx context.Context, path string) error
}

// ValidationCommand is one configured validation step.
type ValidationCommand struct {
	Name     string
	Command  string
	Required bool
}

// ValidationResult is the outcome of one validation command.
type ValidationResult struct {
	Name     string
	Command  string
	ExitCode int
	Output   string
	Err      error
}

// Validator runs validation commands in a directory.
type Validator interface {
	Run(ctx context.Context, dir string, commands []ValidationCommand) ([]ValidationResult, error)
}

// Ensure canonical task type is referenced for doc alignment.
var _ = agentflow.Task{}
