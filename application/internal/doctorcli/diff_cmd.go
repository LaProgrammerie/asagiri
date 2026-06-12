package doctorcli

import (
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/doctor"
	"github.com/LaProgrammerie/asagiri/application/internal/reportdiff"
	"github.com/LaProgrammerie/asagiri/application/internal/reportsink"
	"github.com/LaProgrammerie/asagiri/application/internal/trustcli"
	"github.com/spf13/cobra"
)

func newDoctorDiffCmd() *cobra.Command {
	var jsonOut bool
	var fromPath, toPath string
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Comparer deux snapshots doctor enregistrés",
		Long: `Compare latest.json avec l'entrée d'historique la plus récente.

Nécessite au moins deux sauvegardes (asa doctor --save).`,
		Example: `  asa doctor diff
  asa doctor diff --json
  asa doctor diff --from .asagiri/reports/doctor/history/20260608T120000Z.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cwd, err := os.Getwd()
			if err != nil {
				return err
			}
			report, err := doctor.Build(cwd, doctor.Options{})
			if err != nil {
				return err
			}
			repoRoot := report.Repository.GitRoot
			if repoRoot == "" {
				return reportsink.ErrRuntimeAbsent
			}
			latestRel := reportsink.DoctorLatestRel()
			beforeRel, afterRel, err := trustcli.ResolveDiffPaths(repoRoot, latestRel, fromPath, toPath)
			if err != nil {
				return err
			}
			var before, after doctor.Report
			if err := reportsink.ReadJSONFile(repoRoot, beforeRel, &before); err != nil {
				return err
			}
			if err := reportsink.ReadJSONFile(repoRoot, afterRel, &after); err != nil {
				return err
			}
			diff := reportdiff.DiffDoctor(before, after, reportdiff.ReportPaths{Before: beforeRel, After: afterRel})
			out := cmd.OutOrStdout()
			if jsonOut {
				return reportdiff.FormatJSON(out, diff)
			}
			reportdiff.FormatDoctorText(out, diff)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du diff sur stdout")
	cmd.Flags().StringVar(&fromPath, "from", "", "Snapshot source (défaut: dernière entrée history/)")
	cmd.Flags().StringVar(&toPath, "to", "", "Snapshot cible (défaut: latest.json)")
	return cmd
}
