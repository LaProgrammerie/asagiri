package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/bootstrap"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/store/sqlite"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/workflow"
)

type appContext struct {
	RepoRoot string
	Config   *config.Config
	Store    *sqlite.Store
	DryRun   bool
}

func loadContext(startDir string, dryRunFlag bool) (*appContext, error) {
	repoRoot, err := bootstrap.GitRoot(startDir)
	if err != nil {
		return nil, err
	}
	cfgPath := config.ConfigPath(repoRoot)
	cfg, err := config.Load(cfgPath, repoRoot)
	if err != nil {
		return nil, err
	}
	store, err := sqlite.Open(cfg.StateDBPath(repoRoot))
	if err != nil {
		return nil, err
	}
	if err := store.Migrate(); err != nil {
		_ = store.Close()
		return nil, fmt.Errorf("migrations SQLite: %w", err)
	}
	dryRun := dryRunFlag || strings.TrimSpace(os.Getenv("AGENTFLOW_DRY_RUN")) == "1"
	return &appContext{
		RepoRoot: repoRoot,
		Config:   cfg,
		Store:    store,
		DryRun:   dryRun,
	}, nil
}

func (c *appContext) Close() {
	if c == nil || c.Store == nil {
		return
	}
	_ = c.Store.Close()
}

func (c *appContext) Workflow() *workflow.Service {
	return workflow.NewService(c.RepoRoot, c.Config, c.Store, c.DryRun)
}
