package replay

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Comparator compares replay packages (spec §14).
type Comparator struct{}

// DefaultComparator returns the standard comparator.
func DefaultComparator() *Comparator {
	return &Comparator{}
}

// Compare loads two replay packages and returns a structured comparison.
func (c *Comparator) Compare(ctx context.Context, repoRoot, a, b string) (ReplayComparison, error) {
	_ = ctx
	if err := ValidateReplayID(a); err != nil {
		return ReplayComparison{}, err
	}
	if err := ValidateReplayID(b); err != nil {
		return ReplayComparison{}, err
	}

	pkgA, err := LoadPackage(repoRoot, a)
	if err != nil {
		return ReplayComparison{}, err
	}
	pkgB, err := LoadPackage(repoRoot, b)
	if err != nil {
		return ReplayComparison{}, err
	}

	divs, err := DetectDivergences(repoRoot, a, b)
	if err != nil {
		return ReplayComparison{}, err
	}

	trustA := trustScoresFromDir(pkgA.Path)
	trustB := trustScoresFromDir(pkgB.Path)
	trustDiff := map[string]float64{}
	for dim, scoreA := range trustA {
		if scoreB, ok := trustB[dim]; ok {
			trustDiff[dim] = scoreB - scoreA
		}
	}

	cmp := ReplayComparison{
		ReplayA:     a,
		ReplayB:     b,
		CostDelta:   metricsCost(pkgB.Path) - metricsCost(pkgA.Path),
		TrustDiff:   trustDiff,
		Divergences: divs,
		Differences: ExplainDivergences(divs),
		Warnings:    comparisonWarnings(pkgA, pkgB),
	}
	return cmp, nil
}

func metricsCost(replayDir string) float64 {
	path := filepath.Join(replayDir, "graph", "metrics.json")
	body, err := os.ReadFile(path)
	if err != nil {
		body, err = ReadMaybeCompressed(path)
		if err != nil {
			return 0
		}
	}
	var metrics struct {
		EstimatedCost float64 `json:"estimated_cost"`
		Cost          float64 `json:"cost"`
	}
	if err := json.Unmarshal(body, &metrics); err != nil {
		return 0
	}
	if metrics.Cost > 0 {
		return metrics.Cost
	}
	return metrics.EstimatedCost
}

func comparisonWarnings(a, b ReplayPackage) []string {
	var warnings []string
	if _, err := os.Stat(filepath.Join(a.Path, "graph", "metrics.json")); os.IsNotExist(err) {
		warnings = append(warnings, "replay package missing dashboard metrics")
	}
	if a.Manifest.Runtime.AsagiriVersion != b.Manifest.Runtime.AsagiriVersion && b.Manifest.Runtime.AsagiriVersion != "" {
		warnings = append(warnings, "runtime version mismatch")
	}
	return warnings
}

// FormatTrustDiff returns formatted trust delta lines for display.
func FormatTrustDiff(diff map[string]float64) []string {
	if len(diff) == 0 {
		return nil
	}
	lines := make([]string, 0, len(diff))
	for dim, delta := range diff {
		lines = append(lines, fmt.Sprintf("%s: %+.2f", dim, delta))
	}
	return lines
}
