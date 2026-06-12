package reportdiff

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

func FormatTrustTaskText(w io.Writer, d TrustTaskDiff) {
	writeHeader(w, "task", d.ScopeID, d.Paths)
	_, _ = fmt.Fprintf(w, "Score:   %.1f → %.1f (%+.1f)\n", d.Score.Before, d.Score.After, d.Score.Delta)
	writeVerdict(w, d.Verdict)
	writeDimensions(w, d.Dimensions)
	writeNextAction(w, d.NextAction)
}

func FormatTrustFeatureText(w io.Writer, d TrustFeatureDiff) {
	writeHeader(w, "feature", d.ScopeID, d.Paths)
	_, _ = fmt.Fprintf(w, "Score:   %.1f → %.1f (%+.1f)\n", d.Score.Before, d.Score.After, d.Score.Delta)
	if d.TaskCount != d.TaskCountAfter {
		_, _ = fmt.Fprintf(w, "Tasks:   %d → %d\n", d.TaskCount, d.TaskCountAfter)
	}
	writeVerdict(w, d.Verdict)
	writeNextAction(w, d.NextAction)
}

func FormatTrustRunText(w io.Writer, d TrustRunDiff) {
	writeHeader(w, "run", d.ScopeID, d.Paths)
	_, _ = fmt.Fprintf(w, "Score:   %.1f → %.1f (%+.1f)\n", d.Score.Before, d.Score.After, d.Score.Delta)
	writeVerdict(w, d.Verdict)
	writeNextAction(w, d.NextAction)
}

func FormatDoctorText(w io.Writer, d DoctorDiff) {
	writeHeader(w, "doctor", "", d.Paths)
	_, _ = fmt.Fprintf(w, "Ready:    %t → %t\n", d.Ready.Before, d.Ready.After)
	_, _ = fmt.Fprintf(w, "Warnings: %d → %d (%+d)\n", d.Warnings.Before, d.Warnings.After, d.Warnings.Delta)
	_, _ = fmt.Fprintf(w, "Failures: %d → %d (%+d)\n", d.Failures.Before, d.Failures.After, d.Failures.Delta)
	if d.TrustVerdict.Before != "" || d.TrustVerdict.After != "" {
		writeVerdict(w, d.TrustVerdict)
	}
	writeNextAction(w, d.NextAction)
}

func FormatJSON(w io.Writer, v any) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(v)
}

func writeHeader(w io.Writer, scope, id string, paths ReportPaths) {
	if id != "" {
		_, _ = fmt.Fprintf(w, "Report diff — %s %s\n", scope, id)
	} else {
		_, _ = fmt.Fprintf(w, "Report diff — %s\n", scope)
	}
	_, _ = fmt.Fprintf(w, "Before: %s\n", paths.Before)
	_, _ = fmt.Fprintf(w, "After:  %s\n", paths.After)
	_, _ = fmt.Fprintln(w, strings.Repeat("─", 24))
}

func writeVerdict(w io.Writer, v VerdictDelta) {
	if !v.Changed {
		_, _ = fmt.Fprintf(w, "Verdict: %s (inchangé)\n", v.After)
		return
	}
	_, _ = fmt.Fprintf(w, "Verdict: %s → %s\n", v.Before, v.After)
}

func writeNextAction(w io.Writer, n NextActionDelta) {
	if !n.Changed {
		if n.AfterCommand != "" {
			_, _ = fmt.Fprintf(w, "Next:    %s (inchangé)\n", n.AfterCommand)
		}
		return
	}
	_, _ = fmt.Fprintf(w, "Next:    %s → %s\n", displayCmd(n.BeforeCommand), displayCmd(n.AfterCommand))
}

func writeDimensions(w io.Writer, dims []DimensionDelta) {
	changed := 0
	for _, d := range dims {
		if d.Changed {
			changed++
		}
	}
	if changed == 0 {
		_, _ = fmt.Fprintln(w, "Dimensions: inchangées")
		return
	}
	_, _ = fmt.Fprintln(w, "Dimensions:")
	for _, d := range dims {
		if !d.Changed {
			continue
		}
		label := d.ID
		if d.Label != "" {
			label = d.Label
		}
		_, _ = fmt.Fprintf(w, "  - %s: %.1f → %.1f (%+.1f)\n", label, d.Before, d.After, d.Delta)
	}
}

func displayCmd(cmd string) string {
	cmd = strings.TrimSpace(cmd)
	if cmd == "" {
		return "∅"
	}
	return cmd
}
