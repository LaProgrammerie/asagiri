package cli

import (
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/spf13/cobra"
)

func newNextCmd(dryRun *bool) *cobra.Command {
	var feature string
	var execute bool

	cmd := &cobra.Command{
		Use:   "next",
		Short: "Afficher ou exécuter la prochaine action recommandée",
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
			if feature == "" {
				feature = snap.ActiveFeature
			}
			rec, err := intent.RecommendNext(snap, feature, ctx.Config)
			if err != nil {
				return err
			}
			intent.PrintNextRecommendation(cmd.OutOrStdout(), rec)
			if !execute {
				return nil
			}
			parts := strings.Fields(strings.TrimPrefix(rec.Primitive, "asa "))
			if len(parts) == 0 {
				return nil
			}
			root := newRootCmd()
			root.SetArgs(append(parts, "--dry-run"))
			if !ctx.DryRun {
				root.SetArgs(parts)
			}
			return root.Execute()
		},
	}
	cmd.Flags().StringVar(&feature, "feature", "", "Feature cible")
	cmd.Flags().BoolVar(&execute, "execute", false, "Exécuter la commande recommandée")
	return cmd
}
