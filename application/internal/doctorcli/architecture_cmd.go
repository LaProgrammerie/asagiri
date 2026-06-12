package doctorcli

import (
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/doctor"
	"github.com/spf13/cobra"
)

func newDoctorArchitectureCmd() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "architecture",
		Short: "Croiser tasks, graphs, knowledge, trust et agent ledger (lecture seule)",
		Example: `  asa doctor architecture
  asa doctor architecture --json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			report, err := doctor.BuildArchitecture(cwd)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				return doctor.FormatArchitectureJSON(out, report)
			}
			return doctor.FormatArchitectureText(out, report)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du rapport architecture sur stdout")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	return cmd
}

// DecodeArchitectureJSON parses an architecture report from JSON bytes (test helper).
func DecodeArchitectureJSON(data []byte) (doctor.ArchitectureReport, error) {
	return doctor.DecodeArchitectureJSON(data)
}
