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

// ObservabilityExtractor parses product contracts/observability.yaml.
type ObservabilityExtractor struct{}

type observabilityDocument struct {
	Requirements []string `yaml:"requirements"`
	Metrics      []struct {
		Name string `yaml:"name"`
	} `yaml:"metrics"`
	Traces []struct {
		Name string `yaml:"name"`
	} `yaml:"traces"`
}

func (ObservabilityExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	rel := relProductPath(product, "contracts", "observability.yaml")
	abs := filepath.Join(productDir(repoRoot, product), "contracts", "observability.yaml")
	body, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	var doc observabilityDocument
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, nil, nil, fmt.Errorf("observability: %w", err)
	}

	productID := knowledge.NodeID(knowledge.NodeTypeProduct, product)
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	var warnings []string

	addMetric := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		metricID := knowledge.NodeID(knowledge.NodeTypeMetric, sanitizeStableKey(name))
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   metricID,
			Type: knowledge.NodeTypeMetric,
			Name: name,
			Path: rel,
			Source: knowledge.GraphSource{
				Kind:      "observability",
				Path:      rel,
				Extractor: "observability",
			},
			Confidence: confContractMid,
		}, now))
		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeObserves, productID, metricID),
			From: productID,
			To:   metricID,
			Type: knowledge.EdgeTypeObserves,
			Source: knowledge.GraphSource{
				Kind:      "observability",
				Path:      rel,
				Extractor: "observability",
			},
			Confidence: confContractMid,
		}, now))
	}

	for _, req := range doc.Requirements {
		addMetric(req)
	}
	for _, m := range doc.Metrics {
		addMetric(m.Name)
	}
	for _, tr := range doc.Traces {
		name := strings.TrimSpace(tr.Name)
		if name == "" {
			continue
		}
		traceID := knowledge.NodeID(knowledge.NodeTypeTrace, sanitizeStableKey(name))
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   traceID,
			Type: knowledge.NodeTypeTrace,
			Name: name,
			Path: rel,
			Source: knowledge.GraphSource{
				Kind:      "observability",
				Path:      rel,
				Extractor: "observability",
			},
			Confidence: confContractMid,
		}, now))
		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeObserves, productID, traceID),
			From: productID,
			To:   traceID,
			Type: knowledge.EdgeTypeObserves,
			Source: knowledge.GraphSource{
				Kind:      "observability",
				Path:      rel,
				Extractor: "observability",
			},
			Confidence: confContractMid,
		}, now))
	}

	if len(nodes) == 0 {
		warnings = append(warnings, rel+": observability contract has no metrics or traces")
	}
	return nodes, edges, warnings, nil
}
