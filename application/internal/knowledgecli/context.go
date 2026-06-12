package knowledgecli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Context holds repo state for knowledge CLI commands.
type Context struct {
	RepoRoot string
	Config   *config.Config
	closeFn  func()
}

// NewContext builds a Context with an optional close callback.
func NewContext(repoRoot string, cfg *config.Config, closeFn func()) *Context {
	return &Context{
		RepoRoot: repoRoot,
		Config:   cfg,
		closeFn:  closeFn,
	}
}

// Close releases resources held by the context.
func (c *Context) Close() {
	if c != nil && c.closeFn != nil {
		c.closeFn()
	}
}
