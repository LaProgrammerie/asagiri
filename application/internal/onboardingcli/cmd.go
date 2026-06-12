package onboardingcli

import (
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/spf13/cobra"
)

// Options wires onboarding CLI commands from the root `asa` package.
type Options struct {
	RunOnboardUI func(cmd *cobra.Command, args []string) error
}

// OnboardCommand returns the `asa onboard` command.
func OnboardCommand(opts Options) *cobra.Command {
	cmdOpts := onboarding.Options{}
	cmd := &cobra.Command{
		Use:     "onboard",
		Short:   "Préparer le dépôt (config, docs, validation)",
		Example: "  asa onboard --yes --non-interactive\n  asa onboard --dry-run\n  asa onboard --ui",
		RunE: func(cmd *cobra.Command, args []string) error {
			if cmdOpts.UI {
				return opts.RunOnboardUI(cmd, args)
			}
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			if cmdOpts.NonInteractive && !cmdOpts.Yes {
				return cmd.Help()
			}
			_, err = onboarding.Onboard(cwd, cmdOpts, cmd.InOrStdin(), cmd.OutOrStdout())
			return err
		},
	}
	cmd.Flags().BoolVar(&cmdOpts.Yes, "yes", false, "Accepter les valeurs par défaut")
	cmd.Flags().BoolVar(&cmdOpts.NonInteractive, "non-interactive", false, "Sans prompts (requiert --yes)")
	cmd.Flags().StringVar(&cmdOpts.Stack, "stack", "auto", "Stack: auto|go|php|node")
	cmd.Flags().BoolVar(&cmdOpts.CheckOnly, "check-only", false, "Readiness seulement, sans écriture")
	cmd.Flags().BoolVar(&cmdOpts.Resume, "resume", false, "Reprendre depuis .asagiri/onboarding/state.json")
	cmd.Flags().StringVar(&cmdOpts.Step, "step", "", "Aller à une étape (project|stack|agents|docs|feature|validate)")
	cmd.Flags().BoolVar(&cmdOpts.Plain, "plain", false, "Sortie texte plain")
	cmd.Flags().BoolVar(&cmdOpts.JSON, "json", false, "Sortie JSON")
	cmd.Flags().BoolVar(&cmdOpts.CI, "ci", false, "Mode CI (plain)")
	cmd.Flags().BoolVar(&cmdOpts.Strict, "strict", false, "Traiter les warns comme des échecs")
	cmd.Flags().BoolVar(&cmdOpts.ForceDocs, "force-docs", false, "Écraser docs substantiels")
	cmd.Flags().BoolVar(&cmdOpts.DryRun, "dry-run", false, "Afficher les changements sans écrire")
	cmd.Flags().BoolVar(&cmdOpts.UI, "ui", false, "Ouvrir le wizard TUI")
	return cmd
}

// ReadyCommand returns the `asa ready` command.
func ReadyCommand() *cobra.Command {
	opts := onboarding.Options{CheckOnly: true}
	cmd := &cobra.Command{
		Use:     "ready",
		Short:   "Score de readiness du dépôt",
		Example: "  asa ready\n  asa ready --json\n  asa ready --strict --plain\n  asa ready --autofix",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			_, err = onboarding.Ready(cwd, opts, cmd.OutOrStdout())
			return err
		},
	}
	cmd.Flags().BoolVar(&opts.Plain, "plain", false, "Sortie texte plain")
	cmd.Flags().BoolVar(&opts.JSON, "json", false, "Sortie JSON")
	cmd.Flags().BoolVar(&opts.CI, "ci", false, "Mode CI (plain)")
	cmd.Flags().BoolVar(&opts.Strict, "strict", false, "Traiter les warns comme des échecs")
	cmd.Flags().BoolVar(&opts.Autofix, "autofix", false, "Appliquer les corrections automatiques sûres (.gitignore, etc.)")
	return cmd
}
