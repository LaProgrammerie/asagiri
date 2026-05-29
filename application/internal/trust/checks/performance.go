package checks

import (
	"context"
	"fmt"
	"time"
)

// PerformanceRunner flags performance risks from investigation and dependency density (spec §8).
type PerformanceRunner struct{}

func (PerformanceRunner) Type() string { return typePerformance }

func (PerformanceRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	inv, err := deps.Investigate(ctx, scope.RepoRoot, scope.Flow, scope.Task, deps.Config)
	if err != nil {
		return failedLot3(scope, start, typePerformance, "Performance", "performance.static", err), nil
	}
	if len(inv.LargeFiles) > 2 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "performance.static",
			Message:  fmt.Sprintf("%d large files in change scope", len(inv.LargeFiles)),
		})
	}
	if len(inv.CandidateFiles) > 0 {
		g, gErr := deps.BuildDepGraph(scope.RepoRoot, inv.CandidateFiles)
		if gErr == nil && len(g.Nodes) > 0 && len(g.Edges) > len(g.Nodes)*4 {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "performance.static",
				Message:  "dependency fan-out may impact build and test performance",
			})
			evidence = append(evidence, Evidence{
				Kind:    "graph",
				Source:  "dependency",
				Summary: fmt.Sprintf("%d nodes, %d edges", len(g.Nodes), len(g.Edges)),
			})
		}
	}

	return finishLot3(scope, start, typePerformance, "Performance", findings, evidence), nil
}
