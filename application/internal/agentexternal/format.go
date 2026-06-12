package agentexternal

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatJSON writes report as JSON to w (stdout contract).
func FormatJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatText renders a human-readable audit report.
func FormatText(w io.Writer, report Report) error {
	_, err := fmt.Fprintf(w, "Asagiri Agents External (audit read-only)\n\n")
	if err != nil {
		return err
	}
	if p := strings.TrimSpace(report.Policy); p != "" {
		_, err = fmt.Fprintf(w, "Politique : %s\n\n", p)
		if err != nil {
			return err
		}
	}
	for _, t := range report.Targets {
		if err := formatTarget(w, t); err != nil {
			return err
		}
	}
	if len(report.Notes) > 0 {
		_, err = fmt.Fprintf(w, "\nNotes :\n")
		if err != nil {
			return err
		}
		for _, n := range report.Notes {
			_, err = fmt.Fprintf(w, "  • %s\n", n)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func formatTarget(w io.Writer, t ExternalTarget) error {
	label := t.AgentID
	if ck := strings.TrimSpace(t.ConfigKey); ck != "" && ck != t.AgentID {
		label = fmt.Sprintf("%s (config %s)", t.AgentID, ck)
	}
	_, err := fmt.Fprintf(w, "• %s  [%s]\n", label, t.Status)
	if err != nil {
		return err
	}
	if pt := strings.TrimSpace(t.Provider); pt != "" {
		_, err = fmt.Fprintf(w, "    provider=%s support=%s cli=%s available=%v\n",
			pt, t.SupportLevel, t.CLICommand, t.CLIAvailable)
		if err != nil {
			return err
		}
	}
	if p := strings.TrimSpace(t.DetectedPath); p != "" {
		_, err = fmt.Fprintf(w, "    path=%s writable=%v source=%s\n", p, t.Writable, t.PathSource)
		if err != nil {
			return err
		}
	} else if cp := strings.TrimSpace(t.ConfiguredPath); cp != "" {
		_, err = fmt.Fprintf(w, "    configured=%s source=%s\n", cp, t.PathSource)
		if err != nil {
			return err
		}
	}
	if h := strings.TrimSpace(t.InstalledHash); h != "" {
		_, err = fmt.Fprintf(w, "    installed_hash=%s desired=%s last_synced=%s\n",
			truncateHash(h), truncateHash(t.DesiredHash), truncateHash(t.LastSyncedHash))
		if err != nil {
			return err
		}
	}
	if d := strings.TrimSpace(t.Detail); d != "" {
		_, err = fmt.Fprintf(w, "    %s\n", d)
	}
	return err
}

func truncateHash(h string) string {
	h = strings.TrimSpace(h)
	if h == "" {
		return "—"
	}
	if len(h) <= 12 {
		return h
	}
	return h[:12] + "…"
}
