package cli

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/contextopt"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/cost"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/tui"
)

func mustWd() string {
	wd, err := os.Getwd()
	if err != nil {
		return "."
	}
	return wd
}

func printEstimateBoxed(w io.Writer, est cost.ExecutionEstimate, opt *contextopt.OptimizeResult) {
	cfg := &config.Config{UI: config.UIConfig{Mode: "plain"}}
	ui := tui.NewUI(cfg, w, tui.DetectTTY(w))
	body := fmt.Sprintf("Context: ~%d in / ~%d out tokens\n", est.TotalInputTokens, est.TotalOutputTokens)
	if opt != nil && opt.OriginalTokens > 0 {
		body += fmt.Sprintf("Context savings: %.1f%%\n", opt.SavingsRatio*100)
	}
	body += fmt.Sprintf("Cost: %s\nTime: %s\nBudget: %s\nRisk: medium\n",
		formatMoneyEUR(est.EstimatedCost), est.EstimatedDuration.Round(time.Second), est.BudgetStatus)
	for _, warn := range est.Warnings {
		body += "Warning: " + warn + "\n"
	}
	ui.Box("Estimated execution", body)
}

func formatMoneyEUR(m cost.Money) string {
	cur := m.Currency
	if cur == "" {
		cur = "EUR"
	}
	return fmt.Sprintf("%.2f %s", float64(m.Cents)/100, cur)
}
