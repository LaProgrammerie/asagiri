package doctor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// DecodeArchitectureJSON parses an architecture report from JSON bytes.
func DecodeArchitectureJSON(data []byte) (ArchitectureReport, error) {
	dec := json.NewDecoder(bytes.NewReader(data))
	var report ArchitectureReport
	err := dec.Decode(&report)
	return report, err
}

// FormatArchitectureJSON writes the architecture doctor report as JSON.
func FormatArchitectureJSON(w io.Writer, report ArchitectureReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(report); err != nil {
		return fmt.Errorf("doctor architecture json: %w", err)
	}
	return nil
}

// FormatArchitectureText writes a concise human-readable summary.
func FormatArchitectureText(w io.Writer, report ArchitectureReport) error {
	var b strings.Builder
	b.WriteString("Asagiri Doctor — architecture\n\n")
	if report.Repository.GitRoot != "" {
		fmt.Fprintf(&b, "Repository: %s\n", report.Repository.GitRoot)
	}
	fmt.Fprintf(&b, "Tasks: %d | Execution graphs: %d (%d nodes) | Knowledge nodes: %d\n",
		report.Summary.Tasks, report.Summary.ExecutionGraphs, report.Summary.ExecutionGraphNodes, report.Summary.KnowledgeNodes)
	fmt.Fprintf(&b, "Trust reports: %d | Agent ledger: %d\n\n",
		report.Summary.TrustReports, report.Summary.AgentLedgerEntries)

	if len(report.Findings) == 0 {
		b.WriteString("Aucun écart détecté.\n")
	} else {
		b.WriteString("Écarts:\n")
		for _, f := range report.Findings {
			fmt.Fprintf(&b, "  • [%s] %s", f.Kind, f.Message)
			if f.TaskID != "" {
				fmt.Fprintf(&b, " (task=%s)", f.TaskID)
			}
			if f.NodeID != "" {
				fmt.Fprintf(&b, " (graph=%s node=%s)", f.GraphID, f.NodeID)
			}
			if f.Flow != "" && f.TaskID == "" && f.NodeID == "" {
				fmt.Fprintf(&b, " (flow=%s)", f.Flow)
			}
			b.WriteByte('\n')
		}
	}

	if len(report.Recommendations) > 0 {
		b.WriteString("\nRecommandations:\n")
		for _, a := range report.Recommendations {
			fmt.Fprintf(&b, "  → %s\n    %s\n", a.Title, a.CLI)
		}
	}
	_, err := io.WriteString(w, b.String())
	return err
}
