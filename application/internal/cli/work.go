package cli

import (
	"context"
	"fmt"
	"os"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/intent"
	"github.com/spf13/cobra"
)

func newWorkCmd(dryRun *bool) *cobra.Command {
	var (
		agent      string
		reviewer   string
		sourceName string
		planOnly   bool
		yes        bool
		maxTasks   int
		stopAfter  string
		noReview   bool
		instruction string
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
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()

			snap, err := ctx.snapshot()
			if err != nil {
				return err
			}
			interactive := isInteractive()
			opts := intent.WorkOptions{
				PlanOnly:    planOnly,
				Yes:         yes,
				DryRun:      ctx.DryRun || *dryRun,
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
				WorkingDir:     ctx.RepoRoot,
				Config:         ctx.Config,
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
			plan, err := planner.BuildPlan(cmd.Context(), resolved, snap, ctx.Config, opts)
			if err != nil {
				return err
			}

			if ctx.Config.Intent.DefaultMode == "guided" && !opts.Yes && !opts.PlanOnly && ctx.Config.Work.RequirePlanConfirmation {
				if err := requireConfirm(opts, "Proceed with execution plan?"); err != nil {
					return err
				}
			}

			exec := &intent.Executor{
				Workflow: ctx.Workflow(),
				Config:   ctx.Config,
				SyncFn:   ctx.syncPrimitive,
			}
			result, err := exec.Execute(context.Background(), plan, snap, opts)
			if err != nil {
				return err
			}
			intent.PrintWorkReport(cmd.OutOrStdout(), resolved, plan, result)
			fmt.Fprintf(cmd.OutOrStdout(), "Instruction: %s\n", instruction)
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
	return cmd
}
