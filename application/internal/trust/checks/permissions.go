package checks

import (
	"context"
	"fmt"
	"time"
)

// PermissionsRunner validates permissions contracts against flow sensitivity (spec §8).
type PermissionsRunner struct{}

func (PermissionsRunner) Type() string { return typePermissions }

func (PermissionsRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typePermissions, "Permissions", "permissions.flow", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typePermissions, "Permissions", "permissions.flow", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typePermissions, "Permissions", "permissions.flow", err), nil
	}

	findings := make([]Finding, 0)
	if !contractExists(pctx.productDir, "permissions.yaml") {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "permissions.flow",
			Message:  "contracts/permissions.yaml missing",
		})
	} else {
		doc, pErr := loadPermissionsContract(deps, pctx.productDir)
		if pErr != nil {
			findings = append(findings, Finding{
				Severity: "error",
				Category: "permissions.flow",
				Message:  pErr.Error(),
			})
		} else if len(doc.Roles) == 0 {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "permissions.flow",
				Message:  "permissions contract has no roles",
			})
		}
	}

	if flowHasSensitiveStep(pctx.flow) && !pctx.flow.Security.RequiresAuthentication {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "permissions.flow",
			Message:  "sensitive flow steps without requires_authentication",
		})
	}
	for _, step := range pctx.flow.Steps {
		if step.Sensitive && len(step.Errors) == 0 {
			findings = append(findings, Finding{
				Severity: "error",
				Category: "permissions.flow",
				Message:  fmt.Sprintf("sensitive step %s missing error handling", step.ID),
			})
		}
	}

	return finishLot3(scope, start, typePermissions, "Permissions", findings, nil), nil
}
