package checks

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

const typeFlows = "flows"

// FlowsRunner validates flow YAML integrity and review-style gaps.
type FlowsRunner struct{}

func (FlowsRunner) Type() string { return typeFlows }

func (FlowsRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	if scope.ProductID == "" {
		return skippedFlows(scope, start, "no product resolved"), nil
	}

	flowPath := resolveFlowPath(ProductDir(scope.RepoRoot, scope.ProductID), scope.Flow)
	raw, err := deps.ReadFile(flowPath)
	if err != nil {
		return CheckResult{
			ID:         checkID(typeFlows, scope.TrustID),
			Name:       "Flow integrity",
			Type:       typeFlows,
			Status:     statusFailed,
			Confidence: 0,
			Findings: []Finding{{
				Severity: "error",
				Category: "flow.integrity",
				Message:  fmt.Sprintf("read flow: %v", err),
			}},
			Duration: time.Since(start),
		}, nil
	}

	flow, parseErr := product.ParseFlowYAML(raw)
	if parseErr != nil {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "flow.integrity",
			Message:  parseErr.Error(),
		})
	} else {
		if bundle, bErr := deps.LoadBundle(scope.RepoRoot, scope.ProductID); bErr == nil {
			if fg, ok := bundle.Graphs["flow"]; ok {
				evidence = append(evidence, Evidence{
					Kind:    "graph",
					Source:  "flow",
					Summary: fmt.Sprintf("%d flow graph nodes", len(fg.Nodes)),
				})
			}
		}
		for _, step := range flow.Steps {
			if step.Sensitive && len(step.Errors) == 0 {
				findings = append(findings, Finding{
					Severity: "error",
					Category: "flow.security",
					Message:  fmt.Sprintf("step %s sensitive without errors", step.ID),
				})
			}
			if strings.TrimSpace(step.ContractRef) == "" {
				findings = append(findings, Finding{
					Severity: "warning",
					Category: "flow.integrity",
					Message:  fmt.Sprintf("step %s missing contract_ref", step.ID),
				})
			}
		}
		if flow.Business.Criticality == "high" && len(flow.Metrics) == 0 {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "flow.observability",
				Message:  "high criticality flow without metrics",
			})
		}
		if len(flow.Observability.Metrics) == 0 && len(flow.Observability.Traces) == 0 {
			findings = append(findings, Finding{
				Severity: "info",
				Category: "flow.observability",
				Message:  "flow observability block is empty",
			})
		}
	}

	status := statusFromFindings(findings)
	return CheckResult{
		ID:         checkID(typeFlows, scope.TrustID),
		Name:       "Flow integrity",
		Type:       typeFlows,
		Status:     status,
		Confidence: roundConfidence(confidenceFromStatus(status)),
		Findings:   findings,
		Evidence:   evidence,
		Duration:   time.Since(start),
	}, nil
}

func skippedFlows(scope Scope, start time.Time, msg string) CheckResult {
	return CheckResult{
		ID:         checkID(typeFlows, scope.TrustID),
		Name:       "Flow integrity",
		Type:       typeFlows,
		Status:     statusSkipped,
		Confidence: 0,
		Findings: []Finding{{
			Severity: "info",
			Category: "flow.integrity",
			Message:  msg,
		}},
		Duration: time.Since(start),
	}
}
