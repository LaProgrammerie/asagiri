package agentledger

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatInspectJSON writes an inspect report as JSON.
func FormatInspectJSON(w io.Writer, report InspectReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatInspectText renders a human-readable run inspection.
func FormatInspectText(w io.Writer, report InspectReport) error {
	if _, err := fmt.Fprintf(w, "Asagiri Agent Run %s\n\n", report.RunID); err != nil {
		return err
	}
	contract := "—"
	if report.ContractValid != nil {
		if *report.ContractValid {
			contract = "valid"
		} else {
			contract = "invalid"
		}
	}
	role := dashIfEmpty(report.Role)
	provider := dashIfEmpty(report.Provider)
	phase := dashIfEmpty(report.Phase)
	lines := []string{
		fmt.Sprintf("task_id      : %s", report.TaskID),
		fmt.Sprintf("feature      : %s", report.Feature),
		fmt.Sprintf("agent_id     : %s", report.AgentID),
		fmt.Sprintf("role         : %s", role),
		fmt.Sprintf("provider     : %s", provider),
		fmt.Sprintf("phase        : %s", phase),
		fmt.Sprintf("started_at   : %s", report.StartedAt),
		fmt.Sprintf("ended_at     : %s", report.EndedAt),
		fmt.Sprintf("duration_ms  : %d", report.DurationMS),
		fmt.Sprintf("exit_code    : %d", report.ExitCode),
		fmt.Sprintf("contract     : %s", contract),
		fmt.Sprintf("prompt_hash  : %s", report.PromptHash),
		fmt.Sprintf("context_hash : %s", dashIfEmpty(report.ContextHash)),
		fmt.Sprintf("output_hash  : %s", report.OutputHash),
		fmt.Sprintf("log_dir      : %s", report.LogDir),
	}
	if report.DryRun {
		lines = append(lines, "dry_run      : true")
	}
	for _, line := range lines {
		if _, err := fmt.Fprintln(w, line); err != nil {
			return err
		}
	}
	if _, err := fmt.Fprintln(w, "\nArtefacts"); err != nil {
		return err
	}
	for _, a := range report.Artifacts {
		if err := formatArtifactText(w, a); err != nil {
			return err
		}
	}
	return nil
}

func formatArtifactText(w io.Writer, a Artifact) error {
	status := "absent"
	if a.Exists {
		status = "present"
	}
	_, err := fmt.Fprintf(w, "• %s  %s  path=%s\n", a.Name, status, a.Path)
	if err != nil {
		return err
	}
	if !a.Exists {
		_, err = fmt.Fprintln(w)
		return err
	}
	size := int64(0)
	if a.SizeBytes != nil {
		size = *a.SizeBytes
	}
	_, err = fmt.Fprintf(w, "  size_bytes=%d  modified_at=%s\n\n", size, a.ModifiedAt)
	return err
}

func dashIfEmpty(s string) string {
	if strings.TrimSpace(s) == "" {
		return "—"
	}
	return s
}
