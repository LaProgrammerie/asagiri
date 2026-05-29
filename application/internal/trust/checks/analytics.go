package checks

import (
	"context"
	"fmt"
	"time"
)

// AnalyticsRunner validates analytics events and flow metrics (spec §8).
type AnalyticsRunner struct{}

func (AnalyticsRunner) Type() string { return typeAnalytics }

func (AnalyticsRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeAnalytics, "Analytics", "analytics.contract", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeAnalytics, "Analytics", "analytics.contract", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeAnalytics, "Analytics", "analytics.contract", err), nil
	}

	findings := make([]Finding, 0)
	if !contractExists(pctx.productDir, "events.yaml") {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "analytics.contract",
			Message:  "contracts/events.yaml missing",
		})
	} else {
		events, eErr := loadEventsContract(deps, pctx.productDir)
		if eErr != nil {
			findings = append(findings, Finding{
				Severity: "error",
				Category: "analytics.contract",
				Message:  eErr.Error(),
			})
		} else if len(events.Events) == 0 {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "analytics.contract",
				Message:  "events contract is empty",
			})
		}
	}
	if len(pctx.flow.Metrics) == 0 && pctx.flow.Business.Criticality == "high" {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "analytics.contract",
			Message:  "high criticality flow without product metrics",
		})
	}
	if pctx.flow.Outcome != "" && !contractExists(pctx.productDir, "events.yaml") {
		findings = append(findings, Finding{
			Severity: "info",
			Category: "analytics.contract",
			Message:  fmt.Sprintf("flow outcome %q should emit analytics event", pctx.flow.Outcome),
		})
	}

	return finishLot3(scope, start, typeAnalytics, "Analytics", findings, nil), nil
}
