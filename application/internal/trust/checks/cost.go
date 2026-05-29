package checks

import (
	"context"
	"fmt"
	"time"
)

// CostRunner detects infra and observability cost risks (spec §17).
type CostRunner struct{}

func (CostRunner) Type() string { return typeCost }

func (CostRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeCost, "Cost", "cost.flow", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeCost, "Cost", "cost.flow", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeCost, "Cost", "cost.flow", err), nil
	}

	findings := make([]Finding, 0)
	telemetryCount := len(pctx.flow.Observability.Traces) + len(pctx.flow.Observability.Metrics) + len(pctx.flow.Observability.Logs)

	switch pctx.flow.CostProfile.InfrastructureCostRisk {
	case "high":
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "cost.flow",
			Message:  "infrastructure cost risk marked high",
		})
	case "":
		if telemetryCount > 4 || pctx.flow.Business.Criticality == "high" {
			findings = append(findings, Finding{
				Severity: "info",
				Category: "cost.flow",
				Message:  "cost profile not defined for telemetry-heavy flow",
			})
		}
	}

	if telemetryCount > 6 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "cost.flow",
			Message:  fmt.Sprintf("observability surface may inflate cost (%d signals)", telemetryCount),
		})
	}
	if pctx.flow.CostProfile.ExpectedComplexity == "high" && pctx.flow.Business.MonetizationImpact == "high" {
		findings = append(findings, Finding{
			Severity: "info",
			Category: "cost.flow",
			Message:  "high complexity with monetization impact",
		})
	}

	return finishLot3(scope, start, typeCost, "Cost", findings, nil), nil
}
