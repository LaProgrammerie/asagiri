package extractors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"gopkg.in/yaml.v3"
)

// FlowExtractor parses product flow YAML files.
type FlowExtractor struct{}

func (FlowExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	flowDir := filepath.Join(productDir(repoRoot, product), "flows")
	entries, err := os.ReadDir(flowDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	var warnings []string

	productID := knowledge.NodeID(knowledge.NodeTypeProduct, product)
	productRel := relProductPath(product)
	nodes = append(nodes, stampNode(knowledge.GraphNode{
		ID:   productID,
		Type: knowledge.NodeTypeProduct,
		Name: product,
		Path: productRel,
		Source: knowledge.GraphSource{
			Kind:      "product",
			Path:      productRel,
			Extractor: "flows",
		},
		Confidence: confFlowHigh,
	}, now))

	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".flow.yaml") {
			continue
		}
		abs := filepath.Join(flowDir, ent.Name())
		rel := relProductPath(product, "flows", ent.Name())
		n, e, w, err := extractFlowFile(abs, rel, product, productID, now)
		if err != nil {
			return nil, nil, nil, fmt.Errorf("flows: %s: %w", ent.Name(), err)
		}
		nodes = append(nodes, n...)
		edges = append(edges, e...)
		warnings = append(warnings, w...)
	}
	return nodes, edges, warnings, nil
}

type flowDocument struct {
	ID          string `yaml:"id"`
	Title       string `yaml:"title"`
	EntryScreen string `yaml:"entry_screen"`
	Steps       []struct {
		ID          string   `yaml:"id"`
		Screen      string   `yaml:"screen"`
		Action      string   `yaml:"action"`
		Next        string   `yaml:"next"`
		ContractRef string   `yaml:"contract_ref"`
		Sensitive   bool     `yaml:"sensitive"`
		Errors      []string `yaml:"errors"`
	} `yaml:"steps"`
	Outcome string `yaml:"outcome"`
}

func extractFlowFile(absPath, relPath, product, productID string, now time.Time) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	body, err := os.ReadFile(absPath)
	if err != nil {
		return nil, nil, nil, err
	}
	var doc flowDocument
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, nil, nil, err
	}
	if doc.ID == "" {
		return nil, nil, nil, fmt.Errorf("flow id required")
	}

	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	var warnings []string

	flowID := knowledge.NodeID(knowledge.NodeTypeFlow, doc.ID)
	flowName := doc.Title
	if flowName == "" {
		flowName = doc.ID
	}
	nodes = append(nodes, stampNode(knowledge.GraphNode{
		ID:   flowID,
		Type: knowledge.NodeTypeFlow,
		Name: flowName,
		Path: relPath,
		Properties: map[string]any{
			"product":      product,
			"entry_screen": doc.EntryScreen,
			"outcome":      doc.Outcome,
		},
		Source: knowledge.GraphSource{
			Kind:      "flow",
			Path:      relPath,
			Extractor: "flows",
		},
		Confidence: confFlowHigh,
	}, now))

	edges = append(edges, stampEdge(knowledge.GraphEdge{
		ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, productID, flowID),
		From: productID,
		To:   flowID,
		Type: knowledge.EdgeTypeOwns,
		Source: knowledge.GraphSource{
			Kind:      "flow",
			Path:      relPath,
			Extractor: "flows",
		},
		Confidence: confFlowMid,
	}, now))

	seenActions := map[string]struct{}{}
	for _, step := range doc.Steps {
		if step.ID == "" {
			warnings = append(warnings, fmt.Sprintf("%s: step without id skipped", relPath))
			continue
		}
		stepID := knowledge.NodeID(knowledge.NodeTypeFlowStep, doc.ID+"_"+step.ID)
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   stepID,
			Type: knowledge.NodeTypeFlowStep,
			Name: step.ID,
			Path: relPath,
			Properties: map[string]any{
				"screen": step.Screen,
				"next":   step.Next,
			},
			Source: knowledge.GraphSource{
				Kind:      "flow",
				Path:      relPath,
				Extractor: "flows",
				Evidence:  step.Action,
			},
			Confidence: confFlowMid,
		}, now))

		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, flowID, stepID),
			From: flowID,
			To:   stepID,
			Type: knowledge.EdgeTypeOwns,
			Source: knowledge.GraphSource{
				Kind:      "flow",
				Path:      relPath,
				Extractor: "flows",
			},
			Confidence: confFlowMid,
		}, now))

		if step.Action == "" {
			continue
		}
		actionID := knowledge.NodeID(knowledge.NodeTypeAction, step.Action)
		if _, ok := seenActions[actionID]; !ok {
			seenActions[actionID] = struct{}{}
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   actionID,
				Type: knowledge.NodeTypeAction,
				Name: step.Action,
				Path: relPath,
				Source: knowledge.GraphSource{
					Kind:      "flow",
					Path:      relPath,
					Extractor: "flows",
				},
				Confidence: confFlowHigh,
			}, now))
		}

		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeRequires, flowID, actionID),
			From: flowID,
			To:   actionID,
			Type: knowledge.EdgeTypeRequires,
			Source: knowledge.GraphSource{
				Kind:      "flow",
				Path:      relPath,
				Extractor: "flows",
			},
			Confidence: confFlowHigh,
		}, now))

		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeRequires, stepID, actionID),
			From: stepID,
			To:   actionID,
			Type: knowledge.EdgeTypeRequires,
			Source: knowledge.GraphSource{
				Kind:      "flow",
				Path:      relPath,
				Extractor: "flows",
			},
			Confidence: confFlowMid,
		}, now))

		method, path, ok := parseContractRef(step.ContractRef)
		if !ok {
			if strings.TrimSpace(step.ContractRef) != "" {
				warnings = append(warnings, fmt.Sprintf("%s: unresolved contract_ref %q", relPath, step.ContractRef))
			}
			continue
		}
		apiNode := apiOperationNode(method, path, relPath, "flows", confFlowMid, now)
		nodes = append(nodes, apiNode)
		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeRequires, actionID, apiNode.ID),
			From: actionID,
			To:   apiNode.ID,
			Type: knowledge.EdgeTypeRequires,
			Source: knowledge.GraphSource{
				Kind:      "flow",
				Path:      relPath,
				Extractor: "flows",
				Evidence:  step.ContractRef,
			},
			Confidence: confFlowMid,
		}, now))
	}
	return nodes, edges, warnings, nil
}
