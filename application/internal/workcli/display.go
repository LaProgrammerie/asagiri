package workcli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/contextopt"
	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/tui"
)

// PrintEstimateBoxed renders the V3 execution estimate box.
func PrintEstimateBoxed(w io.Writer, est cost.ExecutionEstimate, opt *contextopt.OptimizeResult) {
	cfg := &config.Config{UI: config.UIConfig{Mode: "plain"}}
	ui := tui.NewUI(cfg, w, tui.DetectTTY(w))
	body := fmt.Sprintf("Context: ~%d in / ~%d out tokens\n", est.TotalInputTokens, est.TotalOutputTokens)
	if opt != nil && opt.OriginalTokens > 0 {
		body += fmt.Sprintf("Context savings: %.1f%%\n", opt.SavingsRatio*100)
	}
	body += fmt.Sprintf("Cost: %s\nTime: %s\nBudget: %s\nConfidence: %.0f%%\n",
		formatMoneyEUR(est.EstimatedCost), est.EstimatedDuration.Round(time.Second), est.BudgetStatus, est.Confidence*100)
	body += explainSteps(est.PlannedSteps)
	for _, warn := range est.Warnings {
		body += "Warning: " + warn + "\n"
	}
	ui.Box("Estimated execution", body)
}

func explainSteps(steps []cost.EstimatedStep) string {
	if len(steps) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\nSteps (why model / tier):\n")
	for _, s := range steps {
		tier := "cloud"
		if s.Local {
			tier = "local"
		}
		reason := s.Reason
		if reason == "" {
			reason = "plan"
		}
		fmt.Fprintf(&b, "  • %s → agent=%s model=%s tier=%s in=%d out=%d — %s\n",
			s.Name, s.Agent, displayModel(s.Model), tier, s.InputTokens, s.OutputTokens, reason)
	}
	return b.String()
}

func displayModel(m string) string {
	if m == "" {
		return "(none)"
	}
	return m
}

// PrintWorkSummary prints the post-execution work summary block.
func PrintWorkSummary(w io.Writer, instruction string, est cost.ExecutionEstimate, exec intent.ExecuteResult) {
	_, _ = fmt.Fprintf(w, "\n── Résumé ──\n")
	_, _ = fmt.Fprintf(w, "Instruction: %s\n", instruction)
	if est.Feature != "" {
		_, _ = fmt.Fprintf(w, "Feature: %s", est.Feature)
		if est.TaskID != "" {
			_, _ = fmt.Fprintf(w, " / task %s", est.TaskID)
		}
		_, _ = fmt.Fprintln(w)
	}
	_, _ = fmt.Fprintf(w, "Estimation: %s, ~%s, budget %s\n",
		formatMoneyEUR(est.EstimatedCost), est.EstimatedDuration.Round(time.Second), est.BudgetStatus)
	if len(exec.Executed) > 0 {
		_, _ = fmt.Fprintf(w, "Exécuté: %d étape(s)\n", len(exec.Executed))
	}
	if exec.LastRunID != "" {
		_, _ = fmt.Fprintf(w, "Dernier run: %s\n", exec.LastRunID)
	}
}

func formatMoneyEUR(m cost.Money) string {
	cur := m.Currency
	if cur == "" {
		cur = "EUR"
	}
	return fmt.Sprintf("%.2f %s", float64(m.Cents)/100, cur)
}
