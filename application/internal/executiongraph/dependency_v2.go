package executiongraph

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

func inferPublicEventEdges(input DependencyInput, bindings []TaskBinding, nodeSet map[string]struct{}, bundle analysis.Bundle, bundleErr error, flow product.Flow) []GraphEdge {
	if bundleErr != nil {
		return nil
	}
	eventGraph := bundle.Graphs["event"]
	if len(eventGraph.Nodes) == 0 {
		return nil
	}

	verifyID := "verify-contracts"
	trustID := "trust-gate"
	deriveID := "derive-contracts"
	_, hasVerify := nodeSet[verifyID]
	_, hasTrust := nodeSet[trustID]
	_, hasDerive := nodeSet[deriveID]

	outcomeEvent := ""
	if flowMatchesPublicEvent(flow.Outcome, eventGraph) {
		outcomeEvent = normalizeEventName(flow.Outcome)
	}

	edges := make([]GraphEdge, 0)
	backwardCompatAdded := false

	for _, b := range bindings {
		if b.NodeID == "" {
			continue
		}
		emitsPublic := bindingEmitsPublicEvent(flow.ID, b, eventGraph) || (outcomeEvent != "" && bindingEmitsFlowOutcome(b, bindings))
		if !emitsPublic {
			continue
		}

		if hasDerive && hasVerify && !backwardCompatAdded {
			edges = append(edges, GraphEdge{
				From:   deriveID,
				To:     verifyID,
				Type:   EdgeTypeValidates,
				Reason: "public event contract requires compatibility validation",
			})
			backwardCompatAdded = true
		}
		if hasVerify {
			edges = append(edges, GraphEdge{
				From:   verifyID,
				To:     b.NodeID,
				Type:   EdgeTypeRequires,
				Reason: "public event change requires backward compatibility check",
			})
		}
		if hasTrust && b.NodeID != trustID {
			edges = append(edges, GraphEdge{
				From:   b.NodeID,
				To:     trustID,
				Type:   EdgeTypeRequires,
				Reason: "public event change requires trust verification",
			})
		}
	}
	return edges
}

func inferArchitectureProjectionEdges(bindings []TaskBinding, bundle analysis.Bundle, bundleErr error) []GraphEdge {
	if bundleErr != nil || len(bindings) < 2 {
		return nil
	}
	depGraph := bundle.Graphs["dependency"]
	if len(depGraph.Edges) == 0 {
		return nil
	}

	nodeByDep := map[string]string{}
	for _, b := range bindings {
		if b.NodeID == "" {
			continue
		}
		for _, depID := range dependencyNodesForBinding(b, depGraph) {
			if prev, ok := nodeByDep[depID]; !ok || bindingPrecedes(b, bindingForNodeID(bindings, prev)) {
				nodeByDep[depID] = b.NodeID
			}
		}
	}

	edges := make([]GraphEdge, 0)
	for _, e := range depGraph.Edges {
		if !strings.EqualFold(strings.TrimSpace(e.Kind), "depends_on") {
			continue
		}
		dependentID := strings.TrimSpace(e.From)
		dependencyID := strings.TrimSpace(e.To)
		fromNode, okFrom := nodeByDep[dependentID]
		toNode, okTo := nodeByDep[dependencyID]
		if !okFrom || !okTo || fromNode == "" || toNode == "" || fromNode == toNode {
			continue
		}
		edges = append(edges, GraphEdge{
			From:   toNode,
			To:     fromNode,
			Type:   EdgeTypeRequires,
			Reason: fmt.Sprintf("architecture projection: %s depends on %s", depNodeLabel(depGraph, dependentID), depNodeLabel(depGraph, dependencyID)),
		})
	}
	return edges
}

func inferHistoricalFailureEdges(ctx context.Context, input DependencyInput, bindings []TaskBinding, nodeSet map[string]struct{}, loadRecent RecentFailuresLoader) []GraphEdge {
	if loadRecent == nil {
		loadRecent = DefaultRecentFailuresLoader{}
	}
	failures, err := loadRecent.RecentFlowFailures(ctx, input.RepoRoot, input.Product, input.Flow, recentFailureLimit)
	if err != nil || len(failures) == 0 {
		return nil
	}

	investigateID := flowInvestigationNodeID(input.Flow)
	if _, ok := nodeSet[investigateID]; !ok {
		return nil
	}

	implNodes := implementationNodeIDs(input.Nodes)
	edges := make([]GraphEdge, 0)
	for _, b := range bindings {
		if b.NodeID == "" || b.NodeID == investigateID {
			continue
		}
		if !implNodes[b.NodeID] {
			continue
		}
		edges = append(edges, GraphEdge{
			From:   investigateID,
			To:     b.NodeID,
			Type:   EdgeTypeRequires,
			Reason: "recent flow failure in runtime memory requires investigation first",
		})
	}
	return edges
}

