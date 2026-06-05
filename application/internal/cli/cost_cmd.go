package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/spf13/cobra"
)

func newCostCmd(dryRun *bool) *cobra.Command {
	cost := &cobra.Command{Use: "cost", Short: "Coûts et modèles"}
	cost.AddCommand(newCostReportCmd(dryRun), newCostModelsCmd(dryRun), newCostTrendsCmd(dryRun))
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
			tokens, err := c.Store.QueryStepTokens(context.Background(), sinceT)
			if err != nil {
				return err
			}
			sav := cost.ComputeSavings(tokens, tot.ActualCostCents, c.Config)
			printCostReport(cmd.OutOrStdout(), days, tot, sav)
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
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Currency: %s\n", c.Config.Pricing.Currency)
			for name, p := range c.Config.Pricing.Models {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  %s: in=%.4f/1M out=%.4f/1M (%s)\n",
					name, p.InputPer1MTokens, p.OutputPer1MTokens, p.Source)
			}
			for id, prof := range c.Config.Models {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "  profile %s: class=%s model=%s\n", id, prof.Class, prof.Model)
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

func stepMixPercent(st telemetry.StepTotals) (local, cloud float64) {
	total := st.LocalSteps + st.CloudSteps
	if total == 0 {
		return 0, 0
	}
	return float64(st.LocalSteps) / float64(total) * 100, float64(st.CloudSteps) / float64(total) * 100
}

func printCostReport(out interface{ Write([]byte) (int, error) }, days int, tot telemetry.CostTotals, sav cost.SavingsReport) {
	w := fmt.Sprintf // alias
	_ = w

	fmt.Fprintf(out, "Last %d days\n%s\n", days, repeatDash(20))
	fmt.Fprintf(out, "Runs: %d\n", tot.RunCount)
	fmt.Fprintf(out, "Actual cost:     %s\n", cost.FormatCents(sav.ActualCostCents, sav.Currency))

	// Local / cloud ratio — always shown when there is token data
	total := sav.TotalInputTokens() + sav.TotalOutputTokens()
	if total > 0 {
		localPct := sav.LocalPct()
		cloudPct := 100 - localPct
		localTok := sav.LocalInputTokens + sav.LocalOutputTokens
		cloudTok := sav.CloudInputTokens + sav.CloudOutputTokens
		fmt.Fprintf(out, "\nLocal / cheap:   %.0f%% (%d tokens) — no LLM cost\n", localPct, localTok)
		fmt.Fprintf(out, "Cloud / premium: %.0f%% (%d tokens)\n", cloudPct, cloudTok)
	}

	// Savings — only when premium_reference_model is configured
	if sav.HasPremiumBaseline() {
		fmt.Fprintf(out, "\nSavings (vs %s)\n%s\n", sav.PremiumReferenceModel, repeatDash(20))
		fmt.Fprintf(out, "Premium equivalent: %s\n", cost.FormatCents(sav.PremiumEquivCents, sav.Currency))
		fmt.Fprintf(out, "Savings:            %s\n", cost.FormatCents(sav.SavingsCents, sav.Currency))
		fmt.Fprintf(out, "Savings rate:       %.1f%%\n", sav.SavingsRate*100)
	}
}

func newCostTrendsCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "trends",
		Short: "Évolution de l'efficience sur deux fenêtres disjointes",
		Long:  "Compare deux fenêtres strictement disjointes : J-30→J-15 (précédente) et J-15→maintenant (actuelle).",
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := loadContext(mustWd(), *dryRun)
			if err != nil {
				return err
			}
			defer c.Close()
			ctx := context.Background()
			now := time.Now()
			// Two strictly disjoint 15-day windows: [J-30, J-15) and [J-15, now)
			// No run can appear in both windows.
			mid := now.AddDate(0, 0, -15)
			start := now.AddDate(0, 0, -30)
			prevW, err := buildWindowBetween(ctx, c.Store, c.Config, start, mid)
			if err != nil {
				return err
			}
			currW, err := buildWindowBetween(ctx, c.Store, c.Config, mid, now)
			if err != nil {
				return err
			}
			printTrendsReport(cmd.OutOrStdout(), prevW, currW, start, mid, now)
			return nil
		},
	}
	return cmd
}


func printTrendsReport(out interface{ Write([]byte) (int, error) }, prev, curr cost.WindowReport, start, mid, end time.Time) {
	const dateFmt = "2006-01-02"
	fmt.Fprintf(out, "Cost efficiency trends\n%s\n", repeatDash(36))
	fmt.Fprintf(out, "\nPrevious window:\n  %s → %s\n", start.Format(dateFmt), mid.Format(dateFmt))
	fmt.Fprintf(out, "Current window:\n  %s → %s\n", mid.Format(dateFmt), end.Format(dateFmt))
	cur := curr.Savings.Currency
	fmt.Fprintf(out, "\nLocal-first rate\n  %.0f%% → %.0f%%\n",
		prev.Savings.LocalPct(), curr.Savings.LocalPct())
	fmt.Fprintf(out, "\nActual spend\n  %s → %s\n",
		cost.FormatCents(prev.ActualCostCents, cur),
		cost.FormatCents(curr.ActualCostCents, cur))
	fmt.Fprintf(out, "\nAvg cost / run\n  %s → %s\n",
		cost.FormatCents(prev.AvgCostPerRunCents, cur),
		cost.FormatCents(curr.AvgCostPerRunCents, cur))
	fmt.Fprintf(out, "\nStrategy score\n  %s → %s\n",
		prev.Strategy.Grade, curr.Strategy.Grade)
	fmt.Fprintf(out, "\nEscalation rate\n  %.0f%% → %.0f%%\n",
		prev.Escalations.EscalationRate*100, curr.Escalations.EscalationRate*100)
	if prev.Savings.HasPremiumBaseline() && curr.Savings.HasPremiumBaseline() {
		fmt.Fprintf(out, "\nSavings (vs %s)\n  %s → %s\n",
			curr.Savings.PremiumReferenceModel,
			cost.FormatCents(prev.Savings.SavingsCents, cur),
			cost.FormatCents(curr.Savings.SavingsCents, cur))
	}
}


// buildWindowBetween constructs a WindowReport for the strictly bounded window [since, until).
// Bounds: started_at >= since AND started_at < until.
// A run with started_at == until belongs to the next window, not this one.
func buildWindowBetween(ctx context.Context, store interface {
	QueryStepTokensBetween(context.Context, time.Time, time.Time) (telemetry.StepTokenTotals, error)
	SummarizeStepsBetween(context.Context, time.Time, time.Time) (telemetry.StepTotals, error)
	QueryRunsBetween(context.Context, time.Time, time.Time) ([]telemetry.RunMetric, error)
}, cfg *config.Config, since, until time.Time) (cost.WindowReport, error) {
	tokens, err := store.QueryStepTokensBetween(ctx, since, until)
	if err != nil {
		return cost.WindowReport{}, err
	}
	steps, err := store.SummarizeStepsBetween(ctx, since, until)
	if err != nil {
		return cost.WindowReport{}, err
	}
	runs, err := store.QueryRunsBetween(ctx, since, until)
	if err != nil {
		return cost.WindowReport{}, err
	}
	actualCents := int64(0)
	for _, r := range runs {
		actualCents += r.ActualCostCents
	}
	label := since.Format("01-02") + "→" + until.Format("01-02")
	tot := telemetry.CostTotals{RunCount: len(runs), ActualCostCents: actualCents}
	return cost.BuildWindowReport(label, tot, tokens, steps, cfg), nil
}
