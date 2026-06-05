package intent

import (
	"fmt"
	"io"
	"strings"
)

// PrintWorkReport renders boxed UX for work (specv2 §12).
func PrintWorkReport(w io.Writer, intent ResolvedIntent, plan ExecutionPlan, result ExecuteResult) {
	_, _ = fmt.Fprintln(w, "Asagiri resolved intent")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 25))
	// Instruction line filled by caller when available
	if intent.Reason != "" && intent.Reason != "deterministic" && intent.Reason != "ollama" {
		_, _ = fmt.Fprintf(w, "Instruction: %s\n", intent.Reason)
	}
	if result.Intent.Action != "" {
		_, _ = fmt.Fprintf(w, "Action:      %s\n", result.Intent.Action)
	}
	_, _ = fmt.Fprintf(w, "Feature:     %s\n", intent.Feature)
	if intent.TaskID != "" {
		_, _ = fmt.Fprintf(w, "Task:        %s\n", intent.TaskID)
	}
	_, _ = fmt.Fprintf(w, "Confidence:  %.2f\n", intent.Confidence)
	_, _ = fmt.Fprintln(w, "Execution plan")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 14))
	for i, step := range plan.Steps {
		_, _ = fmt.Fprintf(w, "%d. %s\n", i+1, formatPrimitive(step))
	}
	for _, line := range result.Executed {
		_, _ = fmt.Fprintf(w, "  → %s\n", line)
	}
	if result.LastRunID != "" {
		_, _ = fmt.Fprintf(w, "Last run: %s\n", result.LastRunID)
	}
}

// PrintContinueReport renders continue UX.
func PrintContinueReport(w io.Writer, feature, taskID, state, nextCmd string) {
	_, _ = fmt.Fprintln(w, "Continuing last active feature")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 30))
	_, _ = fmt.Fprintf(w, "Feature: %s\n", feature)
	if taskID != "" {
		_, _ = fmt.Fprintf(w, "Task:    %s\n", taskID)
	}
	_, _ = fmt.Fprintf(w, "State:   %s\n", state)
	_, _ = fmt.Fprintf(w, "Next:    %s\n", strings.TrimPrefix(nextCmd, "asa "))
	_, _ = fmt.Fprintln(w, "Running:")
	_, _ = fmt.Fprintln(w, nextCmd)
}

// PrintInboxTable renders inbox output.
func PrintInboxTable(w io.Writer, rows []InboxRow) {
	_, _ = fmt.Fprintln(w, "Inbox")
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 5))
	_, _ = fmt.Fprintf(w, "%-8s %-8s %-20s %s\n", "Source", "Status", "Updated", "Feature")
	for _, r := range rows {
		_, _ = fmt.Fprintf(w, "%-8s %-8s %-20s %s\n", r.Source, r.Status, r.Updated, r.Feature)
	}
}

// InboxRow is one inbox line.
type InboxRow struct {
	Source  string
	Status  string
	Updated string
	Feature string
	Path    string
}

// PrintNextRecommendation renders next output.
func PrintNextRecommendation(w io.Writer, rec NextRecommendation) {
	_, _ = fmt.Fprintf(w, "Feature: %s\n", rec.Feature)
	if rec.TaskID != "" {
		_, _ = fmt.Fprintf(w, "Task:    %s\n", rec.TaskID)
	}
	_, _ = fmt.Fprintf(w, "Next action: %s\n", rec.Action)
	_, _ = fmt.Fprintf(w, "Reason: %s\n", rec.Reason)
	_, _ = fmt.Fprintf(w, "Command: %s\n", rec.Primitive)
}
