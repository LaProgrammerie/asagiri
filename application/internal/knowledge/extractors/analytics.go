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

// AnalyticsExtractor parses product contracts/analytics.yaml (spec-my-E §10).
type AnalyticsExtractor struct{}

type analyticsDocument struct {
	Events []any `yaml:"events"`
}

func (AnalyticsExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	rel := relProductPath(product, "contracts", "analytics.yaml")
	abs := filepath.Join(productDir(repoRoot, product), "contracts", "analytics.yaml")
	body, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	entries, warnings, err := parseAnalyticsEntries(body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("analytics: %w", err)
	}
	if len(entries) == 0 {
		return nil, nil, warnings, nil
	}

	productID := knowledge.NodeID(knowledge.NodeTypeProduct, product)
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge

	for _, entry := range entries {
		kind, name, ok := parseAnalyticsEntry(entry)
		if !ok {
			warnings = append(warnings, "analytics: skipped empty entry")
			continue
		}
		switch kind {
		case "metric", "dashboard":
			metricID := knowledge.NodeID(knowledge.NodeTypeMetric, sanitizeStableKey(name))
			props := map[string]any{}
			if kind == "dashboard" {
				props["dashboard"] = true
			}
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:         metricID,
				Type:       knowledge.NodeTypeMetric,
				Name:       name,
				Path:       rel,
				Properties: props,
				Source: knowledge.GraphSource{
					Kind:      "analytics",
					Path:      rel,
					Extractor: "analytics",
				},
				Confidence: confContractMid,
			}, now))
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeConfigures, productID, metricID),
				From: productID,
				To:   metricID,
				Type: knowledge.EdgeTypeConfigures,
				Source: knowledge.GraphSource{
					Kind:      "analytics",
					Path:      rel,
					Extractor: "analytics",
				},
				Confidence: confContractMid,
			}, now))
		case "event":
			eventID := knowledge.NodeID(knowledge.NodeTypeEvent, sanitizeStableKey(name))
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   eventID,
				Type: knowledge.NodeTypeEvent,
				Name: name,
				Path: rel,
				Source: knowledge.GraphSource{
					Kind:      "analytics",
					Path:      rel,
					Extractor: "analytics",
				},
				Confidence: confEventHigh,
			}, now))
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeProduces, productID, eventID),
				From: productID,
				To:   eventID,
				Type: knowledge.EdgeTypeProduces,
				Source: knowledge.GraphSource{
					Kind:      "analytics",
					Path:      rel,
					Extractor: "analytics",
				},
				Confidence: confContractMid,
			}, now))
		default:
			warnings = append(warnings, fmt.Sprintf("analytics: unknown entry kind %q", kind))
		}
	}

	obsAbs := filepath.Join(productDir(repoRoot, product), "contracts", "observability.yaml")
	if obsBody, err := os.ReadFile(obsAbs); err == nil {
		warnings = append(warnings, warnAnalyticsObservabilityGap(obsBody, entries)...)
	}

	return nodes, edges, warnings, nil
}

func parseAnalyticsEntries(body []byte) ([]string, []string, error) {
	var doc analyticsDocument
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, nil, err
	}
	var warnings []string
	var out []string
	for _, item := range doc.Events {
		switch v := item.(type) {
		case string:
			if s := strings.TrimSpace(v); s != "" {
				out = append(out, s)
			}
		case map[string]any:
			if n, _ := v["name"].(string); strings.TrimSpace(n) != "" {
				out = append(out, n)
			} else if n, _ := v["id"].(string); strings.TrimSpace(n) != "" {
				out = append(out, n)
			}
		default:
			warnings = append(warnings, "analytics: unsupported events entry type")
		}
	}
	return out, warnings, nil
}

func parseAnalyticsEntry(entry string) (kind, name string, ok bool) {
	entry = strings.TrimSpace(entry)
	if entry == "" {
		return "", "", false
	}
	if before, after, found := strings.Cut(entry, ":"); found {
		return strings.ToLower(strings.TrimSpace(before)), strings.TrimSpace(after), after != ""
	}
	return "event", entry, true
}

func warnAnalyticsObservabilityGap(obsBody []byte, analyticsEntries []string) []string {
	var doc observabilityDocument
	if err := yaml.Unmarshal(obsBody, &doc); err != nil {
		return nil
	}
	obsMetrics := map[string]struct{}{}
	for _, m := range doc.Metrics {
		if n := strings.TrimSpace(m.Name); n != "" {
			obsMetrics[n] = struct{}{}
		}
	}
	analyticsMetrics := map[string]struct{}{}
	for _, entry := range analyticsEntries {
		kind, name, ok := parseAnalyticsEntry(entry)
		if !ok || (kind != "metric" && kind != "dashboard") {
			continue
		}
		analyticsMetrics[name] = struct{}{}
	}
	var warnings []string
	for name := range obsMetrics {
		if _, ok := analyticsMetrics[name]; !ok {
			warnings = append(warnings, fmt.Sprintf("analytics: observability metric %q has no analytics contract entry", name))
		}
	}
	return warnings
}
