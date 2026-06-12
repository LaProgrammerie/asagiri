package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"gopkg.in/yaml.v3"
)

const typeContracts = "contracts"

// ContractsRunner validates product contracts against flow refs and analysis graphs.
type ContractsRunner struct{}

func (ContractsRunner) Type() string { return typeContracts }

func (ContractsRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	if scope.ProductID == "" {
		return skippedContracts(scope, start, "no product resolved"), nil
	}

	productDir := ProductDir(scope.RepoRoot, scope.ProductID)
	openAPIPath := filepath.Join(productDir, "contracts", "api.openapi.yaml")
	if _, err := os.Stat(openAPIPath); err != nil {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "contract.openapi",
			Message:  "contracts/api.openapi.yaml missing",
		})
	}

	bundle, bundleErr := deps.LoadBundle(scope.RepoRoot, scope.ProductID)
	apiGraph := bundle.Graphs["api"]
	if bundleErr != nil {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "architecture.contract",
			Message:  fmt.Sprintf("analysis bundle unavailable: %v", bundleErr),
		})
	}

	flowPath := resolveFlowPath(productDir, scope.Flow)
	flowRaw, err := deps.ReadFile(flowPath)
	if err != nil {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "flow.integrity",
			Message:  fmt.Sprintf("flow file not found: %s", scope.Flow),
		})
	} else {
		var flow product.Flow
		if err := yaml.Unmarshal(flowRaw, &flow); err != nil {
			findings = append(findings, Finding{
				Severity: "error",
				Category: "flow.integrity",
				Message:  fmt.Sprintf("parse flow: %v", err),
			})
		} else {
			for _, step := range flow.Steps {
				ref := strings.TrimSpace(step.ContractRef)
				if ref == "" {
					continue
				}
				if strings.HasPrefix(ref, "TODO:") {
					findings = append(findings, Finding{
						Severity:     "warning",
						Category:     "architecture.contract",
						Message:      fmt.Sprintf("unresolved contract_ref on step %s: %s", step.ID, ref),
						SuggestedFix: "resolve contract reference in OpenAPI or events",
					})
					continue
				}
				if bundleErr == nil && !graphHasRoute(apiGraph, ref) {
					findings = append(findings, Finding{
						Severity: "warning",
						Category: "contract.openapi",
						Message:  fmt.Sprintf("contract_ref %q not found in api graph", ref),
					})
				}
			}
		}
	}

	if data, err := deps.ReadFile(openAPIPath); err == nil {
		evidence = append(evidence, Evidence{Kind: "file", Source: openAPIPath, Summary: fmt.Sprintf("%d bytes", len(data))})
	}

	status := statusFromFindings(findings)
	return CheckResult{
		ID:         checkID(typeContracts, scope.TrustID),
		Name:       "Contracts",
		Type:       typeContracts,
		Status:     status,
		Confidence: roundConfidence(confidenceFromStatus(status)),
		Findings:   findings,
		Evidence:   evidence,
		Duration:   time.Since(start),
	}, nil
}

func skippedContracts(scope Scope, start time.Time, msg string) CheckResult {
	return CheckResult{
		ID:         checkID(typeContracts, scope.TrustID),
		Name:       "Contracts",
		Type:       typeContracts,
		Status:     statusSkipped,
		Confidence: 0,
		Findings: []Finding{{
			Severity: "info",
			Category: "architecture.contract",
			Message:  msg,
		}},
		Duration: time.Since(start),
	}
}

func resolveFlowPath(productDir, flowID string) string {
	if flowID == "" {
		return ""
	}
	candidates := []string{
		filepath.Join(productDir, "flows", flowID+".flow.yaml"),
		filepath.Join(productDir, "flows", flowID),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return candidates[0]
}
