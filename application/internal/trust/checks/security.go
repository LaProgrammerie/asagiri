package checks

import (
	"context"
	"fmt"
	"time"
)

// SecurityRunner validates auth, permissions, and dangerous actions (spec §16).
type SecurityRunner struct{}

func (SecurityRunner) Type() string { return typeSecurity }

func (SecurityRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeSecurity, "Security", "security.flow", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeSecurity, "Security", "security.flow", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeSecurity, "Security", "security.flow", err), nil
	}

	findings := make([]Finding, 0)

	if pctx.flow.Business.Criticality == "high" && !pctx.flow.Security.RequiresAuthentication {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "security.flow",
			Message:  "high criticality flow without requires_authentication",
		})
	}
	for _, step := range pctx.flow.Steps {
		if step.Sensitive {
			if !pctx.flow.Security.RequiresAuthentication {
				findings = append(findings, Finding{
					Severity: "error",
					Category: "security.flow",
					Message:  fmt.Sprintf("sensitive step %s without auth requirement", step.ID),
				})
			}
			if len(step.Errors) == 0 {
				findings = append(findings, Finding{
					Severity: "error",
					Category: "security.flow",
					Message:  fmt.Sprintf("destructive step %s missing error paths", step.ID),
				})
			}
		}
	}
	for _, ref := range unresolvedContractRefs(pctx.flow) {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "security.contract",
			Message:  fmt.Sprintf("unresolved security-sensitive contract: %s", ref),
		})
	}

	inv, invErr := deps.Investigate(ctx, scope.RepoRoot, scope.Flow, scope.Task, deps.Config)
	if invErr == nil && len(inv.SensitivePaths) > 0 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "security.contract",
			Message:  fmt.Sprintf("%d sensitive paths in change scope", len(inv.SensitivePaths)),
		})
	}
	if !contractExists(pctx.productDir, "permissions.yaml") {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "security.contract",
			Message:  "permissions contract missing for security validation",
		})
	}

	return finishLot3(scope, start, typeSecurity, "Security", findings, nil), nil
}
