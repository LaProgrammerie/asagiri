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

// ContractExtractor parses OpenAPI contract files under a product.
type ContractExtractor struct{}

func (ContractExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	now := time.Now().UTC()
	contractsDir := filepath.Join(productDir(repoRoot, product), "contracts")
	entries, err := os.ReadDir(contractsDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil, nil, nil
		}
		return nil, nil, nil, err
	}

	productID := knowledge.NodeID(knowledge.NodeTypeProduct, product)
	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	var warnings []string
	seenAPI := map[string]struct{}{}

	for _, ent := range entries {
		if ent.IsDir() {
			continue
		}
		name := ent.Name()
		if !strings.Contains(name, "openapi") && !strings.HasSuffix(name, ".yaml") && !strings.HasSuffix(name, ".yml") {
			continue
		}
		if !strings.Contains(strings.ToLower(name), "openapi") {
			continue
		}
		abs := filepath.Join(contractsDir, name)
		rel := relProductPath(product, "contracts", name)
		n, e, w, err := extractOpenAPI(abs, rel, productID, now)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("contracts: skip %s: %v", name, err))
			continue
		}
		for _, node := range n {
			if _, dup := seenAPI[node.ID]; dup {
				continue
			}
			seenAPI[node.ID] = struct{}{}
			nodes = append(nodes, node)
		}
		edges = append(edges, e...)
		warnings = append(warnings, w...)
	}
	return nodes, edges, warnings, nil
}

func extractOpenAPI(absPath, relPath, productID string, now time.Time) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	body, err := os.ReadFile(absPath)
	if err != nil {
		return nil, nil, nil, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return nil, nil, nil, err
	}
	paths, _ := doc["paths"].(map[string]any)
	if paths == nil {
		return nil, nil, nil, fmt.Errorf("no paths in openapi document")
	}

	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	methods := []string{"get", "post", "put", "patch", "delete", "head", "options"}

	for path, rawPathItem := range paths {
		pathItem, ok := rawPathItem.(map[string]any)
		if !ok {
			continue
		}
		for _, method := range methods {
			rawOp, ok := pathItem[method]
			if !ok {
				continue
			}
			op, ok := rawOp.(map[string]any)
			if !ok {
				continue
			}
			apiNode := apiOperationNode(method, path, relPath, "contracts", confContractHigh, now)
			if opID, _ := op["operationId"].(string); opID != "" {
				apiNode.Properties = map[string]any{"operation_id": opID}
			}
			nodes = append(nodes, apiNode)
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, productID, apiNode.ID),
				From: productID,
				To:   apiNode.ID,
				Type: knowledge.EdgeTypeOwns,
				Source: knowledge.GraphSource{
					Kind:      "contract",
					Path:      relPath,
					Extractor: "contracts",
				},
				Confidence: confContractMid,
			}, now))
		}
	}
	return nodes, edges, nil, nil
}
