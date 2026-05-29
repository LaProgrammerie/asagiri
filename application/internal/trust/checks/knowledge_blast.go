package checks

import (
	"context"
	"errors"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

func tryKnowledgeBlastRadius(ctx context.Context, scope Scope, flowID string, deps Dependencies) (BlastRadiusSummary, bool) {
	if deps.KnowledgeBlastRadius != nil {
		return deps.KnowledgeBlastRadius(ctx, scope.RepoRoot, flowID)
	}
	if scope.RepoRoot == "" {
		return BlastRadiusSummary{}, false
	}
	store, err := knowledge.OpenStoreIfExists(scope.RepoRoot)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return BlastRadiusSummary{}, false
		}
		return BlastRadiusSummary{}, false
	}
	defer store.Close()

	result, err := knowledge.NewImpactAnalyzer(store).Analyze(ctx, knowledge.ImpactRequest{Flow: flowID})
	if err != nil {
		return BlastRadiusSummary{}, false
	}
	return BlastRadiusSummaryFromImpact(result, flowID), true
}

// BlastRadiusSummaryFromImpact maps knowledge impact to a blast-radius summary (spec-my-E §17).
func BlastRadiusSummaryFromImpact(result knowledge.ImpactResult, flowID string) BlastRadiusSummary {
	summary := BlastRadiusSummary{
		FlowsImpacted:      len(result.ImpactedFlows),
		CriticalAPIs:       countCriticalAPIs(result.ImpactedAPIs),
		SharedModules:      len(result.ImpactedTests),
		MigrationRisk:      migrationRiskFromImpact(result.Risk),
		PublicContractRisk: publicContractRiskFromImpact(result),
	}
	if summary.FlowsImpacted == 0 && flowID != "" {
		summary.FlowsImpacted = 1
	}
	return summary
}

func countCriticalAPIs(apis []string) int {
	n := 0
	for _, api := range apis {
		u := strings.ToUpper(strings.TrimSpace(api))
		if strings.HasPrefix(u, "POST") || strings.HasPrefix(u, "PUT") ||
			strings.HasPrefix(u, "PATCH") || strings.HasPrefix(u, "DELETE") {
			n++
		}
	}
	if n == 0 && len(apis) > 0 {
		return len(apis)
	}
	return n
}

func migrationRiskFromImpact(risk string) string {
	switch risk {
	case "high":
		return "high"
	case "medium":
		return "medium"
	default:
		return "low"
	}
}

func publicContractRiskFromImpact(result knowledge.ImpactResult) string {
	if len(result.ImpactedAPIs) >= 2 {
		return "high"
	}
	if len(result.ImpactedAPIs) == 1 {
		return "medium"
	}
	return "low"
}
