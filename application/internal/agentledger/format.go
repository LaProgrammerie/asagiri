package agentledger

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatJSON writes the runs report as JSON.
func FormatJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatText renders a human-readable runs list.
func FormatText(w io.Writer, report Report) error {
	if _, err := fmt.Fprintf(w, "Asagiri Agent Runs (%d)\n", report.Count); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(w, "Ledger : %s\n\n", report.LedgerPath)
	if report.Count == 0 {
		_, err := fmt.Fprintln(w, "Aucune exécution enregistrée.")
		return err
	}
	for _, e := range report.Entries {
		if err := formatEntryText(w, e); err != nil {
			return err
		}
	}
	return nil
}

func formatEntryText(w io.Writer, e Entry) error {
	contract := "—"
	if e.ContractValid != nil {
		if *e.ContractValid {
			contract = "valid"
		} else {
			contract = "invalid"
		}
	}
	_, err := fmt.Fprintf(w, "• %s  task=%s  run=%s  feature=%s\n", e.AgentID, e.TaskID, e.RunID, e.Feature)
	if err != nil {
		return err
	}
	phase := e.Phase
	if phase == "" {
		phase = "—"
	}
	role := e.Role
	if role == "" {
		role = "—"
	}
	provider := e.Provider
	if provider == "" {
		provider = "—"
	}
	_, err = fmt.Fprintf(w, "  phase=%s  role=%s  provider=%s  exit=%d  duration_ms=%d\n", phase, role, provider, e.ExitCode, e.DurationMS)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "  prompt_hash=%s  context_hash=%s  output_hash=%s  contract=%s\n",
		shortHash(e.PromptHash), shortHash(e.ContextHash), shortHash(e.OutputHash), contract)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "  log_dir=%s  started=%s\n\n", e.LogDir, e.StartedAt)
	return err
}

func shortHash(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return "—"
	}
	if len(h) <= 12 {
		return h
	}
	return h[:12] + "…"
}
