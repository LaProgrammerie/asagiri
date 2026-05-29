package extractors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"gopkg.in/yaml.v3"
)

// PermissionExtractor parses permissions from contracts/permissions.yaml.
type PermissionExtractor struct{}

func (PermissionExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	rel := relProductPath(product, "contracts", "permissions.yaml")
	abs := filepath.Join(productDir(repoRoot, product), "contracts", "permissions.yaml")
	body, err := os.ReadFile(abs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	ids, warnings, err := parsePermissionIDs(body)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("permissions: %w", err)
	}
	if len(ids) == 0 {
		return nil, nil, warnings, nil
	}

	productID := knowledge.NodeID(knowledge.NodeTypeProduct, product)
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge

	for _, id := range ids {
		permID := knowledge.NodeID(knowledge.NodeTypePermission, sanitizeStableKey(id))
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   permID,
			Type: knowledge.NodeTypePermission,
			Name: id,
			Path: rel,
			Source: knowledge.GraphSource{
				Kind:      "permissions",
				Path:      rel,
				Extractor: "permissions",
			},
			Confidence: confPermissionHigh,
		}, now))
		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, productID, permID),
			From: productID,
			To:   permID,
			Type: knowledge.EdgeTypeOwns,
			Source: knowledge.GraphSource{
				Kind:      "permissions",
				Path:      rel,
				Extractor: "permissions",
			},
			Confidence: confContractMid,
		}, now))
	}
	return nodes, edges, warnings, nil
}

func parsePermissionIDs(body []byte) ([]string, []string, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, nil, err
	}
	var warnings []string
	var ids []string

	switch perms := doc["permissions"].(type) {
	case []any:
		for _, item := range perms {
			switch v := item.(type) {
			case string:
				if v != "" {
					ids = append(ids, v)
				}
			case map[string]any:
				if n, _ := v["id"].(string); n != "" {
					ids = append(ids, n)
				} else if n, _ := v["name"].(string); n != "" {
					ids = append(ids, n)
				}
			}
		}
	case map[string]any:
		for k := range perms {
			ids = append(ids, k)
		}
	default:
		warnings = append(warnings, "permissions.yaml: no permissions section found")
	}
	return ids, warnings, nil
}
