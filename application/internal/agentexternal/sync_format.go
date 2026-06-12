package agentexternal

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatSyncJSON writes sync report as JSON to w.
func FormatSyncJSON(w io.Writer, report SyncReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatSyncText renders a human-readable external sync report.
func FormatSyncText(w io.Writer, report SyncReport) error {
	_, err := fmt.Fprintf(w, "Asagiri Agents External Sync (%s)\n\n", report.Mode)
	if err != nil {
		return err
	}
	for _, it := range report.Items {
		if err := formatSyncItem(w, it); err != nil {
			return err
		}
	}
	if h := strings.TrimSpace(report.Hint); h != "" {
		_, err = fmt.Fprintf(w, "\n→ %s\n", h)
	}
	return err
}

func formatSyncItem(w io.Writer, it SyncItem) error {
	label := it.AgentID
	if ck := strings.TrimSpace(it.ConfigKey); ck != "" && ck != it.AgentID {
		label = fmt.Sprintf("%s (config %s)", it.AgentID, ck)
	}
	_, err := fmt.Fprintf(w, "• %s  %s", label, it.Action)
	if err != nil {
		return err
	}
	if p := strings.TrimSpace(it.Path); p != "" {
		if _, err = fmt.Fprintf(w, "  %s", p); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w)
	if err != nil {
		return err
	}
	if msg := strings.TrimSpace(it.Message); msg != "" {
		_, err = fmt.Fprintf(w, "    %s\n", msg)
	}
	return err
}
