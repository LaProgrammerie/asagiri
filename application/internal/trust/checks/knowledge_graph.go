package checks

import (
	"context"
	"errors"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

const typeKnowledgeGraph = "knowledge-graph"

// KnowledgeGraphRunner applies spec-my-E §17 trust signals from the engineering knowledge graph.
type KnowledgeGraphRunner struct{}

func (KnowledgeGraphRunner) Type() string { return typeKnowledgeGraph }

func (KnowledgeGraphRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	if scope.RepoRoot == "" || scope.Flow == "" {
		return skippedLot3(scope, start, typeKnowledgeGraph, "Knowledge graph", "graph", "no flow scope"), nil
	}
	store, err := knowledge.OpenStoreIfExists(scope.RepoRoot)
	if err != nil {
		if errors.Is(err, knowledge.ErrNotFound) {
			return skippedLot3(scope, start, typeKnowledgeGraph, "Knowledge graph", "graph", "graph.sqlite missing"), nil
		}
		return failedLot3(scope, start, typeKnowledgeGraph, "Knowledge graph", "graph", err), nil
	}
	defer store.Close()

	assess, ok, err := knowledge.AssessTrustFromGraph(ctx, store, scope.Flow, "")
	if err != nil {
		return failedLot3(scope, start, typeKnowledgeGraph, "Knowledge graph", "graph", err), nil
	}
	if !ok {
		return skippedLot3(scope, start, typeKnowledgeGraph, "Knowledge graph", "graph", "assessment unavailable"), nil
	}

	findings := make([]Finding, 0, len(assess.Findings)+4)
	evidence := []Evidence{{
		Kind:    "knowledge_graph",
		Source:  "graph.sqlite",
		Summary: "flow integrity and coverage from engineering knowledge graph",
	}}

	if assess.FlowIntegrity < 0.6 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "flow.integrity",
			Message:  "knowledge graph flow integrity below threshold",
		})
	}
	if assess.ContractCoverage < 0.5 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "contract.validation",
			Message:  "knowledge graph contract coverage incomplete",
		})
	}
	if assess.ObservabilityCoverage < 0.4 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "observability.coverage",
			Message:  "knowledge graph observability coverage low",
		})
	}
	if assess.SecurityImpact == "high" {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "security.impact",
			Message:  "knowledge graph security impact: high",
		})
	}
	for _, f := range assess.Findings {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "knowledge.graph",
			Message:  f,
		})
	}

	conf := assess.RegressionConfidence
	if assess.SecurityImpact == "high" {
		conf *= 0.85
	}
	result := finishLot3WithBlast(scope, start, typeKnowledgeGraph, "Knowledge graph", findings, evidence, &BlastRadiusSummary{
		FlowsImpacted:      1,
		PublicContractRisk: assess.SecurityImpact,
		MigrationRisk:      assess.SecurityImpact,
	})
	result.Confidence = assess.RegressionConfidence
	if result.Confidence <= 0 {
		result.Confidence = assess.FlowIntegrity
	}
	return result, nil
}
