package knowledge

import (
	"context"
	"strings"
)

// TrustGraphAssessment aggregates trust signals from the engineering knowledge graph (spec-my-E §17).
type TrustGraphAssessment struct {
	FlowIntegrity         float64  `json:"flow_integrity"`
	RegressionConfidence  float64  `json:"regression_confidence"`
	ContractCoverage      float64  `json:"contract_coverage"`
	ObservabilityCoverage float64  `json:"observability_coverage"`
	SecurityImpact        string   `json:"security_impact"`
	Findings              []string `json:"findings,omitempty"`
}

// AssessTrustFromGraph derives trust scoring inputs for a flow/action scope.
func AssessTrustFromGraph(ctx context.Context, store GraphStore, flow, action string) (TrustGraphAssessment, bool, error) {
	if store == nil || strings.TrimSpace(flow) == "" {
		return TrustGraphAssessment{}, false, nil
	}
	scope, err := ResolveFlowScope(ctx, store, FlowScopeRequest{Flow: flow, Action: action})
	if err != nil {
		return TrustGraphAssessment{}, false, err
	}
	impact, err := NewImpactAnalyzer(store).Analyze(ctx, ImpactRequest{Flow: flow, Action: action})
	if err != nil {
		return TrustGraphAssessment{}, false, err
	}

	assess := TrustGraphAssessment{
		SecurityImpact: securityImpactLevel(scope, impact),
	}
	assess.FlowIntegrity = flowIntegrityScore(scope)
	assess.ContractCoverage = contractCoverageScore(ctx, store, scope)
	assess.ObservabilityCoverage = observabilityCoverageScore(scope)
	assess.RegressionConfidence = regressionConfidenceScore(scope, impact)
	assess.Findings = trustFindings(scope, impact)

	return assess, true, nil
}

func flowIntegrityScore(scope FlowScopeResult) float64 {
	if len(scope.Flows) == 0 {
		return 0.4
	}
	score := 0.55
	if len(scope.APIs) > 0 {
		score += 0.2
	}
	if len(scope.Events) > 0 {
		score += 0.1
	}
	if len(scope.Files) > 0 {
		score += 0.1
	}
	if len(scope.Tests) > 0 || len(scope.TestFiles) > 0 {
		score += 0.05
	}
	if score > 1 {
		return 1
	}
	return score
}

func contractCoverageScore(ctx context.Context, store GraphStore, scope FlowScopeResult) float64 {
	if len(scope.APIs) == 0 {
		return 0.5
	}
	linked := 0
	apiNodes, err := store.ListNodes(ctx, NodeFilter{Type: NodeTypeAPIOperation})
	if err != nil {
		return 0
	}
	for _, api := range scope.APIs {
		var nodeID string
		for _, n := range apiNodes {
			if n.Name == api || strings.Contains(n.Name, api) {
				nodeID = n.ID
				break
			}
		}
		if nodeID == "" {
			continue
		}
		edges, err := store.ListEdges(ctx, EdgeFilter{FromNodeID: nodeID})
		if err != nil {
			continue
		}
		for _, e := range edges {
			if e.Type == EdgeTypeRequires || e.Type == EdgeTypeValidates {
				linked++
				break
			}
		}
	}
	return float64(linked) / float64(len(scope.APIs))
}

func observabilityCoverageScore(scope FlowScopeResult) float64 {
	if len(scope.APIs) == 0 {
		return 0.5
	}
	if len(scope.Metrics) == 0 {
		return 0.35
	}
	ratio := float64(len(scope.Metrics)) / float64(len(scope.APIs))
	if ratio > 1 {
		return 1
	}
	return 0.4 + 0.6*ratio
}

func regressionConfidenceScore(scope FlowScopeResult, impact ImpactResult) float64 {
	score := 0.45
	if len(scope.Tests) > 0 || len(scope.TestFiles) > 0 {
		score += 0.25
	}
	if len(impact.ImpactedTests) > 0 {
		score += 0.15
	}
	switch impact.Risk {
	case "high":
		score -= 0.15
	case "medium":
		score -= 0.05
	default:
		score += 0.05
	}
	if score < 0 {
		return 0
	}
	if score > 1 {
		return 1
	}
	return score
}

func securityImpactLevel(scope FlowScopeResult, impact ImpactResult) string {
	criticalAPIs := 0
	for _, api := range scope.APIs {
		u := strings.ToUpper(api)
		if strings.HasPrefix(u, "POST") || strings.HasPrefix(u, "PUT") ||
			strings.HasPrefix(u, "PATCH") || strings.HasPrefix(u, "DELETE") {
			criticalAPIs++
		}
	}
	if criticalAPIs >= 2 || impact.Risk == "high" {
		return "high"
	}
	if criticalAPIs == 1 || impact.Risk == "medium" {
		return "medium"
	}
	return "low"
}

func trustFindings(scope FlowScopeResult, impact ImpactResult) []string {
	var out []string
	if len(scope.APIs) > 0 && len(scope.Tests) == 0 && len(scope.TestFiles) == 0 {
		out = append(out, "knowledge graph: APIs without linked tests in flow scope")
	}
	if len(scope.APIs) > 0 && len(scope.Metrics) == 0 {
		out = append(out, "knowledge graph: APIs without observability metrics")
	}
	if len(impact.ImpactedFlows) > 2 {
		out = append(out, "knowledge graph: wide blast radius across flows")
	}
	return out
}
