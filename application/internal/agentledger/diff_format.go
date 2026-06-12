package agentledger

import (
	"encoding/json"
	"fmt"
	"io"
)

// FormatDiffJSON writes a diff report as JSON.
func FormatDiffJSON(w io.Writer, report DiffReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatDiffText renders a human-readable run diff.
func FormatDiffText(w io.Writer, report DiffReport) error {
	status := "différences"
	if report.Identical {
		status = "identiques"
	}
	if _, err := fmt.Fprintf(w, "Asagiri Agent Run Diff\nleft=%s  right=%s  (%s)\n\n", report.LeftRunID, report.RightRunID, status); err != nil {
		return err
	}
	if _, err := fmt.Fprintln(w, "Métadonnées"); err != nil {
		return err
	}
	for _, f := range report.Fields {
		mark := "≠"
		if f.Equal {
			mark = "="
		}
		if _, err := fmt.Fprintf(w, "  %s %s  left=%s  right=%s\n", mark, f.Field, dashIfEmpty(f.Left), dashIfEmpty(f.Right)); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "\nArtefacts"); err != nil {
		return err
	}
	for _, a := range report.Artifacts {
		mark := "≠"
		if a.Equal {
			mark = "="
		}
		_, err := fmt.Fprintf(w, "  %s %s  exists(%t/%t)  size_equal=%t  modified_equal=%t\n",
			mark, a.Name, a.Left.Exists, a.Right.Exists, a.SizeEqual, a.ModifiedAtEqual)
		if err != nil {
			return err
		}
	}
	return nil
}
