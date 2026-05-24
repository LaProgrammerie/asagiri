package cli

import (
	"context"
	"errors"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/cost"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/intent"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/pipeline"
	"github.com/spf13/cobra"
)

func newEstimateCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var budget float64
	cmd := &cobra.Command{
		Use:   "estimate <feature>",
		Short: "Estimer tokens/coût/temps sans exécuter",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			v3 := pipeline.V3Options{
				EstimateOnly: true,
				BudgetMajor:  budget,
				Interactive:  isInteractive(),
			}
			res, err := runEstimateFlow(cmd.Context(), c, args[0], taskID, v3)
			if err != nil {
				var b *cost.BudgetPendingConfirmError
				if errors.As(err, &b) {
					printEstimateBoxed(cmd.OutOrStdout(), res.Estimate, &res.Optimize)
					return err
				}
				return err
			}
			printEstimateBoxed(cmd.OutOrStdout(), res.Estimate, &res.Optimize)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche")
	cmd.Flags().Float64Var(&budget, "budget", 0, "Plafond coût estimé (unité majeure, ex. EUR)")
	return cmd
}

func runEstimateFlow(ctx context.Context, c *appContext, feature, taskID string, v3 pipeline.V3Options) (pipeline.V3Result, error) {
	snap, err := c.snapshot()
	if err != nil {
		return pipeline.V3Result{}, err
	}
	resolver := intent.NewHybridResolver()
	resolved, err := resolver.Resolve(ctx, intent.IntentInput{
		RawInstruction: "estimate " + feature,
		WorkingDir:     c.RepoRoot,
		Config:         c.Config,
		StateSnapshot:  snap,
		Interactive:    isInteractive(),
	})
	if err != nil {
		return pipeline.V3Result{}, err
	}
	resolved.Feature = feature
	resolved.TaskID = taskID
	planner := &intent.DefaultPlanner{}
	plan, err := planner.BuildPlan(ctx, resolved, snap, c.Config, intent.WorkOptions{PlanOnly: true, Interactive: isInteractive()})
	if err != nil {
		return pipeline.V3Result{}, err
	}
	app := pipeline.App{
		RepoRoot: c.RepoRoot,
		Config:   c.Config,
		Store:    c.Store,
		Executor: nil,
	}
	return pipeline.RunV3Pipeline(ctx, app, resolved, plan, v3)
}

