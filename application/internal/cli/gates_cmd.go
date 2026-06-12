package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
	"github.com/spf13/cobra"
)

func newGatesCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gates",
		Short: "Work validation gates (human review, …)",
	}
	cmd.AddCommand(newGatesSubmitCmd(dryRun))
	return cmd
}

func newGatesSubmitCmd(dryRun *bool) *cobra.Command {
	var taskID, verdict, filePath, note string

	cmd := &cobra.Command{
		Use:   "submit human_review",
		Short: "Write a human review verdict file for a task",
		Example: `  asa gates submit human_review --task task-1 --verdict pass
  asa gates submit human_review --task task-1 --file ./review.yaml`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if args[0] != "human_review" {
				return fmt.Errorf("unsupported gate %q (MVP: human_review only)", args[0])
			}
			if taskID == "" {
				return fmt.Errorf("--task is required")
			}
			if filePath == "" && verdict == "" {
				return fmt.Errorf("provide --verdict or --file")
			}

			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()

			if _, err := ctx.Store.GetTask(taskID); err != nil {
				return fmt.Errorf("task %q not found in local state (check task id): %w", taskID, err)
			}

			var notes []string
			if note != "" {
				notes = []string{note}
			}
			verdictFile := ""
			if ctx.Config != nil {
				verdictFile = ctx.Config.Work.Gates.HumanReview.VerdictFile
			}
			path, err := workflow.WriteHumanReviewVerdictFile(ctx.RepoRoot, taskID, verdictFile, verdict, notes, filePath)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Human review verdict written: %s\n", path)
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "Task ID (required)")
	cmd.Flags().StringVar(&verdict, "verdict", "", "Verdict: pass, warn, or fail")
	cmd.Flags().StringVar(&filePath, "file", "", "Path to a full verdict YAML file")
	cmd.Flags().StringVar(&note, "note", "", "Optional note when using --verdict")
	return cmd
}
