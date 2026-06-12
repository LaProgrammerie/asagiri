package cli

import (
	"errors"
	"fmt"
	"io/fs"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
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
		// Guided_Remediation (R6, exigences 6.3/7.7) : un prérequis de
		// configuration absent doit produire un message actionnable nommant
		// l'élément manquant ET au moins une action de résolution, sans panic.
		// Le cas le plus courant — fichier de config inexistant — est enrichi
		// ici car config.Load se contente d'envelopper l'erreur d'E/S brute.
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf(
				"configuration Asagiri absente (%s introuvable) : lancez `asa init` pour la créer (ou copiez %s vers %s) : %w",
				config.DefaultConfigRel, config.DefaultExampleRel, config.DefaultConfigRel, err,
			)
		}
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
	dryRun := dryRunFlag || env.DryRunEnabled()
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
