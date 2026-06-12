package agentscli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/agentanalytics"
	"github.com/LaProgrammerie/asagiri/application/internal/agentexternal"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/LaProgrammerie/asagiri/application/internal/agentslist"
	"github.com/LaProgrammerie/asagiri/application/internal/agentsync"
	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/spf13/cobra"
)

// RootCommand returns the `asa agents` command tree (sans sous-commande watch).
func RootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "AgentSpec registry et visualisation",
	}
	cmd.AddCommand(
		newListCmd(),
		newShowCmd(),
		newRunCmd(),
		newRunsCmd(),
		newDiffCmd(),
		newExportCmd(),
		newStatsCmd(),
		newSyncCmd(),
		newExternalCmd(),
	)
	return cmd
}

func newListCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "list",
		Short:   "Lister les AgentSpec (registry ou templates embarqués)",
		Example: "  asa agents list\n  asa agents list --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			report, err := agentslist.Build(repoRoot, cfg)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentslist.FormatJSON(out, report)
			}
			return agentslist.FormatText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newShowCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "show <id>",
		Short:   "Afficher un AgentSpec par id",
		Example: "  asa agents show dev\n  asa agents show dev --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			entry, err := agentslist.Show(repoRoot, args[0], cfg)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentslist.FormatShowJSON(out, entry)
			}
			return agentslist.FormatShowText(out, entry)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func loadRepoConfig() (string, *config.Config, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	repoRoot, err := bootstrap.GitRoot(cwd)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil {
		return "", nil, fmt.Errorf("agents: config: %w", err)
	}
	return repoRoot, cfg, nil
}

func newStatsCmd() *cobra.Command {
	var jsonOut bool
	var agentID string
	var provider string
	cmd := &cobra.Command{
		Use:     "stats",
		Short:   "Statistiques agrégées des exécutions agents (ledger)",
		Example: "  asa agents stats\n  asa agents stats --json\n  asa agents stats --agent dev\n  asa agents stats --provider exec --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, _, err := loadRepoConfig()
			if err != nil {
				return err
			}
			report, err := agentanalytics.Build(repoRoot, agentanalytics.Options{
				AgentID:  agentID,
				Provider: provider,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentanalytics.FormatJSON(out, report)
			}
			return agentanalytics.FormatText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.Flags().StringVar(&agentID, "agent", "", "Filtrer par agent_id")
	cmd.Flags().StringVar(&provider, "provider", "", "Filtrer par provider")
	cmd.SilenceUsage = true
	return cmd
}

func newRunCmd() *cobra.Command {
	var jsonOut bool
	var preview bool
	var includePrompt bool
	cmd := &cobra.Command{
		Use:     "run <run_id>",
		Short:   "Inspecter un run agent (ledger + artefacts logs)",
		Example: "  asa agents run run-abc\n  asa agents run run-abc --json\n  asa agents run run-abc --preview --json\n  asa agents run run-abc --preview --include-prompt",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, _, err := loadRepoConfig()
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if preview {
				report, err := agentledger.ReplayPreview(repoRoot, args[0], agentledger.ReplayPreviewOptions{
					IncludePrompt: includePrompt,
				})
				if err != nil {
					return err
				}
				if jsonOut {
					return agentledger.FormatReplayPreviewJSON(out, report)
				}
				return agentledger.FormatReplayPreviewText(out, report)
			}
			report, err := agentledger.Inspect(repoRoot, args[0])
			if err != nil {
				return err
			}
			if jsonOut {
				return agentledger.FormatInspectJSON(out, report)
			}
			return agentledger.FormatInspectText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.Flags().BoolVar(&preview, "preview", false, "Aperçu replay read-only (artefacts + contenu JSON)")
	cmd.Flags().BoolVar(&includePrompt, "include-prompt", false, "Inclure le contenu complet de prompt.md (avec --preview)")
	cmd.SilenceUsage = true
	return cmd
}

func newExportCmd() *cobra.Command {
	var jsonOut bool
	var outputDir string
	var includePrompt bool
	cmd := &cobra.Command{
		Use:     "export <run_id>",
		Short:   "Exporter un bundle read-only d'un run agent",
		Example: "  asa agents export run-abc --output ./export/run-abc\n  asa agents export run-abc --output ./export/run-abc --include-prompt --json",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, _, err := loadRepoConfig()
			if err != nil {
				return err
			}
			report, err := agentledger.Export(repoRoot, args[0], agentledger.ExportOptions{
				OutputDir:     outputDir,
				IncludePrompt: includePrompt,
			})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentledger.FormatExportJSON(out, report)
			}
			return agentledger.FormatExportText(out, report)
		},
	}
	cmd.Flags().StringVar(&outputDir, "output", "", "Répertoire de sortie du bundle (défaut: .asagiri/exports/agents/<run_id>)")
	cmd.Flags().BoolVar(&includePrompt, "include-prompt", false, "Inclure le contenu prompt dans replay-preview.json")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newDiffCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "diff <left_run_id> <right_run_id>",
		Short:   "Comparer deux runs agent (ledger + artefacts)",
		Example: "  asa agents diff run-a run-b\n  asa agents diff run-a run-b --json",
		Args:    cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, _, err := loadRepoConfig()
			if err != nil {
				return err
			}
			report, err := agentledger.Diff(repoRoot, args[0], args[1])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentledger.FormatDiffJSON(out, report)
			}
			return agentledger.FormatDiffText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.SilenceUsage = true
	return cmd
}

