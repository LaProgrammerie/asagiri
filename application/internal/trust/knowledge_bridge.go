package trust

import (
	"context"
	"errors"
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
)

// AssessTrustFromRepo opens the local graph and returns trust scoring inputs (spec-my-E §17).
func AssessTrustFromRepo(ctx context.Context, repoRoot, flow, action string) (knowledge.TrustGraphAssessment, bool, error) {
	if repoRoot == "" || flow == "" {
		return knowledge.TrustGraphAssessment{}, false, fmt.Errorf("trust graph: repo root and flow required")
	}
	store, err := knowledge.OpenStoreIfExists(repoRoot)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return knowledge.TrustGraphAssessment{}, false, nil
		}
		return knowledge.TrustGraphAssessment{}, false, err
	}
	defer func() { _ = store.Close() }()
	return knowledge.AssessTrustFromGraph(ctx, store, flow, action)
}

// FlowIntegrityFromGraph returns flow integrity score from the knowledge graph.
func FlowIntegrityFromGraph(ctx context.Context, repoRoot, flow, action string) (float64, bool, error) {
	assess, ok, err := AssessTrustFromRepo(ctx, repoRoot, flow, action)
	if err != nil || !ok {
		return 0, ok, err
	}
	return assess.FlowIntegrity, true, nil
}

// RegressionConfidenceFromGraph returns regression confidence from graph scope and impact.
func RegressionConfidenceFromGraph(ctx context.Context, repoRoot, flow, action string) (float64, bool, error) {
	assess, ok, err := AssessTrustFromRepo(ctx, repoRoot, flow, action)
	if err != nil || !ok {
		return 0, ok, err
	}
	return assess.RegressionConfidence, true, nil
}

// BlastRadiusFromGraph maps knowledge impact analysis to trust blast radius (spec-my-E §17).
func BlastRadiusFromGraph(ctx context.Context, repoRoot string, req knowledge.ImpactRequest) (*BlastRadiusReport, knowledge.ImpactResult, error) {
	if repoRoot == "" {
		return nil, knowledge.ImpactResult{}, fmt.Errorf("knowledge blast radius: repo root required")
	}
	store, err := knowledge.OpenStore(repoRoot)
	if err != nil {
		return nil, knowledge.ImpactResult{}, err
	}
	defer func() { _ = store.Close() }()

	result, err := knowledge.NewImpactAnalyzer(store).Analyze(ctx, req)
	if err != nil {
		return nil, knowledge.ImpactResult{}, err
	}
	summary := checks.BlastRadiusSummaryFromImpact(result, req.Flow)
	report := &BlastRadiusReport{
		FlowsImpacted:      summary.FlowsImpacted,
		CriticalAPIs:       summary.CriticalAPIs,
		SharedModules:      summary.SharedModules,
		MigrationRisk:      summary.MigrationRisk,
		PublicContractRisk: summary.PublicContractRisk,
	}
	return report, result, nil
}
