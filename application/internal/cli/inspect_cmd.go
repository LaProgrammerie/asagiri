package cli

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/spf13/cobra"
)

func newInspectCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "inspect",
		Short: "Inspection locale (symbol, tests, diff)",
	}
	cmd.AddCommand(
		&cobra.Command{
			Use:   "symbol <name>",
			Short: "Trouver un symbole Go",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				c, err := loadContext(mustWd(), *dryRun)
				if err != nil {
					return err
				}
				defer c.Close()
				inv, err := investigation.Run(cmd.Context(), c.RepoRoot, "", "", c.Config)
				if err != nil {
					return err
				}
				found := investigation.FindSymbolFiles(c.RepoRoot, inv.CandidateFiles, args[0]) //nolint:staticcheck
				for _, f := range found {
					fmt.Fprintln(cmd.OutOrStdout(), f)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "tests <path>",
			Short: "Tests liés à un chemin",
			Args:  cobra.ExactArgs(1),
			RunE: func(cmd *cobra.Command, args []string) error {
				c, err := loadContext(mustWd(), *dryRun)
				if err != nil {
					return err
				}
				defer c.Close()
				inv, _ := investigation.Run(cmd.Context(), c.RepoRoot, args[0], "", c.Config)
				for _, t := range inv.RelatedTests {
					fmt.Fprintln(cmd.OutOrStdout(), t)
				}
				return nil
			},
		},
		&cobra.Command{
			Use:   "diff",
			Short: "Résumé diff git",
			RunE: func(cmd *cobra.Command, args []string) error {
				c, err := loadContext(mustWd(), *dryRun)
				if err != nil {
					return err
				}
				defer c.Close()
				out, err := exec.Command("git", "-C", c.RepoRoot, "diff", "--stat").Output()
				if err != nil {
					return err
				}
				lines := strings.Split(strings.TrimSpace(string(out)), "\n")
				if len(lines) > 30 {
					lines = append(lines[:30], "...")
				}
				fmt.Fprintln(cmd.OutOrStdout(), strings.Join(lines, "\n"))
				return nil
			},
		},
	)
	return cmd
}

