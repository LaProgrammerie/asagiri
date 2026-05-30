package cli

import (
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	var full bool
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Vérifier l'environnement et la configuration",
		Example: "  asa init\n  asa doctor\n  asa doctor --full",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			checks, err := bootstrap.DoctorWithOptions(cwd, bootstrap.DoctorOptions{Full: full})
			if err != nil {
				return err
			}
			return bootstrap.FormatDoctor(cmd.OutOrStdout(), checks)
		},
	}
	cmd.Flags().BoolVar(&full, "full", false, "Inclure checks onboarding (gitignore, agents, docs, kiro)")
	return cmd
}
