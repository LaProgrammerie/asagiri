package agentledger

import (
	"encoding/json"
	"fmt"
	"io"
)

// FormatExportJSON writes the export report as JSON.
func FormatExportJSON(w io.Writer, report ExportReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatExportText renders a human-readable export summary.
func FormatExportText(w io.Writer, report ExportReport) error {
	if _, err := fmt.Fprintf(w, "Asagiri Agent Run Export %s\n", report.RunID); err != nil {
		return err
	}
	_, err := fmt.Fprintf(w, "Output : %s\nManifest : %s\nFiles : %d\n",
		report.OutputDir, report.ManifestPath, len(report.Files))
	return err
}
