package checks

import (
	"context"
	"fmt"
	"time"
)

// ArchitectureRunner validates architecture implications and analysis graphs (spec §8).
type ArchitectureRunner struct{}

func (ArchitectureRunner) Type() string { return typeArchitecture }

func (ArchitectureRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.ProductID == "" {
		return skippedLot3(scope, start, typeArchitecture, "Architecture", "architecture.flow", "no product resolved"), nil
	}
	pctx, skipped, err := loadProductContext(scope, deps)
	if skipped {
		return skippedLot3(scope, start, typeArchitecture, "Architecture", "architecture.flow", "no product resolved"), nil
	}
	if err != nil {
		return failedLot3(scope, start, typeArchitecture, "Architecture", "architecture.flow", err), nil
	}

	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	if len(pctx.flow.ArchitectureImplications) == 0 {
		findings = append(findings, Finding{
			Severity: "info",
			Category: "architecture.flow",
			Message:  "no architecture_implications documented on flow",
		})
	}
	if pctx.bundleErr != nil {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "architecture.flow",
			Message:  fmt.Sprintf("analysis bundle unavailable: %v", pctx.bundleErr),
		})
	} else {
		if _, ok := pctx.bundle.Graphs["api"]; !ok {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "architecture.flow",
				Message:  "api graph missing from analysis bundle",
			})
		}
		if _, ok := pctx.bundle.Graphs["dependency"]; !ok {
			findings = append(findings, Finding{
				Severity: "warning",
				Category: "architecture.flow",
				Message:  "dependency graph missing from analysis bundle",
			})
		}
		for name, g := range pctx.bundle.Graphs {
			evidence = append(evidence, Evidence{
				Kind:    "graph",
				Source:  name,
				Summary: fmt.Sprintf("%d nodes", len(g.Nodes)),
			})
		}
	}

	return finishLot3(scope, start, typeArchitecture, "Architecture", findings, evidence), nil
}
