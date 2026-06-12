package agentslist

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatJSON writes a single JSON document to w (stdout with --json).
func FormatJSON(w io.Writer, report Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatShowJSON writes one entry as JSON.
func FormatShowJSON(w io.Writer, entry Entry) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(entry)
}

// FormatText renders a human-readable agents list.
func FormatText(w io.Writer, report Report) error {
	if _, err := fmt.Fprintf(w, "Asagiri Agents (%d)\n", len(report.Agents)); err != nil {
		return err
	}
	if report.Registry.UsingEmbeddedDefaults {
		_, _ = fmt.Fprintf(w, "Registry : templates embarqués (%s absent ou vide)\n\n", report.Registry.Path)
	} else {
		_, _ = fmt.Fprintf(w, "Registry : %s (%d fichier(s))\n\n", report.Registry.Path, report.Registry.FileCount)
	}
	for _, entry := range report.Agents {
		if err := formatEntryText(w, entry); err != nil {
			return err
		}
	}
	return nil
}

// FormatShowText renders one agent entry.
func FormatShowText(w io.Writer, entry Entry) error {
	if _, err := fmt.Fprintf(w, "Agent %s\n", entry.ID); err != nil {
		return err
	}
	return formatEntryText(w, entry)
}

func formatEntryText(w io.Writer, entry Entry) error {
	_, err := fmt.Fprintf(w, "• %s  role=%s  version=%s  source=%s\n", entry.ID, entry.Role, entry.Version, entry.Source)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, "  hash=%s  output=%s\n", entry.ContentHash, entry.OutputFormat)
	if err != nil {
		return err
	}
	if entry.StoredHash != "" {
		_, err = fmt.Fprintf(w, "  stored_hash=%s\n", entry.StoredHash)
		if err != nil {
			return err
		}
	}
	if len(entry.ProviderTargets) > 0 {
		_, err = fmt.Fprintf(w, "  provider_targets=%s\n", strings.Join(entry.ProviderTargets, ", "))
		if err != nil {
			return err
		}
	}
	if entry.ProviderSupport != nil && entry.ProviderSupport.Summary != "" {
		_, err = fmt.Fprintf(w, "  provider_support=%s\n", entry.ProviderSupport.Summary)
		if err != nil {
			return err
		}
	}
	for _, warn := range entry.Warnings {
		_, err = fmt.Fprintf(w, "  warn: %s\n", warn)
		if err != nil {
			return err
		}
	}
	_, err = fmt.Fprintln(w)
	return err
}
