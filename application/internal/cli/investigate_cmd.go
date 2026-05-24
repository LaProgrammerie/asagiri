package cli

import (
	"fmt"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/investigation"
	"github.com/spf13/cobra"
)

func newInvestigateCmd(dryRun *bool) *cobra.Command {
	var taskID string
	cmd := &cobra.Command{
		Use:   "investigate <feature>",
		Short: "Investigation locale du dépôt",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			res, err := investigation.Run(cmd.Context(), c.RepoRoot, args[0], taskID, c.Config)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Candidates: %d\nTests: %d\nSensitive: %d\nLarge: %d\n",
				len(res.CandidateFiles), len(res.RelatedTests), len(res.SensitivePaths), len(res.LargeFiles))
			for _, f := range res.CandidateFiles {
				fmt.Fprintf(cmd.OutOrStdout(), "  - %s\n", f)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche")
	return cmd
}
