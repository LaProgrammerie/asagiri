package checks

import (
	"fmt"
	"math"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
)

const (
	statusPassed  = "passed"
	statusFailed  = "failed"
	statusSkipped = "skipped"
	statusWarn    = "warn"
)

func checkID(typ, trustID string) string {
	return fmt.Sprintf("%s-%s", typ, trustID)
}

// confidenceFromStatus derives check confidence from outcome only; finding penalties live in DefaultScorer.
func confidenceFromStatus(status string) float64 {
	switch status {
	case statusPassed:
		return 1.0
	case statusWarn:
		return 0.75
	case statusFailed:
		return 0
	case statusSkipped:
		return 0
	default:
		return 0
	}
}

const testsNoCandidatesCap = 0.5

func countSeverity(findings []Finding, severity string) int {
	n := 0
	for _, f := range findings {
		if f.Severity == severity {
			n++
		}
	}
	return n
}

func statusFromFindings(findings []Finding) string {
	if countSeverity(findings, "critical") > 0 || countSeverity(findings, "error") > 0 {
		return statusFailed
	}
	if countSeverity(findings, "warning") > 0 {
		return statusWarn
	}
	return statusPassed
}

func graphHasRoute(g analysis.Graph, ref string) bool {
	ref = strings.TrimSpace(ref)
	for _, n := range g.Nodes {
		if strings.Contains(n.Name, ref) || strings.Contains(n.ID, ref) {
			return true
		}
	}
	return false
}

func roundConfidence(v float64) float64 {
	return math.Round(v*1000) / 1000
}
