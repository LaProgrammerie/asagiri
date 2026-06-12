package workcli

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
)

// ContextDeps wires repo state into WorkContext from the root `asa` package.
type ContextDeps struct {
	RepoRoot string
	Config   *config.Config
	Store    *sqlite.Store
	DryRun   bool
	Snapshot func() (intent.StateSnapshot, error)
	Workflow func() *workflow.Service
	SyncFn   func(context.Context, []string, bool) error
	Close    func()
}

// WorkContext holds repo state for the work command.
type WorkContext struct {
	RepoRoot string
	Config   *config.Config
	Store    *sqlite.Store
	DryRun   bool
	snapshot func() (intent.StateSnapshot, error)
	workflow func() *workflow.Service
	syncFn   func(context.Context, []string, bool) error
	closeFn  func()
}

// NewWorkContext builds a WorkContext from injected dependencies.
func NewWorkContext(deps ContextDeps) *WorkContext {
	return &WorkContext{
		RepoRoot: deps.RepoRoot,
		Config:   deps.Config,
		Store:    deps.Store,
		DryRun:   deps.DryRun,
		snapshot: deps.Snapshot,
		workflow: deps.Workflow,
		syncFn:   deps.SyncFn,
		closeFn:  deps.Close,
	}
}

// Close releases resources held by the context.
func (w *WorkContext) Close() {
	if w != nil && w.closeFn != nil {
		w.closeFn()
	}
}

// Snapshot returns the current intent state snapshot.
func (w *WorkContext) Snapshot() (intent.StateSnapshot, error) {
	if w.snapshot == nil {
		return intent.StateSnapshot{}, nil
	}
	return w.snapshot()
}

// Workflow returns the workflow service for this repo.
func (w *WorkContext) Workflow() *workflow.Service {
	if w.workflow == nil {
		return nil
	}
	return w.workflow()
}

// SyncPrimitive runs a source sync primitive.
func (w *WorkContext) SyncPrimitive(ctx context.Context, args []string, force bool) error {
	if w.syncFn == nil {
		return nil
	}
	return w.syncFn(ctx, args, force)
}

// Options wires the work CLI command from `internal/cli`.
type Options struct {
	DryRun          *bool
	LoadWorkContext func(startDir string, dryRun bool) (*WorkContext, error)
}
