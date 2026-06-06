package cli

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/version"
	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
	"github.com/spf13/cobra"
)

// Execute runs the asa CLI.
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
		Use:     "asa",
		Short:   "Asagiri — orchestration locale pour workflows agentiques",
		Long:    rootLong,
		Example: rootExample,
		RunE:    runRootUICommand(&dryRun),
	}
	root.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "Simuler les exécutions sans lancer d'agent ni commandes externes")

	// ── Groupes visuels (progressive disclosure) ────────────────────────────
	root.AddGroup(
		&cobra.Group{ID: "start",    Title: "Pour commencer"},
		&cobra.Group{ID: "workflow", Title: "Workflow unitaire"},
		&cobra.Group{ID: "tools",   Title: "Outils avancés"},
		&cobra.Group{ID: "system",  Title: "Système"},
	)
	root.AddCommand(inGroup("start",
		newOnboardCmd(),
		newWorkCmd(&dryRun),
		newContinueCmd(&dryRun),
		newNextCmd(&dryRun),
		newStatusCmd(&dryRun),
		newInboxCmd(&dryRun),
	)...)
	root.AddCommand(inGroup("workflow",
		newInitCmd(),
		newDoctorCmd(),
		newSpecCmd(&dryRun),
		newPlanCmd(&dryRun),
		newEnrichCmd(&dryRun),
		newDevCmd(&dryRun),
		newVerifyCmd(&dryRun),
		newReviewCmd(&dryRun),
		newPRCmd(&dryRun),
		newResumeCmd(&dryRun),
		newReportCmd(&dryRun),
		newSyncCmd(&dryRun),
		newEstimateCmd(&dryRun),
	)...)
	root.AddCommand(inGroup("tools",
		newToolsCmd(&dryRun),
		newTrustCmd(&dryRun),
		newReplayCmd(&dryRun),
		newKnowledgeCmd(&dryRun),
		newGraphCmd(&dryRun),
		newInvestigateCmd(&dryRun),
		newAnalysisCmd(),
		newFlowsCmd(&dryRun),
		newContractsCmd(&dryRun),
		newArchitectureCmd(&dryRun),
		newImpactCmd(),
		newInspectCmd(&dryRun),
		newContextCmd(&dryRun),
		newProductCmd(&dryRun),
	)...)
	versionCmd := &cobra.Command{
		Use:   "version",
		Short: "Afficher la version",
		RunE: func(cmd *cobra.Command, args []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), version.String())
			return nil
		},
	}
	root.AddCommand(inGroup("system",
		newReadyCmd(),
		newIndexCmd(&dryRun),
		newCostCmd(&dryRun),
		newMcpCmd(&dryRun),
		newPrototypeCmd(&dryRun),
		newDaemonCmd(&dryRun),
		newSessionCmd(&dryRun),
		newRuntimeCmd(&dryRun),
		newSkillsCmd(),
		newMemoryCmd(),
		newDocsCmd(),
		newCleanCmd(&dryRun),
		newMissionCmd(&dryRun),
		newDashboardCmd(&dryRun),
		newRunsCmd(&dryRun),
		newAgentsCmd(&dryRun),
		newFlowCmd(&dryRun),
		newExplainCmd(&dryRun),
		versionCmd,
	)...)

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
			agent := resolveWorkAgent(cmd, agentName, (*config.Config).WorkSpecAgent, ctx.Config)
			runID, err := ctx.Workflow().SpecFeature(cmd.Context(), args[0], agent)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "spec run créé: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&agentName, "agent", config.DefaultAgentSpec, "Agent utilisé pour la phase spec")
	cmd.AddCommand(newSpecGenerateFromProductCmd(dryRun))
	return cmd
}

func newPlanCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "plan run: %s (%d tasks)\n", runID, len(tasks))
			return nil
		},
	}
	cmd.AddCommand(newPlanGraphCmd(), newPlanExplainCmd())
	return cmd
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
			agent := resolveWorkAgent(cmd, agentName, (*config.Config).WorkEnricherAgent, ctx.Config)
			runID, err := ctx.Workflow().EnrichFeature(context.Background(), args[0], taskID, agent, force)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "enrich run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à enrichir")
	cmd.Flags().StringVar(&agentName, "agent", config.DefaultAgentEnrich, "Agent d'enrichissement")
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
			agent := resolveWorkAgent(cmd, agentName, (*config.Config).WorkDevAgent, ctx.Config)
			runID, err := ctx.Workflow().DevFeature(context.Background(), args[0], taskID, agent, force)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "dev run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à implémenter")
	cmd.Flags().StringVar(&agentName, "agent", config.DefaultAgentDev, "Agent d'implémentation")
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	return cmd
}

func newVerifyCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var force bool
	var investigateOnFailure bool
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
				if investigateOnFailure && !ctx.DryRun && !*dryRun {
					req := investigation.Request{
						Symptom:         "verify failed for " + args[0],
						Feature:         args[0],
						TaskID:          taskID,
						FromFailedTests: true,
						Depth:           investigation.DepthCI,
						NoCloud:         true,
						RepoRoot:        ctx.RepoRoot,
					}
					if rep, invErr := investigation.RunInvestigation(cmd.Context(), req, ctx.Config); invErr == nil {
						_ = investigation.FeedMemory(ctx.RepoRoot, rep)
						_, _ = fmt.Fprintf(cmd.OutOrStdout(), "investigation on failure: %s\n", rep.ID)
					}
				}
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "verify run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à vérifier")
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	cmd.Flags().BoolVar(&investigateOnFailure, "investigate-on-failure", false, "Lancer une investigation locale si la vérification échoue")
	cmd.AddCommand(newVerifyTrustCmd())
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
			agent := resolveWorkAgent(cmd, agentName, (*config.Config).WorkReviewerAgent, ctx.Config)
			runID, err := ctx.Workflow().ReviewFeature(context.Background(), args[0], taskID, agent, force)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "review run: %s\n", runID)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche à reviewer")
	cmd.Flags().StringVar(&agentName, "agent", config.DefaultAgentReviewer, "Agent de review")
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
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s\n", run.ID, run.Feature, run.Status, run.UpdatedAt.Format("2006-01-02 15:04:05"))
			}
			return nil
		},
	}
}

func newResumeCmd(dryRun *bool) *cobra.Command {
	var force bool
	var execute bool
	var maxSteps int
	cmd := &cobra.Command{
		Use:     "resume <run-id>",
		Short:   "Reprendre un run interrompu",
		Example: "  asa resume run-2026-05-17-001\n  asa resume run-2026-05-17-001 --execute",
		Args:    cobra.ExactArgs(1),
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
			if execute {
				var steps []string
				if ctx.DryRun {
					steps, err = wf.ResumeRunDryExecute(cmd.Context(), args[0], force, maxSteps)
				} else {
					steps, err = wf.ResumeRunExecute(cmd.Context(), args[0], force, maxSteps)
				}
				if err != nil {
					return err
				}
				if len(steps) == 0 {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "run terminé: aucun step à reprendre")
					return nil
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "steps exécutés (%d): %s\n", len(steps), strings.Join(steps, ", "))
				return nil
			}
			next, err := wf.ResumeRun(args[0], force)
			if err != nil {
				return err
			}
			if next == "" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "run terminé: aucun step à reprendre")
				return nil
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "prochain step: %s\n", next)
			return nil
		},
	}
	cmd.Flags().BoolVar(&force, "force", false, "Relancer une étape déjà réussie")
	cmd.Flags().BoolVar(&execute, "execute", false, "Enchaîner les steps restants (agents réels hors --dry-run global)")
	cmd.Flags().IntVar(&maxSteps, "max-steps", workflow.DefaultResumeMaxSteps, "Nombre maximum de steps à enchaîner avec --execute")
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "report markdown: %s\nreport json: %s\n", md, js)
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "worktrees supprimés: %d\n", removed)
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "checklist PR: %s\n", checklist)
			return nil
		},
	}
}


// inGroup assigns GroupID to each command and returns the slice.
func inGroup(id string, cmds ...*cobra.Command) []*cobra.Command {
	for _, c := range cmds {
		c.GroupID = id
	}
	return cmds
}

// newToolsCmd is a discovery index for advanced tools — not a wrapper.
func newToolsCmd(dryRun *bool) *cobra.Command {
	_ = dryRun
	return &cobra.Command{
		Use:   "tools",
		Short: "Répertoire des outils avancés",
		Long: `Répertoire des outils avancés disponibles dans asa.

Ces commandes restent accessibles directement (asa trust, asa replay, …).
Ce répertoire facilite leur découverte.

  asa trust        — Moteur de confiance (gates, replay sécurisé)
  asa replay       — Capturer, rejouer, comparer des workflows
  asa knowledge    — Graphe de connaissance engineering
  asa graph        — Graphes d'exécution multi-agents
  asa investigate  — Investigation structurée locale
  asa analysis     — Couche d'analyse structurelle (graphes)
  asa flows        — Extraire et inspecter les flows produit
  asa contracts    — Extraire les contrats système
  asa architecture — Projeter les implications système
  asa impact       — Analyser l'impact des changements
  asa inspect      — Inspection locale (symbol, tests, diff)
  asa context      — Afficher ou optimiser le contexte prévu
  asa product      — Review produit

Aide détaillée : asa <outil> --help`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
}
