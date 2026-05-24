package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/spf13/cobra"
)

func newCostCmd(dryRun *bool) *cobra.Command {
	cost := &cobra.Command{Use: "cost", Short: "Coûts et modèles"}
	cost.AddCommand(newCostReportCmd(dryRun), newCostModelsCmd(dryRun))
	return cost
}

func newCostReportCmd(dryRun *bool) *cobra.Command {
	var since string
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Historique des coûts",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			days, err := parseSinceDays(since)
			if err != nil {
				return err
			}
			sinceT := time.Now().AddDate(0, 0, -days)
			tot, err := telemetry.SummarizeSince(context.Background(), c.Store, sinceT)
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Last %d days\n%s\nRuns: %d\nEstimated cost: %.2f\nActual cost: %.2f\n",
				days, repeatDash(11), tot.RunCount,
				float64(tot.EstimatedCostCents)/100,
				float64(tot.ActualCostCents)/100)
			return nil
		},
	}
	cmd.Flags().StringVar(&since, "since", "7d", "Fenêtre (ex. 7d)")
	return cmd
}

func newCostModelsCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "models",
		Short: "Lister modèles et pricing configurés",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			fmt.Fprintf(cmd.OutOrStdout(), "Currency: %s\n", c.Config.Pricing.Currency)
			for name, p := range c.Config.Pricing.Models {
				fmt.Fprintf(cmd.OutOrStdout(), "  %s: in=%.4f/1M out=%.4f/1M (%s)\n",
					name, p.InputPer1MTokens, p.OutputPer1MTokens, p.Source)
			}
			for id, prof := range c.Config.Models {
				fmt.Fprintf(cmd.OutOrStdout(), "  profile %s: class=%s model=%s\n", id, prof.Class, prof.Model)
			}
			return nil
		},
	}
}

func parseSinceDays(s string) (int, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return 7, nil
	}
	var d int
	if _, err := fmt.Sscanf(s, "%dd", &d); err != nil {
		return 7, err
	}
	return d, nil
}

func repeatDash(n int) string {
	return strings.Repeat("-", n)
}
