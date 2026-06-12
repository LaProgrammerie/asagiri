package checks

import (
	"context"
	"fmt"
	"time"
)

// BlastRadiusRunner estimates change impact from analysis graphs (spec §12).
type BlastRadiusRunner struct{}

func (BlastRadiusRunner) Type() string { return typeBlastRadius }

func (BlastRadiusRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeBlastRadius, "Blast radius", "blast.radius", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeBlastRadius, "Blast radius", "blast.radius", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeBlastRadius, "Blast radius", "blast.radius", err), nil
	}

	findings := make([]Finding, 0)
	br, usedGraph := tryKnowledgeBlastRadius(ctx, scope, scope.Flow, deps)
	if !usedGraph {
		br = computeBlastRadius(pctx.flow, pctx.bundle, pctx.bundleErr)
	}

	if pctx.bundleErr != nil {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "blast.radius",
			Message:  fmt.Sprintf("analysis bundle unavailable: %v", pctx.bundleErr),
		})
	}
	if br.PublicContractRisk == "high" {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "blast.radius",
			Message:  "public contract risk is high",
		})
	}
	if br.FlowsImpacted > 3 {
		findings = append(findings, Finding{
			Severity: "info",
			Category: "blast.radius",
			Message:  fmt.Sprintf("%d flows in impact surface", br.FlowsImpacted),
		})
	}

	return finishLot3WithBlast(scope, start, typeBlastRadius, "Blast radius", findings, nil, &br), nil
}
