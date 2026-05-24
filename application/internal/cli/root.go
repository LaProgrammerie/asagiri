package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/version"
	"github.com/spf13/cobra"
)

// Execute runs the agentflow CLI.
func Execute() error {
	return RootCommand().Execute()
}

// RootCommand returns the cobra command tree wired for Execute().
func RootCommand() *cobra.Command {
	return newRootCmd()
}

func newRootCmd() *cobra.Command {
	var dryRun bool

	root := &cobra.Command{
		Use:     "agentflow",
		Short:   "Orchestrateur CLI local pour workflows de développement agentique",
		Long:    rootLong,
		Example: rootExample,
	}
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simuler les exécutions sans lancer d'agent ni commandes externes")

	root.AddCommand(
		newInitCmd(),
		newDoctorCmd(),
		newSpecCmd(&dryRun),
		newPlanCmd(&dryRun),
		newEnrichCmd(&dryRun),
		newDevCmd(&dryRun),
		newVerifyCmd(&dryRun),
		newReviewCmd(&dryRun),
		newStatusCmd(&dryRun),
		newResumeCmd(&dryRun),
		newReportCmd(&dryRun),
		newCleanCmd(&dryRun),
		newPRCmd(&dryRun),
		newIndexCmd(&dryRun),
		newWorkCmd(&dryRun),
		newContinueCmd(&dryRun),
		newNextCmd(&dryRun),
		newInboxCmd(&dryRun),
		newSyncCmd(&dryRun),
		newEstimateCmd(&dryRun),
		newInvestigateCmd(&dryRun),
		newContextCmd(&dryRun),
		newCostCmd(&dryRun),
		newInspectCmd(&dryRun),
		newMcpCmd(&dryRun),
		newDocsCmd(),
		&cobra.Command{
			Use:   "version",
			Short: "Afficher la version",
			RunE: func(cmd *cobra.Command, args []string) error {
				fmt.Fprintln(cmd.OutOrStdout(), version.Version)
				return nil
			},
		},
	)

	return root
}

func newSpecCmd(dryRun *bool) *cobra.Command {
	var agentName string
	cmd := &cobra.Command{
		Use:   "spec <feature>",
		Short: "Lire ou produire une spec via un agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID, err := ctx.Workflow().SpecFeature(cmd.Context(), args[0], agentName)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "spec run créé: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&agentName, "agent", "kiro", "Agent utilisé pour la phase spec")
	return cmd
}

func newPlanCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "plan <feature>",
		Short: "Générer ou normaliser le plan de tâches",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID, tasks, err := ctx.Workflow().PlanFeature(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "plan run: %s (%d tasks)\n", runID, len(tasks))
			return nil
		},
	}
}

func newEnrichCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var agentName string
	var force bool
	cmd := &cobra.Command{
		Use:   "enrich <feature>",
		Short: "Enrichir une tâche (ex. via Ollama)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID, err := ctx.Workflow().EnrichFeature(context.Background(), args[0], taskID, agentName, force)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "enrich run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à enrichir")
	cmd.Flags().StringVar(&agentName, "agent", "ollama", "Agent d'enrichissement")
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	return cmd
}

func newDevCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var agentName string
	var force bool
	cmd := &cobra.Command{
		Use:   "dev <feature>",
		Short: "Lancer l'implémentation d'une feature ou tâche",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID, err := ctx.Workflow().DevFeature(context.Background(), args[0], taskID, agentName, force)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "dev run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à implémenter")
	cmd.Flags().StringVar(&agentName, "agent", "cursor", "Agent d'implémentation")
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	return cmd
}

func newVerifyCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var force bool
	cmd := &cobra.Command{
		Use:   "verify <feature>",
		Short: "Exécuter les validations locales",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID, err := ctx.Workflow().VerifyFeature(context.Background(), args[0], taskID, force)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "verify run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à vérifier")
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	return cmd
}

func newReviewCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var agentName string
	var force bool
	cmd := &cobra.Command{
		Use:   "review <feature>",
		Short: "Lancer une review indépendante",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID, err := ctx.Workflow().ReviewFeature(context.Background(), args[0], taskID, agentName, force)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "review run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à reviewer")
	cmd.Flags().StringVar(&agentName, "agent", "codex", "Agent de review")
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	return cmd
}

func newStatusCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Afficher l'état des runs",
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runs, err := ctx.Workflow().Status(20)
			if err != nil {
				return err
			}
			for _, run := range runs {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", run.ID, run.Feature, run.Status, run.UpdatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func newResumeCmd(dryRun *bool) *cobra.Command {
	var force bool
	var execute bool
	cmd := &cobra.Command{
		Use:   "resume <run-id>",
		Short: "Reprendre un run interrompu",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			wf := ctx.Workflow()
			if execute && ctx.DryRun {
				next, err := wf.ResumeRunDryExecute(cmd.Context(), args[0], force)
				if err != nil {
					return err
				}
				fmt.Fprintf(cmd.OutOrStdout(), "reprise dry-run exécutée: %s\n", next)
				return nil
			}
			next, err := wf.ResumeRun(args[0], force)
			if err != nil {
				return err
			}
			if next == "" {
				fmt.Fprintln(cmd.OutOrStdout(), "run terminé: aucun step à reprendre")
				return nil
			}
			fmt.Fprintf(cmd.OutOrStdout(), "prochain step: %s\n", next)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	cmd.Flags().BoolVar(&execute, "execute", false, "Exécuter le prochain step (dry-run uniquement)")
	return cmd
}

func newReportCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "report <run-id>",
		Short: "Afficher le rapport d'un run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			md, js, err := ctx.Workflow().GenerateReport(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "report markdown: %s\nreport json: %s\n", md, js)
			return nil
		},
	}
}

func newCleanCmd(dryRun *bool) *cobra.Command {
	var onlyMerged bool
	var onlyFailed bool
	cmd := &cobra.Command{
		Use:   "clean",
		Short: "Nettoyer worktrees et artefacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			removed, err := ctx.Workflow().Clean(context.Background(), onlyMerged, onlyFailed)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "worktrees supprimés: %d\n", removed)
			return nil
		},
	}
	cmd.Flags().BoolVar(&onlyMerged, "merged", false, "Nettoyer uniquement les tâches mergées")
	cmd.Flags().BoolVar(&onlyFailed, "failed", false, "Nettoyer uniquement les tâches en échec")
	return cmd
}

func newPRCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "pr <feature>",
		Short: "Exporter diff et checklist de PR",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			checklist, err := ctx.Workflow().PreparePR(context.Background(), args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "checklist PR: %s\n", checklist)
			return nil
		},
	}
}
