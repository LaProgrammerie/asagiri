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
	confCodeHigh = 0.90
	confCodeMid  = 0.86
)

// CodeExtractor scans Go source files for files, modules, symbols, and import edges.
type CodeExtractor struct{}

func (CodeExtractor) Extract(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
	_ = ctx
	_ = product
	now := time.Now().UTC()

	var nodes []knowledge.GraphNode
	var edges []knowledge.GraphEdge
	var warnings []string

	seenModule := map[string]struct{}{}
	seenSymbol := map[string]struct{}{}
	fileBodies := map[string]string{}

	err := walkGoFiles(repoRoot, false, func(rel string) error {
		abs := filepath.Join(repoRoot, rel)
		body, err := os.ReadFile(abs)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("code: skip %s: %v", rel, err))
			return nil
		}
		fileBodies[rel] = string(body)

		parsed, err := ast.ParseGoFile(abs)
		if err != nil {
			warnings = append(warnings, fmt.Sprintf("code: parse %s: %v", rel, err))
			return nil
		}

		fileID := knowledge.NodeID(knowledge.NodeTypeFile, pathToStableKey(rel))
		nodes = append(nodes, stampNode(knowledge.GraphNode{
			ID:   fileID,
			Type: knowledge.NodeTypeFile,
			Name: filepath.Base(rel),
			Path: rel,
			Properties: map[string]any{
				"package": parsed.Package,
			},
			Source: knowledge.GraphSource{
				Kind:      "code",
				Path:      rel,
				Extractor: "code",
			},
			Confidence: confCodeHigh,
		}, now))

		modKey := sanitizeStableKey(parsed.Package)
		modID := knowledge.NodeID(knowledge.NodeTypeModule, modKey)
		if _, ok := seenModule[modID]; !ok {
			seenModule[modID] = struct{}{}
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   modID,
				Type: knowledge.NodeTypeModule,
				Name: parsed.Package,
				Path: filepath.ToSlash(filepath.Dir(rel)),
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
				},
				Confidence: confCodeMid,
			}, now))
		}

		edges = append(edges, stampEdge(knowledge.GraphEdge{
			ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, modID, fileID),
			From: modID,
			To:   fileID,
			Type: knowledge.EdgeTypeOwns,
			Source: knowledge.GraphSource{
				Kind:      "code",
				Path:      rel,
				Extractor: "code",
			},
			Confidence: confCodeMid,
		}, now))

		for _, imp := range parsed.Imports {
			depModID := knowledge.NodeID(knowledge.NodeTypeModule, sanitizeStableKey(imp))
			if _, ok := seenModule[depModID]; !ok {
				seenModule[depModID] = struct{}{}
				nodes = append(nodes, stampNode(knowledge.GraphNode{
					ID:   depModID,
					Type: knowledge.NodeTypeModule,
					Name: imp,
					Source: knowledge.GraphSource{
						Kind:      "code",
						Path:      rel,
						Extractor: "code",
						Evidence:  imp,
					},
					Confidence: confCodeMid,
				}, now))
			}
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeDependsOn, fileID, depModID),
				From: fileID,
				To:   depModID,
				Type: knowledge.EdgeTypeDependsOn,
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
					Evidence:  imp,
				},
				Confidence: confCodeMid,
			}, now))
		}

		methodSyms := parseGoMethods(string(body))
		methodNames := map[string]struct{}{}
		for _, ms := range methodSyms {
			methodNames[ms.methodName] = struct{}{}
			symID, symName := methodSymbolIDs(ms.typeName, ms.methodName)
			if _, ok := seenSymbol[symID]; ok {
				continue
			}
			seenSymbol[symID] = struct{}{}
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   symID,
				Type: knowledge.NodeTypeSymbol,
				Name: symName,
				Path: rel,
				Properties: map[string]any{
					"package": parsed.Package,
					"kind":    "method",
				},
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
				},
				Confidence: confCodeHigh,
			}, now))
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, fileID, symID),
				From: fileID,
				To:   symID,
				Type: knowledge.EdgeTypeOwns,
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
				},
				Confidence: confCodeMid,
			}, now))
		}

		for _, fn := range parsed.Funcs {
			if isGoTestHarnessFunc(fn) {
				continue
			}
			if _, isMethod := methodNames[fn]; isMethod {
				continue
			}
			symID := knowledge.NodeID(knowledge.NodeTypeSymbol, sanitizeStableKey(parsed.Package+"_"+fn))
			if _, ok := seenSymbol[symID]; ok {
				continue
			}
			seenSymbol[symID] = struct{}{}
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   symID,
				Type: knowledge.NodeTypeSymbol,
				Name: fn,
				Path: rel,
				Properties: map[string]any{
					"package": parsed.Package,
					"kind":    "func",
				},
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
				},
				Confidence: confCodeMid,
			}, now))
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeOwns, fileID, symID),
				From: fileID,
				To:   symID,
				Type: knowledge.EdgeTypeOwns,
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
				},
				Confidence: confCodeMid,
			}, now))
		}

		for _, typ := range parsed.Types {
			symID := knowledge.NodeID(knowledge.NodeTypeSymbol, sanitizeStableKey(typ))
			if _, ok := seenSymbol[symID]; ok {
				continue
			}
			seenSymbol[symID] = struct{}{}
			nodes = append(nodes, stampNode(knowledge.GraphNode{
				ID:   symID,
				Type: knowledge.NodeTypeSymbol,
				Name: typ,
				Path: rel,
				Properties: map[string]any{
					"package": parsed.Package,
					"kind":    "type",
				},
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      rel,
					Extractor: "code",
				},
				Confidence: confCodeMid,
			}, now))
		}
		return nil
	})
	if err != nil {
		return nil, nil, nil, err
	}

	emitEdges, emitWarn := linkCodeEmits(nodes, fileBodies, now)
	edges = append(edges, emitEdges...)
	warnings = append(warnings, emitWarn...)

	return nodes, edges, warnings, nil
}

