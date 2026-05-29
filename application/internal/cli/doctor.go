package cli

import (
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/spf13/cobra"
)

func newDoctorCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "doctor",
		Short:   "Vérifier l'environnement et la configuration",
		Example: "  asa init\n  asa doctor",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			checks, err := bootstrap.Doctor(cwd)
			if err != nil {
				return err
			}
			return bootstrap.FormatDoctor(cmd.OutOrStdout(), checks)
		},
	}
}
