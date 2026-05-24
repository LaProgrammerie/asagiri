package cli

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/contextopt"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/spf13/cobra"
)

func newContextCmd(dryRun *bool) *cobra.Command {
	var taskID string
	var show, optimize bool
	cmd := &cobra.Command{
		Use:   "context <feature>",
		Short: "Afficher ou optimiser le contexte prévu",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !show && !optimize {
				return fmt.Errorf("spécifier --show ou --optimize")
			}
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			inv, err := investigation.Run(cmd.Context(), c.RepoRoot, args[0], taskID, c.Config)
			if err != nil {
				return err
			}
			entries, err := contextopt.Collect(c.RepoRoot, args[0], c.Config, contextopt.CollectOpts{})
			if err != nil {
				return err
			}
			contextopt.ScoreByKeywords(entries, taskID+" "+args[0], args[0])
			reduced, _ := contextopt.Reduce(entries, c.Config, contextopt.ReduceOpts{})
			pack := contextopt.BuildPack(c.Config, contextopt.PackInput{
				Feature:      args[0],
				TaskID:       taskID,
				Inv:          inv,
				ReducedFiles: reduced,
				OutputFormat: "markdown",
			})
			rawTok := contextopt.PackApproxTokens(pack, c.Config.TokenEst)
			if show {
				fmt.Fprint(cmd.OutOrStdout(), contextopt.RenderPackMarkdown(pack))
			}
			if optimize {
				fmt.Fprintf(cmd.OutOrStdout(), "Original context: ~%d tokens\nOptimized: ~%d tokens\n", rawTok, rawTok)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche")
	cmd.Flags().BoolVar(&show, "show", false, "Afficher le contexte pack")
	cmd.Flags().BoolVar(&optimize, "optimize", false, "Afficher le résumé tokens")
	return cmd
}
