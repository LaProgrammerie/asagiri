package cloud

import (
	"encoding/json"
	"fmt"
	"io"
)

// FormatStatusJSON writes status report as JSON.
func FormatStatusJSON(w io.Writer, report StatusReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatStatusText writes human-readable status.
func FormatStatusText(w io.Writer, report StatusReport) error {
	_, _ = fmt.Fprintf(w, "Cloud Team (opt-in)\n")
	_, _ = fmt.Fprintf(w, "  activé      : %v\n", report.Enabled)
	_, _ = fmt.Fprintf(w, "  base_url    : %s\n", report.BaseURL)
	_, _ = fmt.Fprintf(w, "  token       : %s\n", tokenLabel(report.TokenPresent))
	_, _ = fmt.Fprintf(w, "  token_path  : %s\n", report.TokenPath)
	_, _ = fmt.Fprintf(w, "  projet lié  : %v\n", report.Linked)
	if report.ProjectID != "" {
		_, _ = fmt.Fprintf(w, "  project_id  : %s\n", report.ProjectID)
	}
	if report.Reachable && report.Me != nil {
		_, _ = fmt.Fprintf(w, "  API         : OK (%s)\n", report.Me.Email)
	} else if report.TokenPresent {
		_, _ = fmt.Fprintf(w, "  API         : non vérifiée\n")
	}
	if report.Error != "" {
		_, _ = fmt.Fprintf(w, "  erreur      : %s\n", report.Error)
	}
	if !report.Enabled {
		_, _ = fmt.Fprintln(w, "\n→ cloud désactivé — le CLI reste 100 % local")
	} else if !report.Linked {
		_, _ = fmt.Fprintln(w, "\n→ asa cloud link <project-id>")
	} else if !report.TokenPresent {
		_, _ = fmt.Fprintln(w, "\n→ asa cloud login --token <token>")
	}
	return nil
}

func tokenLabel(present bool) string {
	if present {
		return "présent"
	}
	return "absent"
}

// FormatPushJSON writes push report as JSON.
func FormatPushJSON(w io.Writer, report PushReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatPushText writes human-readable push summary.
func FormatPushText(w io.Writer, report PushReport) error {
	_, _ = fmt.Fprintf(w, "Cloud push (%s)\n", report.Mode)
	_, _ = fmt.Fprintf(w, "  project_id : %s\n", report.ProjectID)
	_, _ = fmt.Fprintf(w, "  base_url   : %s\n", report.BaseURL)
	_, _ = fmt.Fprintf(w, "  runs       : %d\n", len(report.Items))
	for _, item := range report.Items {
		_, _ = fmt.Fprintf(w, "\n- %s (%d entrées, %s)\n", item.LocalRunID, item.EntryCount, item.RunStatus)
		if item.CloudRunID != "" {
			_, _ = fmt.Fprintf(w, "  cloud_run_id: %s\n", item.CloudRunID)
		}
		if item.Error != "" {
			_, _ = fmt.Fprintf(w, "  erreur: %s\n", item.Error)
		}
	}
	if report.Hint != "" {
		_, _ = fmt.Fprintf(w, "\n→ %s\n", report.Hint)
	}
	return nil
}

// FormatLinkJSON writes link report as JSON.
func FormatLinkJSON(w io.Writer, report LinkReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatLinkText writes human-readable link summary.
func FormatLinkText(w io.Writer, report LinkReport) error {
	_, _ = fmt.Fprintf(w, "Projet cloud lié\n")
	_, _ = fmt.Fprintf(w, "  project_id : %s\n", report.ProjectID)
	if report.ProjectName != "" {
		_, _ = fmt.Fprintf(w, "  name       : %s\n", report.ProjectName)
	}
	if report.ProjectSlug != "" {
		_, _ = fmt.Fprintf(w, "  slug       : %s\n", report.ProjectSlug)
	}
	_, _ = fmt.Fprintf(w, "  config     : %s\n", report.ConfigPath)
	return nil
}
