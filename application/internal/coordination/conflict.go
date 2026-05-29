package coordination

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

// ConflictCategory classifies detected coordination conflicts (spec-my-D §15).
type ConflictCategory string

const (
	ConflictConcurrentEdit ConflictCategory = "concurrent_edit"
	ConflictFileOverlap    ConflictCategory = "file_overlap"
	ConflictContractDrift  ConflictCategory = "contract_drift"
	ConflictFlowDrift      ConflictCategory = "flow_drift"
	ConflictTrustDowngrade ConflictCategory = "trust_downgrade"
)

// Conflict describes one detected issue.
type Conflict struct {
	Category ConflictCategory `json:"category"`
	Message  string           `json:"message"`
	NodeIDs  []string         `json:"node_ids,omitempty"`
}

// ConflictDetector finds cross-agent conflicts.
type ConflictDetector interface {
	Detect(ctx context.Context, graph ExecutionGraph) ([]Conflict, error)
}

// DefaultConflictDetector applies file overlap, contract/flow drift, and trust downgrade heuristics.
type DefaultConflictDetector struct{}

// Detect scans the graph for coordination conflicts.
func (DefaultConflictDetector) Detect(_ context.Context, graph ExecutionGraph) ([]Conflict, error) {
	var conflicts []Conflict
	conflicts = append(conflicts, detectFileOverlaps(graph)...)
	conflicts = append(conflicts, detectContractDrift(graph)...)
	conflicts = append(conflicts, detectFlowDrift(graph)...)
	conflicts = append(conflicts, detectTrustDowngrade(graph)...)
	return conflicts, nil
}

func detectFileOverlaps(graph ExecutionGraph) []Conflict {
	fileNodes := make(map[string][]string)
	for _, n := range graph.Nodes {
		for _, f := range n.Outputs {
			f = normalizePath(f)
			if f == "" {
				continue
			}
			fileNodes[f] = append(fileNodes[f], n.ID)
		}
	}
	var out []Conflict
	for path, nodes := range fileNodes {
		if len(nodes) < 2 {
			continue
		}
		out = append(out, Conflict{
			Category: ConflictFileOverlap,
			Message:  fmt.Sprintf("file %q touched by multiple nodes", path),
			NodeIDs:  append([]string(nil), nodes...),
		})
	}
	return out
}

func detectContractDrift(graph ExecutionGraph) []Conflict {
	var contractNodes []executiongraph.GraphNode
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeContractGeneration {
			contractNodes = append(contractNodes, n)
		}
	}
	if len(contractNodes) < 2 {
		return nil
	}
	sets := make([]map[string]struct{}, len(contractNodes))
	for i, n := range contractNodes {
		sets[i] = outputSet(n.Outputs)
	}
	for i := 0; i < len(contractNodes); i++ {
		for j := i + 1; j < len(contractNodes); j++ {
			if !setsOverlap(sets[i], sets[j]) && len(sets[i]) > 0 && len(sets[j]) > 0 {
				return []Conflict{{
					Category: ConflictContractDrift,
					Message:  "conflicting contract_generation outputs",
					NodeIDs:  []string{contractNodes[i].ID, contractNodes[j].ID},
				}}
			}
		}
	}
	return nil
}

func detectFlowDrift(graph ExecutionGraph) []Conflict {
	typeCounts := make(map[executiongraph.NodeType]int)
	for _, n := range graph.Nodes {
		typeCounts[n.Type]++
	}
	var archNodes []string
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeArchitectureDerivation {
			archNodes = append(archNodes, n.ID)
		}
	}
	if len(archNodes) >= 2 {
		return []Conflict{{
			Category: ConflictFlowDrift,
			Message:  "duplicate architecture_derivation nodes may diverge on flow",
			NodeIDs:  archNodes,
		}}
	}
	return nil
}

func detectTrustDowngrade(graph ExecutionGraph) []Conflict {
	failedValidation := make(map[string]struct{})
	for _, n := range graph.Nodes {
		if n.Type == executiongraph.NodeTypeValidation && n.Status == executiongraph.NodeStatusFailed {
			failedValidation[n.ID] = struct{}{}
		}
	}
	if len(failedValidation) == 0 {
		return nil
	}
	var out []Conflict
	for _, n := range graph.Nodes {
		if n.Type != executiongraph.NodeTypeTrustVerification {
			continue
		}
		for failedID := range failedValidation {
			if dependsOn(graph, n.ID, failedID) {
				out = append(out, Conflict{
					Category: ConflictTrustDowngrade,
					Message:  "trust verification scheduled after failed validation",
					NodeIDs:  []string{n.ID, failedID},
				})
			}
		}
	}
	return out
}

func dependsOn(graph ExecutionGraph, from, to string) bool {
	visited := make(map[string]struct{})
	var walk func(string) bool
	walk = func(id string) bool {
		if id == to {
			return true
		}
		if _, ok := visited[id]; ok {
			return false
		}
		visited[id] = struct{}{}
		for _, e := range graph.Edges {
			if e.From == id {
				if walk(e.To) {
					return true
				}
			}
		}
		return false
	}
	return walk(from)
}

func normalizePath(p string) string {
	p = strings.TrimSpace(p)
	p = strings.TrimPrefix(p, "./")
	if strings.Contains(p, "..") {
		return ""
	}
	return p
}

func outputSet(paths []string) map[string]struct{} {
	set := make(map[string]struct{}, len(paths))
	for _, p := range paths {
		if n := normalizePath(p); n != "" {
			set[n] = struct{}{}
		}
	}
	return set
}

func setsOverlap(a, b map[string]struct{}) bool {
	for k := range a {
		if _, ok := b[k]; ok {
			return true
		}
	}
	return len(a) == 0 && len(b) == 0
}
