package cli

import (
	"fmt"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/spf13/cobra"
)

func newContinueCmd(dryRun *bool) *cobra.Command {
	var feature, runID string
	var yes bool

	cmd := &cobra.Command{
		Use:   "continue",
		Short: "Reprendre le travail le plus pertinent",
		RunE: func(cmd *cobra.Command, args []string) error {
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			ctx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()

			snap, err := ctx.snapshot()
			if err != nil {
				return err
			}

			if feature == "" && runID == "" {
				if run, err := intent.FindResumableRun(snap, "", ""); err == nil {
					runID = run.ID
					feature = run.Feature
				} else if snap.ActiveFeature != "" {
					feature = snap.ActiveFeature
				}
			}
			if feature == "" {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aucune feature à reprendre. Utilisez: asa inbox")
				return nil
			}

			rec, err := intent.RecommendNext(snap, feature, ctx.Config)
			if err != nil {
				return err
			}

			intent.PrintContinueFromRecommendation(cmd.OutOrStdout(), rec)

			if yes {
				fs := intent.FeatureStateFor(snap, feature)
				if fs.PendingGate != nil && fs.PendingGate.Blocking && fs.PendingGate.Phase == gates.PendingPhaseSubmit {
					return fmt.Errorf("complete the gate action before --yes (see Next step above)")
				}
				work := newWorkCmd(dryRun)
				workArgs := []string{fmt.Sprintf("reprends %s", feature), "--yes"}
				work.SetArgs(workArgs)
				return work.Execute()
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&feature, "feature", "", "Feature à reprendre")
	cmd.Flags().StringVar(&runID, "run", "", "Run à reprendre")
	cmd.Flags().BoolVar(&yes, "yes", false, "Exécuter la prochaine action")
	return cmd
}
