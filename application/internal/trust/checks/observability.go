package checks

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// ObservabilityRunner validates observability contracts and flow telemetry (spec §15).
type ObservabilityRunner struct{}

func (ObservabilityRunner) Type() string { return typeObservability }

func (ObservabilityRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeObservability, "Observability", "observability.flow", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeObservability, "Observability", "observability.flow", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeObservability, "Observability", "observability.flow", err), nil
	}

	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	if !contractExists(pctx.productDir, "observability.yaml") {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "observability.contract",
			Message:  "contracts/observability.yaml missing",
		})
	} else {
		obs, oErr := loadObservabilityContract(deps, pctx.productDir)
		if oErr != nil {
			findings = append(findings, Finding{
				Severity: "error",
				Category: "observability.contract",
				Message:  oErr.Error(),
			})
		} else if len(obs.Requirements) == 0 {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "observability.contract",
				Message:  "observability contract has no requirements",
			})
		}
	}

	traceSet := make(map[string]struct{})
	for _, t := range pctx.flow.Observability.Traces {
		traceSet[strings.TrimSpace(t)] = struct{}{}
	}
	metricSet := make(map[string]struct{})
	for _, m := range pctx.flow.Observability.Metrics {
		metricSet[strings.TrimSpace(m)] = struct{}{}
	}
	for _, m := range pctx.flow.Metrics {
		metricSet[strings.TrimSpace(m)] = struct{}{}
	}

	for _, step := range pctx.flow.Steps {
		startTrace := fmt.Sprintf("%s.start", step.Action)
		if _, ok := traceSet[startTrace]; !ok {
			if _, ok := traceSet["onboarding.start"]; ok && step.Action == "click_get_started" {
				evidence = append(evidence, Evidence{Kind: "trace", Source: step.ID, Summary: "onboarding.start traced"})
				continue
			}
			if step.Sensitive || pctx.flow.Business.Criticality == "high" {
				findings = append(findings, Finding{
					Severity: "warning",
					Category: "observability.flow",
					Message:  fmt.Sprintf("%s trace missing for step %s", startTrace, step.ID),
				})
			}
		}
		failMetric := fmt.Sprintf("%s_failure", step.Action)
		if step.Sensitive && len(step.Errors) > 0 {
			if _, ok := metricSet[failMetric]; !ok {
				findings = append(findings, Finding{
					Severity: "warning",
					Category: "observability.flow",
					Message:  fmt.Sprintf("%s failure metric missing", step.Action),
				})
			}
		}
	}

	if len(pctx.flow.Observability.Traces) == 0 && len(pctx.flow.Observability.Metrics) == 0 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "observability.flow",
			Message:  "flow observability block is empty",
		})
	}
	if pctx.flow.Business.Criticality == "high" && !contractExists(pctx.productDir, "observability.yaml") {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "observability.contract",
			Message:  "no dashboard contract for high criticality flow",
		})
	}

	return finishLot3(scope, start, typeObservability, "Observability", findings, evidence), nil
}
