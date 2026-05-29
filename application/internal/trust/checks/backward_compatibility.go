package checks

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// BackwardCompatibilityRunner detects breaking contract and unresolved API refs (spec §8).
type BackwardCompatibilityRunner struct{}

func (BackwardCompatibilityRunner) Type() string { return typeBackwardCompatibility }

func (BackwardCompatibilityRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeBackwardCompatibility, "Backward compatibility", "compatibility.contract", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeBackwardCompatibility, "Backward compatibility", "compatibility.contract", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeBackwardCompatibility, "Backward compatibility", "compatibility.contract", err), nil
	}

	findings := make([]Finding, 0)
	if !contractExists(pctx.productDir, "api.openapi.yaml") {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "compatibility.contract",
			Message:  "contracts/api.openapi.yaml missing",
		})
	}
	for _, ref := range unresolvedContractRefs(pctx.flow) {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "compatibility.contract",
			Message:  fmt.Sprintf("unresolved contract_ref may break clients: %s", ref),
			SuggestedFix: "resolve OpenAPI operation or event reference",
		})
	}
	if pctx.bundleErr == nil {
		apiGraph := pctx.bundle.Graphs["api"]
		for _, step := range pctx.flow.Steps {
			ref := strings.TrimSpace(step.ContractRef)
			if ref == "" || strings.HasPrefix(ref, "TODO:") {
				continue
			}
			if !graphHasRoute(apiGraph, ref) {
				findings = append(findings, Finding{
					Severity: "warning",
					Category: "compatibility.contract",
					Message:  fmt.Sprintf("contract_ref %q not in api graph", ref),
				})
			}
		}
	}

	return finishLot3(scope, start, typeBackwardCompatibility, "Backward compatibility", findings, nil), nil
}
