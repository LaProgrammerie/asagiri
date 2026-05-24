package intent

import (
	"fmt"
	"io"
	"strings"
)

// PrintWorkReport renders boxed UX for work (specv2 §12).
func PrintWorkReport(w io.Writer, intent ResolvedIntent, plan ExecutionPlan, result ExecuteResult) {
	fmt.Fprintln(w, "AgentFlow resolved intent")
	fmt.Fprintln(w, strings.Repeat("─", 25))
	// Instruction line filled by caller when available
	if intent.Reason != "" && intent.Reason != "deterministic" && intent.Reason != "ollama" {
		fmt.Fprintf(w, "Instruction: %s\n", intent.Reason)
	}
	if result.Intent.Action != "" {
		fmt.Fprintf(w, "Action:      %s\n", result.Intent.Action)
	}
	fmt.Fprintf(w, "Feature:     %s\n", intent.Feature)
	if intent.TaskID != "" {
		fmt.Fprintf(w, "Task:        %s\n", intent.TaskID)
	}
	fmt.Fprintf(w, "Confidence:  %.2f\n", intent.Confidence)
	fmt.Fprintln(w, "Execution plan")
	fmt.Fprintln(w, strings.Repeat("─", 14))
	for i, step := range plan.Steps {
		fmt.Fprintf(w, "%d. %s\n", i+1, formatPrimitive(step))
	}
	for _, line := range result.Executed {
		fmt.Fprintf(w, "  → %s\n", line)
	}
	if result.LastRunID != "" {
		fmt.Fprintf(w, "Last run: %s\n", result.LastRunID)
	}
}

// PrintContinueReport renders continue UX.
func PrintContinueReport(w io.Writer, feature, taskID, state, nextCmd string) {
	fmt.Fprintln(w, "Continuing last active feature")
	fmt.Fprintln(w, strings.Repeat("─", 30))
	fmt.Fprintf(w, "Feature: %s\n", feature)
	if taskID != "" {
		fmt.Fprintf(w, "Task:    %s\n", taskID)
	}
	fmt.Fprintf(w, "State:   %s\n", state)
	fmt.Fprintf(w, "Next:    %s\n", strings.TrimPrefix(nextCmd, "agentflow "))
	fmt.Fprintln(w, "Running:")
	fmt.Fprintln(w, nextCmd)
}

// PrintInboxTable renders inbox output.
func PrintInboxTable(w io.Writer, rows []InboxRow) {
	fmt.Fprintln(w, "Inbox")
	fmt.Fprintln(w, strings.Repeat("─", 5))
	fmt.Fprintf(w, "%-8s %-8s %-20s %s\n", "Source", "Status", "Updated", "Feature")
	for _, r := range rows {
		fmt.Fprintf(w, "%-8s %-8s %-20s %s\n", r.Source, r.Status, r.Updated, r.Feature)
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
	fmt.Fprintf(w, "Feature: %s\n", rec.Feature)
	if rec.TaskID != "" {
		fmt.Fprintf(w, "Task:    %s\n", rec.TaskID)
	}
	fmt.Fprintf(w, "Next action: %s\n", rec.Action)
	fmt.Fprintf(w, "Reason: %s\n", rec.Reason)
	fmt.Fprintf(w, "Command: %s\n", rec.Primitive)
}