type goMethod struct {
	typeName   string
	methodName string
}

func methodSymbolIDs(typeName, methodName string) (id, display string) {
	key := sanitizeStableKey(typeName + "_" + strings.ToLower(methodName))
	display = typeName + "::" + methodName
	return knowledge.NodeID(knowledge.NodeTypeSymbol, key), display
}

func parseGoMethods(source string) []goMethod {
	var out []goMethod
	lines := strings.Split(source, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "func (") {
			continue
		}
		rest := strings.TrimPrefix(line, "func (")
		closeIdx := strings.Index(rest, ")")
		if closeIdx < 0 {
			continue
		}
		recv := strings.TrimSpace(rest[:closeIdx])
		after := strings.TrimSpace(rest[closeIdx+1:])
		typeName := receiverTypeName(recv)
		nameParts := strings.FieldsFunc(after, func(r rune) bool {
			return r == '(' || r == ' ' || r == '\t'
		})
		if len(nameParts) == 0 {
			continue
		}
		methodName := nameParts[0]
		if typeName == "" || methodName == "" {
			continue
		}
		out = append(out, goMethod{typeName: typeName, methodName: methodName})
	}
	return out
}

// isGoTestHarnessFunc reports Test*/Benchmark* names; godoc Example* live in *_test.go only.
func isGoTestHarnessFunc(name string) bool {
	return strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Benchmark")
}

func receiverTypeName(recv string) string {
	recv = strings.TrimSpace(recv)
	parts := strings.Fields(recv)
	if len(parts) == 0 {
		return ""
	}
	typ := parts[len(parts)-1]
	typ = strings.TrimPrefix(typ, "*")
	return typ
}

func linkCodeEmits(nodes []knowledge.GraphNode, bodies map[string]string, now time.Time) ([]knowledge.GraphEdge, []string) {
	eventByName := map[string]string{}
	for _, n := range nodes {
		if n.Type == knowledge.NodeTypeEvent {
			eventByName[n.Name] = n.ID
		}
	}
	if len(eventByName) == 0 {
		return nil, nil
	}

	var edges []knowledge.GraphEdge
	var warnings []string
	seen := map[string]struct{}{}

	for _, n := range nodes {
		if n.Type != knowledge.NodeTypeSymbol {
			continue
		}
		body, ok := bodies[n.Path]
		if !ok {
			continue
		}
		for eventName, eventID := range eventByName {
			if !strings.Contains(body, eventName) {
				continue
			}
			edgeID := knowledge.EdgeID(knowledge.EdgeTypeEmits, n.ID, eventID)
			if _, dup := seen[edgeID]; dup {
				continue
			}
			seen[edgeID] = struct{}{}
			edges = append(edges, stampEdge(knowledge.GraphEdge{
				ID:   edgeID,
				From: n.ID,
				To:   eventID,
				Type: knowledge.EdgeTypeEmits,
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      n.Path,
					Extractor: "code",
					Evidence:  eventName,
				},
				Confidence: confCodeMid,
			}, now))
		}
	}
	return edges, warnings
}
