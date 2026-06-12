package extractors

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/ast"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

const (
	confTestHigh = 0.90
	confTestMid  = 0.88
)

// TestExtractor scans Go test files and links tests to symbols and source files.
type TestExtractor struct{}

func (TestExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	_ = product
	now := time.Now().UTC()

	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	var warnings []string

	symbolByKey, idxWarn := buildSymbolIndex(repoRoot)
	warnings = append(warnings, idxWarn...)

	err := walkGoFiles(repoRoot, true, func(rel string) error {
		abs := filepath.Join(repoRoot, rel)
		body, err := os.ReadFile(abs)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("tests: skip %s: %v", rel, err))
			return nil
		}
		source := strings.TrimSuffix(rel, "_test.go") + ".go"
		sourceID := knowledge.NodeID(knowledge.NodeTypeFile, pathToStableKey(source))

		parsed, err := ast.ParseGoFile(abs)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("tests: parse %s: %v", rel, err))
			return nil
		}

		for _, fn := range parsed.Funcs {
			if !strings.HasPrefix(fn, "Test") {
				continue
			}
			testID := knowledge.NodeID(knowledge.NodeTypeTest, testNodeKey(fn))
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   testID,
				Type: knowledge.NodeTypeTest,
				Name: fn,
				Path: rel,
				Source: knowledge.GraphSource{
					Kind:      "test",
					Path:      rel,
					Extractor: "tests",
				},
				Confidence: confTestHigh,
			}, now))

			if fileExists(filepath.Join(repoRoot, source)) {
				edges = append(edges, stampEdge(knowledge.GraphEdge{
					ID:   knowledge.EdgeID(knowledge.EdgeTypeTests, testID, sourceID),
					From: testID,
					To:   sourceID,
					Type: knowledge.EdgeTypeTests,
					Source: knowledge.GraphSource{
						Kind:      "test",
						Path:      rel,
						Extractor: "tests",
					},
					Confidence: confTestMid,
				}, now))
			}

			for symID := range matchTestToSymbols(fn, string(body), symbolByKey) {
				edges = append(edges, stampEdge(knowledge.GraphEdge{
					ID:   knowledge.EdgeID(knowledge.EdgeTypeTests, symID, testID),
					From: symID,
					To:   testID,
					Type: knowledge.EdgeTypeTests,
					Source: knowledge.GraphSource{
						Kind:      "test",
						Path:      rel,
						Extractor: "tests",
						Evidence:  fn,
					},
					Confidence: confTestHigh,
				}, now))
			}
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	return nodes, edges, warnings, nil
}

func testNodeKey(fn string) string {
	base := strings.TrimPrefix(fn, "Test")
	if base == "" {
		return sanitizeStableKey(fn)
	}
	return sanitizeStableKey(base + "Test")
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func matchTestToSymbols(testName, body string, symbolByKey map[string]string) map[string]struct{} {
	matches := map[string]struct{}{}
	subject := strings.TrimPrefix(testName, "Test")
	if subject == "" {
		return matches
	}
	candidates := []string{
		subject,
		strings.TrimSuffix(subject, "s"),
	}
	for _, c := range candidates {
		if c == "" {
			continue
		}
		lower := strings.ToLower(c)
		for key, symID := range symbolByKey {
			if strings.Contains(strings.ToLower(key), lower) {
				matches[symID] = struct{}{}
			}
		}
	}
	// Exact symbol name mentions in test body only (avoids O(tests×symbols) substring noise).
	for _, symID := range symbolByKey {
		nameKey := strings.TrimPrefix(symID, string(knowledge.NodeTypeSymbol)+":")
		if nameKey == "" {
			continue
		}
		if strings.Contains(body, nameKey) {
			matches[symID] = struct{}{}
		}
	}
	return matches
}

func buildSymbolIndex(repoRoot string) (map[string]string, []string) {
	index := map[string]string{}
	var warnings []string
	_ = walkGoFiles(repoRoot, false, func(rel string) error {
		abs := filepath.Join(repoRoot, rel)
		body, err := os.ReadFile(abs)
		if err != nil {
			return nil
		}
		parsed, err := ast.ParseGoFile(abs)
		if err != nil {
			return nil
		}
		methodNames := map[string]struct{}{}
		for _, ms := range parseGoMethods(string(body)) {
			methodNames[ms.methodName] = struct{}{}
			symID, symName := methodSymbolIDs(ms.typeName, ms.methodName)
			key := strings.TrimPrefix(symID, string(knowledge.NodeTypeSymbol)+":")
			index[strings.ToLower(key)] = symID
			index[strings.ToLower(symName)] = symID
		}
		for _, fn := range parsed.Funcs {
			if strings.HasPrefix(fn, "Test") {
				continue
			}
			if _, isMethod := methodNames[fn]; isMethod {
				continue
			}
			symID := knowledge.NodeID(knowledge.NodeTypeSymbol, sanitizeStableKey(parsed.Package+"_"+fn))
			key := strings.TrimPrefix(symID, string(knowledge.NodeTypeSymbol)+":")
			index[strings.ToLower(key)] = symID
		}
		for _, typ := range parsed.Types {
			symID := knowledge.NodeID(knowledge.NodeTypeSymbol, sanitizeStableKey(typ))
			key := strings.TrimPrefix(symID, string(knowledge.NodeTypeSymbol)+":")
			index[strings.ToLower(key)] = symID
		}
		return nil
	})
	return index, warnings
}

// RegisterSymbolIndex builds a lookup table from symbol nodes for test linking.
func RegisterSymbolIndex(nodes []knowledge.GraphNode) map[string]string {
	out := make(map[string]string)
	for _, n := range nodes {
		if n.Type != knowledge.NodeTypeSymbol {
			continue
		}
		key := strings.TrimPrefix(n.ID, string(knowledge.NodeTypeSymbol)+":")
		out[strings.ToLower(key)] = n.ID
		out[strings.ToLower(n.Name)] = n.ID
	}
	return out
}
