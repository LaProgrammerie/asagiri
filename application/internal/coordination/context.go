package coordination

import (
	"context"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

// ContextPack is the minimal context delivered to an agent (spec-my-D §12).
type ContextPack struct {
	NodeID               string   `json:"node_id"`
	Product              string   `json:"product,omitempty"`
	Flow                 string   `json:"flow,omitempty"`
	Files                []string `json:"files,omitempty"`
	TrustConstraints     []string `json:"trust_constraints,omitempty"`
	InvestigationOutputs []string `json:"investigation_outputs,omitempty"`
}

// HandoffHints tune context reduction for a node.
type HandoffHints struct {
	AllowedFiles   []string
	MaxFiles       int
	KnowledgeStore knowledge.GraphStore
	GraphFlow      string
	GraphAction    string
}

// ReduceContext filters node outputs and injects graph metadata (spec-my-D §12).
// Returns the pack and approximate byte sizes before and after reduction.
func ReduceContext(node executiongraph.GraphNode, graph ExecutionGraph, hints HandoffHints) (ContextPack, int, int) {
	before := packByteEstimate(node.Outputs, graph.Product, graph.Flow, nil, nil)

	files := filterFiles(node.Outputs, hints)
	investigation := collectInvestigationOutputs(graph, node.ID)
	trust := defaultTrustConstraints(node)

	pack := ContextPack{
		NodeID:               node.ID,
		Product:              graph.Product,
		Flow:                 graph.Flow,
		Files:                files,
		TrustConstraints:     trust,
		InvestigationOutputs: investigation,
	}
	flow := hints.GraphFlow
	if flow == "" {
		flow = graph.Flow
	}
	if hints.KnowledgeStore != nil && flow != "" {
		enriched, err := EnrichContextFromGraph(context.Background(), pack, hints.KnowledgeStore, flow, hints.GraphAction)
		if err == nil {
			pack = enriched
		}
	}
	after := packByteEstimate(pack.Files, pack.Product, pack.Flow, pack.TrustConstraints, pack.InvestigationOutputs)
	return pack, before, after
}

// ReduceContextAndEmit reduces context and records agent.context_reduced when emitter is set.
func ReduceContextAndEmit(
	emitter *CoordinationEmitter,
	graphID, flowID string,
	node executiongraph.GraphNode,
	graph ExecutionGraph,
	hints HandoffHints,
) (ContextPack, error) {
	pack, before, after := ReduceContext(node, graph, hints)
	if emitter != nil {
		_ = emitter.EmitContextReduced(graphID, flowID, node.ID, before, after)
	}
	return pack, nil
}

func filterFiles(outputs []string, hints HandoffHints) []string {
	allowed := make(map[string]struct{}, len(hints.AllowedFiles))
	for _, f := range hints.AllowedFiles {
		f = strings.TrimSpace(f)
		if f != "" {
			allowed[f] = struct{}{}
		}
	}

	var files []string
	for _, p := range outputs {
		p = strings.TrimSpace(p)
		if p == "" || strings.Contains(p, "..") {
			continue
		}
		if len(allowed) > 0 {
			if _, ok := allowed[p]; !ok {
				continue
			}
		}
		files = append(files, p)
	}
	sort.Strings(files)

	max := hints.MaxFiles
	if max <= 0 {
		max = 32
	}
	if len(files) > max {
		files = files[:max]
	}
	return files
}

func collectInvestigationOutputs(graph ExecutionGraph, skipNodeID string) []string {
	var out []string
	for _, n := range graph.Nodes {
		if n.Type != executiongraph.NodeTypeInvestigation {
			continue
		}
		if n.ID == skipNodeID {
			continue
		}
		out = append(out, n.Outputs...)
	}
	sort.Strings(out)
	return out
}

func defaultTrustConstraints(node executiongraph.GraphNode) []string {
	switch node.Type {
	case executiongraph.NodeTypeTrustVerification, executiongraph.NodeTypeValidation:
		return []string{"strict_trust_gates", "required_checks_must_pass"}
	case executiongraph.NodeTypeReview:
		if node.Risk == executiongraph.RiskLevelHigh || node.Risk == executiongraph.RiskLevelCritical {
			return []string{"security_review_required"}
		}
	}
	return nil
}

func packByteEstimate(files []string, product, flow string, trust, investigation []string) int {
	n := len(product) + len(flow)
	for _, s := range files {
		n += len(s)
	}
	for _, s := range trust {
		n += len(s)
	}
	for _, s := range investigation {
		n += len(s)
	}
	return n
}
