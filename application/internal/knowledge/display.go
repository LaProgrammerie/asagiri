package knowledge

import (
	"fmt"
	"io"
	"strings"
)

// FormatStaleness renders the §21 terminal template.
func FormatStaleness(report StalenessReport) string {
	if !report.Stale {
		return ""
	}
	var b strings.Builder
	_, _ = fmt.Fprintln(&b, "Knowledge graph stale")
	_, _ = fmt.Fprintln(&b, "─────────────────────")
	_, _ = fmt.Fprintf(&b, "%d files changed since last build\n", report.FilesChanged)
	if report.EdgesOutdated > 0 {
		_, _ = fmt.Fprintf(&b, "%d edges may be outdated\n", report.EdgesOutdated)
	}
	if report.RecommendCommand != "" {
		_, _ = fmt.Fprintf(&b, "Run: %s\n", report.RecommendCommand)
	}
	return b.String()
}

// FormatKnowledgeBuild renders the §23 terminal template.
func FormatKnowledgeBuild(result BuildResult) string {
	var b strings.Builder
	_, _ = fmt.Fprintln(&b, "Asagiri Knowledge Graph")
	_, _ = fmt.Fprintln(&b, "═══════════════════════")
	_, _ = fmt.Fprintln(&b, "Build complete")
	_, _ = fmt.Fprintf(&b, "Nodes:        %d\n", result.Nodes)
	_, _ = fmt.Fprintf(&b, "Edges:        %d\n", result.Edges)
	if len(result.Sources) > 0 {
		_, _ = fmt.Fprintf(&b, "Sources:      %s\n", strings.Join(result.Sources, ", "))
	}
	if result.AvgConfidence > 0 {
		_, _ = fmt.Fprintf(&b, "Confidence:   %.2f avg\n", result.AvgConfidence)
	}
	stale := result.StaleFiles
	_, _ = fmt.Fprintf(&b, "Stale:        %d\n", stale)
	if result.Rebuilt {
		_, _ = fmt.Fprintln(&b, "Mode:         full rebuild (no prior metadata)")
	} else if len(result.SkippedExtractors) > 0 {
		_, _ = fmt.Fprintf(&b, "Skipped:      %s (unchanged)\n", strings.Join(result.SkippedExtractors, ", "))
	}
	if len(result.Warnings) > 0 {
		_, _ = fmt.Fprintln(&b, "Top warnings")
		_, _ = fmt.Fprintln(&b, "────────────")
		limit := len(result.Warnings)
		if limit > 5 {
			limit = 5
		}
		for i := 0; i < limit; i++ {
			_, _ = fmt.Fprintf(&b, "- %s\n", result.Warnings[i])
		}
	}
	return b.String()
}

// WriteKnowledgeBuild prints build UX to out.
func WriteKnowledgeBuild(out io.Writer, result BuildResult, staleness StalenessReport) {
	if staleness.Stale {
		_, _ = fmt.Fprint(out, FormatStaleness(staleness))
	}
	_, _ = fmt.Fprint(out, FormatKnowledgeBuild(result))
}