func newRunsCmd() *cobra.Command {
	var jsonOut bool
	var taskID string
	cmd := &cobra.Command{
		Use:     "runs",
		Short:   "Historique des exécutions agents (ledger JSONL)",
		Example: "  asa agents runs\n  asa agents runs --json\n  asa agents runs --task task-1 --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, _, err := loadRepoConfig()
			if err != nil {
				return err
			}
			report, err := agentledger.List(repoRoot, agentledger.ListOptions{TaskID: taskID})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentledger.FormatJSON(out, report)
			}
			return agentledger.FormatText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.Flags().StringVar(&taskID, "task", "", "Filtrer par task_id")
	cmd.SilenceUsage = true
	return cmd
}

func newSyncCmd() *cobra.Command {
	var jsonOut bool
	var write bool
	var force bool
	var check bool
	var agentID string
	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "Synchroniser le registry AgentSpec depuis les templates embarqués",
		Example: "  asa agents sync\n  asa agents sync --write\n  asa agents sync --write --force\n  asa agents sync --agent dev --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, _, err := loadRepoConfig()
			if err != nil {
				return err
			}
			opts := agentsync.Options{
				AgentID: agentID,
				Write:   write,
				Force:   force,
			}
			if check && write {
				return fmt.Errorf("agents sync: --check et --write sont exclusifs")
			}
			report, err := agentsync.Sync(repoRoot, opts)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()
			if jsonOut {
				if err := agentsync.FormatJSON(out, report); err != nil {
					return err
				}
				if h := report.Hint; h != "" && !write {
					_, _ = fmt.Fprintln(errOut, "→", h)
				}
			} else {
				if err := agentsync.FormatText(out, report); err != nil {
					return err
				}
			}

			if agentsync.HasBlockingConflicts(report) {
				return errSyncConflict
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.Flags().BoolVar(&write, "write", false, "Écrire sous .asagiri/agents/ (dry-run par défaut)")
	cmd.Flags().BoolVar(&force, "force", false, "Écraser les fichiers modifiés")
	cmd.Flags().BoolVar(&check, "check", false, "Vérifier sans écrire (comportement par défaut)")
	cmd.Flags().StringVar(&agentID, "agent", "", "Synchroniser un seul agent (id template)")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}

func newExternalCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "external",
		Short:   "Audit et sync opt-in des profils provider externes",
		Example: "  asa agents external\n  asa agents external --json\n  asa agents external sync --write --agent dev",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			report, err := agentexternal.Audit(repoRoot, cfg)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return agentexternal.FormatJSON(out, report)
			}
			return agentexternal.FormatText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.AddCommand(newExternalSyncCmd())
	cmd.SilenceUsage = true
	return cmd
}

func newExternalSyncCmd() *cobra.Command {
	var jsonOut bool
	var write bool
	var force bool
	var agentID string
	cmd := &cobra.Command{
		Use:     "sync",
		Short:   "Synchroniser les profils provider vers les chemins externes explicites",
		Example: "  asa agents external sync\n  asa agents external sync --write\n  asa agents external sync --write --force --agent dev --json",
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadRepoConfig()
			if err != nil {
				return err
			}
			opts := agentexternal.SyncOptions{
				AgentID: agentID,
				Write:   write,
				Force:   force,
			}
			report, err := agentexternal.Sync(repoRoot, cfg, opts)
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			errOut := cmd.ErrOrStderr()
			if jsonOut {
				if err := agentexternal.FormatSyncJSON(out, report); err != nil {
					return err
				}
				if h := report.Hint; h != "" && !write {
					_, _ = fmt.Fprintln(errOut, "→", h)
				}
			} else {
				if err := agentexternal.FormatSyncText(out, report); err != nil {
					return err
				}
			}

			if agentexternal.HasBlockingSyncConflicts(report) {
				return errExternalSyncConflict
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON sur stdout")
	cmd.Flags().BoolVar(&write, "write", false, "Écrire les profils provider (dry-run par défaut)")
	cmd.Flags().BoolVar(&force, "force", false, "Écraser un profil externe modifié")
	cmd.Flags().StringVar(&agentID, "agent", "", "Synchroniser un seul agent (id)")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}

var errExternalSyncConflict = externalSyncConflictError{}

type externalSyncConflictError struct{}

func (externalSyncConflictError) Error() string {
	return "agents external sync: conflits non résolus (utiliser --force ou corriger manuellement)"
}

var errSyncConflict = syncConflictError{}

type syncConflictError struct{}

func (syncConflictError) Error() string {
	return "agents sync: conflits non résolus (utiliser --force ou corriger manuellement)"
}
