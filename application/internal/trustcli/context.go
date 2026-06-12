package trustcli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

// WorkContext holds repo state for trust work/diff subcommands.
type WorkContext struct {
	RepoRoot string
	Config   *config.Config
	Store    *sqlite.Store
	closeFn  func()
}

// NewWorkContext builds a WorkContext with an optional close callback.
func NewWorkContext(repoRoot string, cfg *config.Config, store *sqlite.Store, closeFn func()) *WorkContext {
	return &WorkContext{
		RepoRoot: repoRoot,
		Config:   cfg,
		Store:    store,
		closeFn:  closeFn,
	}
}

// Close releases resources held by the context.
func (w *WorkContext) Close() {
	if w != nil && w.closeFn != nil {
		w.closeFn()
	}
}
