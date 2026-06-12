package doctorcli

import (
	"bytes"
	"encoding/json"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/doctor"
	"github.com/LaProgrammerie/asagiri/application/internal/reportsink"
	"github.com/LaProgrammerie/asagiri/application/internal/trustcli"
	"github.com/spf13/cobra"
)

// ErrFailed is returned when doctor exits non-zero in strict or blocking mode.
var ErrFailed = doctorExitError{}

type doctorExitError struct{}

func (doctorExitError) Error() string { return "doctor: contrôle bloquant ou mode strict" }

// RootCommand returns the `asa doctor` command tree.
func RootCommand() *cobra.Command {
	var full bool
	var jsonOut bool
	var strict bool
	var save bool
	cmd := &cobra.Command{
		Use:     "doctor",
		Short:   "Vérifier l'environnement et la configuration",
		Example: "  asa init\n  asa doctor\n  asa doctor --full\n  asa doctor --json\n  asa doctor --full --strict\n  asa doctor --save",
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			report, err := doctor.Build(cwd, doctor.Options{Full: full})
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			var textErr error
			if jsonOut {
				if err := doctor.FormatJSON(out, report); err != nil {
					return err
				}
			} else {
				textErr = doctor.FormatText(out, report, strict)
			}
			if save {
				if err := saveDoctorReport(cmd, report); err != nil {
					return err
				}
			}
			if jsonOut && doctor.ShouldFail(report, strict) {
				return ErrFailed
			}
			if textErr != nil {
				return textErr
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&full, "full", false, "Inclure checks onboarding (gitignore, agents, docs, kiro)")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du rapport doctor sur stdout")
	cmd.Flags().BoolVar(&strict, "strict", false, "Code de sortie non nul si des avertissements sont présents")
	cmd.Flags().BoolVar(&save, "save", false, "Enregistrer le rapport JSON sous .asagiri/reports/doctor/latest.json (confirmation sur stderr)")
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	cmd.AddCommand(newDoctorDiffCmd(), newDoctorArchitectureCmd())
	return cmd
}

func saveDoctorReport(cmd *cobra.Command, report doctor.Report) error {
	repoRoot := report.Repository.GitRoot
	if repoRoot == "" {
		return reportsink.ErrRuntimeAbsent
	}
	rel, err := reportsink.SaveDoctor(repoRoot, report)
	if err != nil {
		return err
	}
	trustcli.PrintReportSaved(cmd, rel)
	return nil
}

// DecodeJSON parses a doctor report from JSON bytes (test helper).
func DecodeJSON(data []byte) (doctor.Report, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var report doctor.Report
	err := dec.Decode(&report)
	return report, err
}
