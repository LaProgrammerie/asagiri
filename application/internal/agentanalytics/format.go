package agentanalytics

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatJSON writes the stats report as JSON.
func FormatJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatText renders a human-readable stats summary.
func FormatText(w io.Writer, report Report) error {
	if _, err := fmt.Fprintf(w, "Asagiri Agent Stats\n"); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "Ledger : %s\n", report.LedgerPath)
	if f := formatFilter(report.Filter); f != "" {
		_, _ = fmt.Fprintf(w, "Filtre : %s\n", f)
	}
	_, _ = fmt.Fprintln(w)
	if err := formatStatsBlock(w, "Global", report.Global); err != nil {
		return err
	}
	if len(report.ByAgent) > 0 {
		if _, err := fmt.Fprintln(w, "Par agent"); err != nil {
			return err
		}
		for _, g := range report.ByAgent {
			if err := formatStatsBlock(w, "  "+g.ID, g.Stats); err != nil {
				return err
			}
		}
	}
	if len(report.ByProvider) > 0 {
		if _, err := fmt.Fprintln(w, "Par provider"); err != nil {
			return err
		}
		for _, g := range report.ByProvider {
			if err := formatStatsBlock(w, "  "+g.ID, g.Stats); err != nil {
				return err
			}
		}
	}
	return nil
}

func formatFilter(f Filter) string {
	parts := []string{}
	if strings.TrimSpace(f.AgentID) != "" {
		parts = append(parts, "agent="+f.AgentID)
	}
	if strings.TrimSpace(f.Provider) != "" {
		parts = append(parts, "provider="+f.Provider)
	}
	return strings.Join(parts, " ")
}

func formatStatsBlock(w io.Writer, label string, s Stats) error {
	contract := "—"
	if s.ContractValidRatio != nil {
		contract = fmt.Sprintf("%.0f%% (%d/%d)", *s.ContractValidRatio*100, s.ContractValid, s.ContractChecked)
	}
	last := s.LastRunAt
	if last == "" {
		last = "—"
	}
	_, err := fmt.Fprintf(w, "%s: runs=%d  ok=%d  fail=%d  avg_ms=%.0f  p95_ms=%d  contract=%s  last=%s\n",
		label, s.TotalRuns, s.SuccessCount, s.FailureCount, s.AvgDurationMS, s.P95DurationMS, contract, last)
	return err
}
