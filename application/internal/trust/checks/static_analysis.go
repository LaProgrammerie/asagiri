package checks

import (
	"context"
	"fmt"
	"time"
)

const typeStaticAnalysis = "static-analysis"

// StaticAnalysisRunner runs grep/symbol/dependency investigation (spec §8).
type StaticAnalysisRunner struct{}

func (StaticAnalysisRunner) Type() string { return typeStaticAnalysis }

func (StaticAnalysisRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	inv, err := deps.Investigate(ctx, scope.RepoRoot, scope.Flow, scope.Task, deps.Config)
	if err != nil {
		return CheckResult{
			ID:         checkID(typeStaticAnalysis, scope.TrustID),
			Name:       "Static analysis",
			Type:       typeStaticAnalysis,
			Status:     statusFailed,
			Confidence: 0,
			Findings: []Finding{{
				Severity: "error",
				Category: "implementation.static",
				Message:  fmt.Sprintf("investigation failed: %v", err),
			}},
			Duration: time.Since(start),
		}, nil
	}

	for _, e := range inv.Errors {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "implementation.static",
			Message:  e,
		})
	}
	if len(inv.LargeFiles) > 3 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "implementation.static",
			Message:  fmt.Sprintf("%d large files in scope", len(inv.LargeFiles)),
		})
	}
	if len(inv.SensitivePaths) > 0 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "flow.security",
			Message:  fmt.Sprintf("%d sensitive paths detected", len(inv.SensitivePaths)),
		})
	}

	if len(inv.CandidateFiles) > 0 {
		g, gErr := deps.BuildDepGraph(scope.RepoRoot, inv.CandidateFiles)
		if gErr == nil && len(g.Edges) > len(g.Nodes)*3 && len(g.Nodes) > 0 {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "architecture.dependency",
				Message:  "dependency graph density is high",
			})
		}
		if gErr == nil {
			evidence = append(evidence, Evidence{
				Kind:    "graph",
				Source:  "dependency",
				Summary: fmt.Sprintf("%d nodes, %d edges", len(g.Nodes), len(g.Edges)),
			})
		}
	}

	evidence = append(evidence, Evidence{
		Kind:    "investigation",
		Source:  scope.Flow,
		Summary: fmt.Sprintf("%d candidates, %d related tests", len(inv.CandidateFiles), len(inv.RelatedTests)),
	})

	status := statusFromFindings(findings)
	if status == statusPassed && len(inv.Errors) > 0 {
		status = statusFailed
	}
	conf := confidenceFromStatus(status)

	return CheckResult{
		ID:         checkID(typeStaticAnalysis, scope.TrustID),
		Name:       "Static analysis",
		Type:       typeStaticAnalysis,
		Status:     status,
		Confidence: roundConfidence(conf),
		Findings:   findings,
		Evidence:   evidence,
		Duration:   time.Since(start),
	}, nil
}
