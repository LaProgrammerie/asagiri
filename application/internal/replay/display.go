package replay

import (
	"fmt"
	"io"
	"strings"
)

// FormatReplayCreate renders create command output.
func FormatReplayCreate(pkg ReplayPackage) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Asagiri Replay Engine")
	fmt.Fprintln(&b, "═════════════════════")
	fmt.Fprintf(&b, "Replay: %s\n", pkg.ID)
	fmt.Fprintf(&b, "Path:   %s\n", pkg.Path)
	if len(pkg.Manifest.Artifacts) > 0 {
		fmt.Fprintln(&b, "Artifacts")
		fmt.Fprintln(&b, "─────────")
		for _, a := range pkg.Manifest.Artifacts {
			fmt.Fprintf(&b, "✓ %s\n", a)
		}
	}
	return b.String()
}

// FormatReplayRun renders run command output (spec §26).
func FormatReplayRun(result ReplayResult) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Asagiri Replay Engine")
	fmt.Fprintln(&b, "═════════════════════")
	fmt.Fprintf(&b, "Replay: %s\n", result.ReplayID)
	if len(result.Artifacts) > 0 {
		fmt.Fprintln(&b, "Artifacts")
		fmt.Fprintln(&b, "─────────")
		for _, name := range artifactDisplayOrder(result.Artifacts) {
			mark := "✗"
			if result.Artifacts[name] {
				mark = "✓"
			}
			fmt.Fprintf(&b, "%s %s\n", mark, name)
		}
	}
	fmt.Fprintln(&b, "Replay mode")
	fmt.Fprintln(&b, "───────────")
	fmt.Fprintf(&b, "Mode:    %s\n", result.Mode)
	fmt.Fprintf(&b, "Offline: %t\n", result.Offline)
	if len(result.Warnings) > 0 {
		fmt.Fprintln(&b, "Warnings")
		fmt.Fprintln(&b, "────────")
		for _, w := range result.Warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
	}
	return b.String()
}

// FormatReplayComparison renders compare output (spec §14, §26).
func FormatReplayComparison(cmp ReplayComparison) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Replay Comparison")
	fmt.Fprintln(&b, "─────────────────")
	fmt.Fprintf(&b, "Replay A: %s\n", cmp.ReplayA)
	fmt.Fprintf(&b, "Replay B: %s\n", cmp.ReplayB)
	if cmp.CostDelta != 0 {
		fmt.Fprintf(&b, "Cost delta: %+.2f€\n", cmp.CostDelta)
	}
	if len(cmp.TrustDiff) > 0 {
		fmt.Fprintln(&b, "Trust score diff:")
		for _, line := range FormatTrustDiff(cmp.TrustDiff) {
			fmt.Fprintf(&b, "- %s\n", line)
		}
	}
	if len(cmp.Differences) > 0 {
		fmt.Fprintln(&b, "Differences:")
		for _, d := range cmp.Differences {
			fmt.Fprintf(&b, "- %s\n", d)
		}
	}
	if len(cmp.Warnings) > 0 {
		fmt.Fprintln(&b, "Warnings")
		fmt.Fprintln(&b, "────────")
		for _, w := range cmp.Warnings {
			fmt.Fprintf(&b, "- %s\n", w)
		}
	}
	return b.String()
}

// FormatReplayExplain renders explain output.
func FormatReplayExplain(cmp ReplayComparison) string {
	var b strings.Builder
	fmt.Fprintln(&b, "Replay Divergence Explanation")
	fmt.Fprintln(&b, "─────────────────────────────")
	for _, d := range cmp.Divergences {
		fmt.Fprintf(&b, "[%s] %s\n", d.Kind, d.Message)
	}
	if len(b.String()) == len("Replay Divergence Explanation\n─────────────────────────────\n") {
		fmt.Fprintln(&b, "No divergences detected.")
	}
	return b.String()
}

// WriteReplayCreate prints create UX.
func WriteReplayCreate(out io.Writer, pkg ReplayPackage) {
	fmt.Fprint(out, FormatReplayCreate(pkg))
}

// WriteReplayRun prints run UX.
func WriteReplayRun(out io.Writer, result ReplayResult) {
	fmt.Fprint(out, FormatReplayRun(result))
}

// WriteReplayComparison prints compare UX.
func WriteReplayComparison(out io.Writer, cmp ReplayComparison) {
	fmt.Fprint(out, FormatReplayComparison(cmp))
}

// WriteReplayExplain prints explain UX.
func WriteReplayExplain(out io.Writer, cmp ReplayComparison) {
	fmt.Fprint(out, FormatReplayExplain(cmp))
}

func artifactDisplayOrder(m map[string]bool) []string {
	order := []string{"execution_graph", "trust_report", "investigation_report", "runtime_events", "handoffs"}
	var out []string
	seen := map[string]struct{}{}
	for _, k := range order {
		if _, ok := m[k]; ok {
			out = append(out, k)
			seen[k] = struct{}{}
		}
	}
	for k := range m {
		if _, ok := seen[k]; !ok {
			out = append(out, k)
		}
	}
	return out
}
