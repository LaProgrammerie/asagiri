package checks

import (
	"context"
	"fmt"
	"time"
)

// MigrationSafetyRunner assesses migration risk from shared dependencies (spec §8).
type MigrationSafetyRunner struct{}

func (MigrationSafetyRunner) Type() string { return typeMigrationSafety }

func (MigrationSafetyRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeMigrationSafety, "Migration safety", "migration.dependency", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeMigrationSafety, "Migration safety", "migration.dependency", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeMigrationSafety, "Migration safety", "migration.dependency", err), nil
	}

	findings := make([]Finding, 0)
	br := computeBlastRadius(pctx.flow, pctx.bundle, pctx.bundleErr)

	if pctx.bundleErr != nil {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "migration.dependency",
			Message:  fmt.Sprintf("cannot assess shared modules: %v", pctx.bundleErr),
		})
	}
	switch br.MigrationRisk {
	case "high":
		findings = append(findings, Finding{
			Severity: "error",
			Category: "migration.dependency",
			Message:  fmt.Sprintf("migration risk high (%d shared modules)", br.SharedModules),
		})
	case "medium":
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "migration.dependency",
			Message:  fmt.Sprintf("migration risk medium (%d shared modules)", br.SharedModules),
		})
	}
	if len(unresolvedContractRefs(pctx.flow)) > 0 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "migration.dependency",
			Message:  "unresolved contract refs increase migration risk",
		})
	}

	return finishLot3(scope, start, typeMigrationSafety, "Migration safety", findings, nil), nil
}
