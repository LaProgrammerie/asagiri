package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/bootstrap"
	"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialiser AgentFlow dans le dépôt courant",
		Long:  "Crée .agentflow/, copie config.yaml si absent, et initialise state.sqlite.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if err := bootstrap.Init(cwd); err != nil {
				return err
			}
			fmt.Fprintln(cmd.OutOrStdout(), "AgentFlow initialisé.")
			fmt.Fprintln(cmd.OutOrStdout(), "Vérifiez que .agentflow/state.sqlite et worktrees/ sont dans .gitignore.")
			return nil
		},
	}
}
