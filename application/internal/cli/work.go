package cli

import (
	"context"
	"errors"
	"os"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/pipeline"
	"github.com/spf13/cobra"
)

func newWorkCmd(dryRun *bool) *cobra.Command {
	var (
		agent          string
		reviewer       string
		sourceName     string
		planOnly       bool
		yes            bool
		maxTasks       int
		stopAfter      string
		noReview       bool
		instruction    string
		estimateOnly   bool
		budgetEUR      float64
		preferLocal    bool
		maxInTok       int
		maxOutTok      int
		maxDurationMin int
		showCtxPlan    bool
		noCloud        bool
		allowCloud     bool
		allowOver      bool
		noCtxReduce    bool
	)

	cmd := &cobra.Command{
		Use:   `work "<instruction>"`,
		Short: "Exécuter une intention en langage naturel",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instruction = args[0]
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			actx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer actx.Close()

			snap, err := actx.snapshot()
			if err != nil {
				return err
			}
			interactive := isInteractive()
			opts := intent.WorkOptions{
				PlanOnly:    planOnly,
				Yes:         yes,
				DryRun:      actx.DryRun || *dryRun,
				MaxTasks:    maxTasks,
				StopAfter:   stopAfter,
				NoReview:    noReview,
				Agent:       agent,
				Reviewer:    reviewer,
				Interactive: interactive,
			}

			resolver := intent.NewHybridResolver()
			resolved, err := resolver.Resolve(cmd.Context(), intent.IntentInput{
				RawInstruction: instruction,
				WorkingDir:     actx.RepoRoot,
				Config:         actx.Config,
				StateSnapshot:  snap,
				Interactive:    interactive,
			})
			if err != nil {
				return err
			}
			if sourceName != "" {
				resolved.Source = sourceName
			}

			planner := &intent.DefaultPlanner{}
			plan, err := planner.BuildPlan(cmd.Context(), resolved, snap, actx.Config, opts)
			if err != nil {
				return err
			}

			v3opts := pipeline.V3Options{
				EstimateOnly:    estimateOnly,
				BudgetMajor:     budgetEUR,
				PreferLocal:     preferLocal,
				MaxInputTokens:  maxInTok,
				MaxOutputTokens: maxOutTok,
				ShowContextPlan: showCtxPlan,
				NoCloud:         noCloud,
				AllowCloud:      allowCloud,
				AllowOverBudget: allowOver,
				NoContextReduce: noCtxReduce,
				Interactive:     interactive,
				PlanOnly:        planOnly,
				DryRun:          opts.DryRun,
				Yes:             yes,
				Agent:           agent,
				Reviewer:        reviewer,
				MaxTasks:        maxTasks,
				StopAfter:       stopAfter,
				NoReview:        noReview,
			}
			if maxDurationMin > 0 {
				v3opts.MaxDuration = time.Duration(maxDurationMin) * time.Minute
			}

			if actx.Config.Intent.DefaultMode == "guided" && !opts.Yes && !opts.PlanOnly && !estimateOnly && actx.Config.Work.RequirePlanConfirmation {
				if err := requireConfirm(opts, "Proceed with execution plan?"); err != nil {
					return err
				}
			}

			app := pipeline.App{
				RepoRoot: actx.RepoRoot,
				Config:   actx.Config,
				Store:    actx.Store,
				Executor: &intent.Executor{
					Workflow: actx.Workflow(),
					Config:   actx.Config,
					SyncFn:   actx.syncPrimitive,
				},
			}

			v3res, err := pipeline.RunV3Pipeline(context.Background(), app, resolved, plan, v3opts)
			if err != nil {
				var pc *cost.BudgetPendingConfirmError
				if errors.As(err, &pc) && !yes {
					if confirmErr := requireConfirm(opts, pc.Error()); confirmErr != nil {
						return confirmErr
					}
					v3opts.UserConfirmedBudget = true
					v3res, err = pipeline.RunV3Pipeline(context.Background(), app, resolved, plan, v3opts)
				}
				if err != nil {
					return err
				}
			}

			printEstimateBoxed(cmd.OutOrStdout(), v3res.Estimate, &v3res.Optimize)

			if estimateOnly || planOnly {
				return nil
			}

			intent.PrintWorkReport(cmd.OutOrStdout(), resolved, plan, v3res.Exec)
			printWorkSummary(cmd.OutOrStdout(), instruction, v3res.Estimate, v3res.Exec)
			return nil
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "", "Agent d'implémentation")
	cmd.Flags().StringVar(&reviewer, "reviewer", "", "Agent de review")
	cmd.Flags().StringVar(&sourceName, "source", "", "Source externe (notion|local)")
	cmd.Flags().BoolVar(&planOnly, "plan-only", false, "Afficher le plan sans exécuter")
	cmd.Flags().BoolVar(&yes, "yes", false, "Mode auto sans confirmation")
	cmd.Flags().IntVar(&maxTasks, "max-tasks", 0, "Nombre max de tâches par run")
	cmd.Flags().StringVar(&stopAfter, "stop-after", "", "Arrêter après une étape (verify, dev, …)")
	cmd.Flags().BoolVar(&noReview, "no-review", false, "Désactiver la review")
	cmd.Flags().BoolVar(&estimateOnly, "estimate-only", false, "Estimation sans exécution")
	cmd.Flags().Float64Var(&budgetEUR, "budget", 0, "Budget run EUR")
	cmd.Flags().BoolVar(&preferLocal, "prefer-local", false, "Préférer étapes locales")
	cmd.Flags().IntVar(&maxInTok, "max-input-tokens", 0, "Plafond tokens entrée")
	cmd.Flags().IntVar(&maxOutTok, "max-output-tokens", 0, "Plafond tokens sortie")
	cmd.Flags().IntVar(&maxDurationMin, "max-duration", 0, "Durée max (minutes)")
	cmd.Flags().BoolVar(&showCtxPlan, "show-context-plan", false, "Afficher chemin context pack")
	cmd.Flags().BoolVar(&noCloud, "no-cloud", false, "Interdire cloud")
	cmd.Flags().BoolVar(&allowCloud, "allow-cloud", false, "Autoriser cloud explicitement")
	cmd.Flags().BoolVar(&allowOver, "allow-over-budget", false, "Dépasser le budget")
	cmd.Flags().BoolVar(&noCtxReduce, "no-context-reduction", false, "Désactiver réduction contexte")
	return cmd
}