func bindingEmitsPublicEvent(flowID string, b TaskBinding, eventGraph analysis.Graph) bool {
	flowID = strings.TrimSpace(flowID)
	if flowID == "" {
		return false
	}
	derived := normalizeEventName(flowID + "." + b.Action)
	return graphHasEvent(eventGraph, derived)
}

func bindingEmitsFlowOutcome(b TaskBinding, bindings []TaskBinding) bool {
	maxIdx := -1
	for _, other := range bindings {
		if other.StepIndex > maxIdx {
			maxIdx = other.StepIndex
		}
	}
	return b.StepIndex >= 0 && b.StepIndex == maxIdx
}

func flowMatchesPublicEvent(outcome string, eventGraph analysis.Graph) bool {
	outcome = strings.TrimSpace(outcome)
	if outcome == "" {
		return false
	}
	return graphHasEvent(eventGraph, normalizeEventName(outcome))
}

func graphHasEvent(g analysis.Graph, name string) bool {
	name = normalizeEventName(name)
	if name == "" {
		return false
	}
	for _, n := range g.Nodes {
		if n.Kind != "event" {
			continue
		}
		if normalizeEventName(n.Name) == name {
			return true
		}
		if strings.HasSuffix(normalizeEventName(strings.TrimPrefix(n.ID, "event:")), name) {
			return true
		}
	}
	return false
}

func normalizeEventName(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "_", ".")
	s = strings.TrimPrefix(s, "event:")
	return s
}

func dependencyNodesForBinding(b TaskBinding, depGraph analysis.Graph) []string {
	if len(b.ScopePaths) == 0 {
		return nil
	}
	var matched []string
	for _, n := range depGraph.Nodes {
		if scopeMatchesDepNode(b.ScopePaths, n) {
			matched = append(matched, n.ID)
		}
	}
	return matched
}

func scopeMatchesDepNode(scopes []string, n analysis.Node) bool {
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		scope = strings.TrimSuffix(scope, "/**")
		switch {
		case strings.HasPrefix(n.ID, "file:"):
			path := strings.TrimPrefix(n.ID, "file:")
			if pathMatchesScope(path, scope) {
				return true
			}
		case strings.HasPrefix(n.ID, "pkg:"):
			pkg := strings.TrimPrefix(n.ID, "pkg:")
			if strings.Contains(scope, pkg) || strings.Contains(pkg, scope) {
				return true
			}
			if strings.Contains(scope, filepath.Base(pkg)) {
				return true
			}
		case strings.HasPrefix(n.ID, "import:"):
			imp := strings.TrimPrefix(n.ID, "import:")
			if strings.Contains(scope, imp) {
				return true
			}
		default:
			if n.Name != "" && strings.Contains(scope, n.Name) {
				return true
			}
		}
	}
	return false
}

func pathMatchesScope(path, scope string) bool {
	path = filepath.ToSlash(path)
	scope = filepath.ToSlash(scope)
	return path == scope || strings.HasPrefix(path, scope+"/") || strings.HasPrefix(scope, path+"/")
}

func depNodeLabel(g analysis.Graph, id string) string {
	for _, n := range g.Nodes {
		if n.ID == id {
			if n.Name != "" {
				return n.Name
			}
			return id
		}
	}
	return id
}

func bindingForNodeID(bindings []TaskBinding, nodeID string) TaskBinding {
	for _, b := range bindings {
		if b.NodeID == nodeID {
			return b
		}
	}
	return TaskBinding{}
}

func bindingPrecedes(a, b TaskBinding) bool {
	if a.StepIndex < 0 {
		return false
	}
	if b.StepIndex < 0 {
		return true
	}
	return a.StepIndex < b.StepIndex
}

func flowInvestigationNodeID(flowID string) string {
	return "investigate-" + flowSlug(flowID)
}

func loadFlowForInference(input DependencyInput, readFile func(string) ([]byte, error)) (product.Flow, error) {
	if input.RepoRoot == "" || input.Flow == "" {
		return product.Flow{}, nil
	}
	path := resolveFlowPath(productDir(input.RepoRoot, input.Product), input.Flow)
	raw, err := readFile(path)
	if err != nil {
		return product.Flow{}, err
	}
	return product.ParseFlowYAML(raw)
}

func implementationNodeIDs(nodes []GraphNode) map[string]bool {
	out := make(map[string]bool)
	for _, n := range nodes {
		if n.Type == NodeTypeImplementation {
			out[n.ID] = true
		}
	}
	return out
}
