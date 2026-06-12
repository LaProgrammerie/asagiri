package coordination

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

const coordinationGraphFile = "coordination-graph.json"

// CoordinationLinkKind classifies runtime coordination edges (spec-my-D §17).
type CoordinationLinkKind string

const (
	LinkAgent         CoordinationLinkKind = "agent"
	LinkTask          CoordinationLinkKind = "task"
	LinkFlow          CoordinationLinkKind = "flow"
	LinkBranch        CoordinationLinkKind = "branch"
	LinkReview        CoordinationLinkKind = "review"
	LinkTrust         CoordinationLinkKind = "trust"
	LinkInvestigation CoordinationLinkKind = "investigation"
)

// CoordinationLink connects two entities in the runtime coordination graph.
type CoordinationLink struct {
	Kind CoordinationLinkKind `json:"kind"`
	From string               `json:"from"`
	To   string               `json:"to"`
}

// CoordinationGraph models agent↔task↔flow runtime links (spec-my-D §17).
type CoordinationGraph struct {
	GraphID string             `json:"graph_id"`
	Product string             `json:"product,omitempty"`
	Flow    string             `json:"flow,omitempty"`
	Links   []CoordinationLink `json:"links"`
}

// BuildCoordinationGraph derives links from an execution graph and assignments.
func BuildCoordinationGraph(graph ExecutionGraph, assignments []AgentAssignment) CoordinationGraph {
	asgByNode := make(map[string]AgentAssignment, len(assignments))
	for _, a := range assignments {
		asgByNode[a.NodeID] = a
	}

	cg := CoordinationGraph{
		GraphID: graph.ID,
		Product: graph.Product,
		Flow:    graph.Flow,
	}

	if graph.Flow != "" {
		cg.Links = append(cg.Links, CoordinationLink{Kind: LinkFlow, From: graph.Product, To: graph.Flow})
	}

	for _, node := range graph.Nodes {
		asg, ok := asgByNode[node.ID]
		agentRef := node.Agent
		if ok && asg.AgentRef != "" {
			agentRef = asg.AgentRef
		}
		if agentRef != "" {
			cg.Links = append(cg.Links, CoordinationLink{
				Kind: LinkAgent,
				From: agentRef,
				To:   node.ID,
			})
		}
		if node.Task != "" {
			cg.Links = append(cg.Links, CoordinationLink{
				Kind: LinkTask,
				From: node.ID,
				To:   node.Task,
			})
		}
		switch node.Type {
		case executiongraph.NodeTypeInvestigation:
			cg.Links = append(cg.Links, CoordinationLink{Kind: LinkInvestigation, From: node.ID, To: graph.Flow})
		case executiongraph.NodeTypeReview:
			cg.Links = append(cg.Links, CoordinationLink{Kind: LinkReview, From: node.ID, To: agentRef})
		case executiongraph.NodeTypeTrustVerification:
			cg.Links = append(cg.Links, CoordinationLink{Kind: LinkTrust, From: node.ID, To: graph.Flow})
		}
		if node.Risk == executiongraph.RiskLevelHigh || node.Risk == executiongraph.RiskLevelCritical {
			cg.Links = append(cg.Links, CoordinationLink{Kind: LinkBranch, From: node.ID, To: string(node.Risk)})
		}
	}
	return cg
}

// PersistCoordinationGraph writes .asagiri/graphs/<graphID>/coordination-graph.json.
func PersistCoordinationGraph(repoRoot string, cg CoordinationGraph) (string, error) {
	if strings.TrimSpace(repoRoot) == "" {
		return "", fmt.Errorf("%w: repo root required", ErrInvalidGraph)
	}
	if err := executiongraph.ValidateGraphID(cg.GraphID); err != nil {
		return "", err
	}
	dir := filepath.Join(repoRoot, ".asagiri", "graphs", cg.GraphID)
	relToRepo, err := filepath.Rel(repoRoot, dir)
	if err != nil || strings.HasPrefix(relToRepo, "..") {
		return "", fmt.Errorf("%w: graph path must not escape repo", ErrInvalidGraph)
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", fmt.Errorf("persist coordination graph: %w", err)
	}
	body, err := json.MarshalIndent(cg, "", "  ")
	if err != nil {
		return "", fmt.Errorf("persist coordination graph: %w", err)
	}
	path := filepath.Join(dir, coordinationGraphFile)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return "", fmt.Errorf("persist coordination graph: %w", err)
	}
	return path, nil
}
