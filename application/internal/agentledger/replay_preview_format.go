package agentledger

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// FormatReplayPreviewJSON writes a replay preview report as JSON.
func FormatReplayPreviewJSON(w io.Writer, report ReplayPreviewReport) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

// FormatReplayPreviewText renders a human-readable replay preview.
func FormatReplayPreviewText(w io.Writer, report ReplayPreviewReport) error {
	if _, err := fmt.Fprintf(w, "Asagiri Agent Run Preview %s\n\n", report.RunID); err != nil {
		return err
	}
	lines := []string{
		fmt.Sprintf("task_id      : %s", report.TaskID),
		fmt.Sprintf("agent_id     : %s", report.AgentID),
		fmt.Sprintf("provider     : %s", dashIfEmpty(report.Provider)),
		fmt.Sprintf("log_dir      : %s", report.LogDir),
		fmt.Sprintf("prompt_hash  : %s", report.PromptHash),
		fmt.Sprintf("context_hash : %s", dashIfEmpty(report.ContextHash)),
		fmt.Sprintf("output_hash  : %s", report.OutputHash),
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
		if err := formatPreviewArtifactText(w, a); err != nil {
			return err
		}
	}
	return nil
}

func formatPreviewArtifactText(w io.Writer, a PreviewArtifact) error {
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
	_, err = fmt.Fprintf(w, "  size_bytes=%d  modified_at=%s\n", size, a.ModifiedAt)
	if err != nil {
		return err
	}
	if strings.TrimSpace(a.Content) == "" {
		_, err = fmt.Fprintln(w)
		return err
	}
	label := "content"
	if a.Name == "prompt.md" {
		label = "prompt"
	}
	_, err = fmt.Fprintf(w, "  %s:\n%s\n\n", label, indentBlock(a.Content, "    "))
	return err
}

func indentBlock(s, prefix string) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	for i, line := range lines {
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
