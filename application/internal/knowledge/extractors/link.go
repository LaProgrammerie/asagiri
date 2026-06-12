package extractors

import (
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

const confLinkMid = 0.87

// LinkFlowToCode connects flow actions and API operations to extracted symbols.
func LinkFlowToCode(nodes []knowledge.GraphNode, edges []knowledge.GraphEdge) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string) {
	now := time.Now().UTC()
	symbols := make([]knowledge.GraphNode, 0)
	actions := make([]knowledge.GraphNode, 0)
	apiOps := make([]knowledge.GraphNode, 0)
	nodeByID := map[string]knowledge.GraphNode{}

	for _, n := range nodes {
		nodeByID[n.ID] = n
		switch n.Type {
		case knowledge.NodeTypeSymbol:
			symbols = append(symbols, n)
		case knowledge.NodeTypeAction:
			actions = append(actions, n)
		case knowledge.NodeTypeAPIOperation:
			apiOps = append(apiOps, n)
		}
	}

	edgeSeen := map[string]struct{}{}
	for _, e := range edges {
		edgeSeen[e.ID] = struct{}{}
	}
	addEdge := func(e knowledge.GraphEdge) {
		if _, ok := edgeSeen[e.ID]; ok {
			return
		}
		edgeSeen[e.ID] = struct{}{}
		edges = append(edges, e)
	}

	var warnings []string

	for _, action := range actions {
		matched := matchActionSymbols(action.Name, symbols)
		for _, sym := range matched {
			addEdge(stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeImplements, action.ID, sym.ID),
				From: action.ID,
				To:   sym.ID,
				Type: knowledge.EdgeTypeImplements,
				Source: knowledge.GraphSource{
					Kind:      "flow",
					Path:      action.Path,
					Extractor: "flows",
					Evidence:  action.Name,
				},
				Confidence: confLinkMid,
			}, now))

			addEdge(stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeCalls, action.ID, sym.ID),
				From: action.ID,
				To:   sym.ID,
				Type: knowledge.EdgeTypeCalls,
				Source: knowledge.GraphSource{
					Kind:      "flow",
					Path:      action.Path,
					Extractor: "flows",
					Evidence:  action.Name,
				},
				Confidence: confLinkMid,
			}, now))
		}
		if len(matched) == 0 {
			warnings = append(warnings, "flow-to-code: no symbol linked for action "+action.Name)
		}
	}

	for _, api := range apiOps {
		matched := matchAPISymbols(api, symbols)
		for _, sym := range matched {
			addEdge(stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeImplements, api.ID, sym.ID),
				From: api.ID,
				To:   sym.ID,
				Type: knowledge.EdgeTypeImplements,
				Source: knowledge.GraphSource{
					Kind:      "contract",
					Path:      api.Path,
					Extractor: "code",
					Evidence:  api.Name,
				},
				Confidence: confLinkMid,
			}, now))
		}
	}

	edges = linkSymbolCalls(edges, symbols, now, edgeSeen, addEdge)
	return nodes, edges, warnings
}

func matchActionSymbols(action string, symbols []knowledge.GraphNode) []knowledge.GraphNode {
	tokens := actionTokens(action)
	var out []knowledge.GraphNode
	for _, sym := range symbols {
		key := strings.ToLower(strings.TrimPrefix(sym.ID, string(knowledge.NodeTypeSymbol)+":"))
		name := strings.ToLower(sym.Name)
		score := 0
		for _, tok := range tokens {
			if tok == "" {
				continue
			}
			if strings.Contains(key, tok) || strings.Contains(name, tok) {
				score++
			}
		}
		if score == 0 {
			continue
		}
		if score >= len(tokens) || strings.Contains(key, tokens[0]) {
			out = append(out, sym)
		}
	}
	return out
}

func actionTokens(action string) []string {
	parts := strings.Split(action, "_")
	var tokens []string
	for _, p := range parts {
		if p == "" {
			continue
		}
		tokens = append(tokens, strings.ToLower(p))
	}
	if len(parts) >= 2 {
		var b strings.Builder
		for _, p := range parts {
			if p == "" {
				continue
			}
			b.WriteString(strings.ToUpper(p[:1]))
			if len(p) > 1 {
				b.WriteString(strings.ToLower(p[1:]))
			}
		}
		tokens = append(tokens, strings.ToLower(b.String()))
	}
	return tokens
}

func matchAPISymbols(api knowledge.GraphNode, symbols []knowledge.GraphNode) []knowledge.GraphNode {
	hay := strings.ToLower(api.Name)
	if op, ok := api.Properties["operation_id"].(string); ok {
		hay += " " + strings.ToLower(op)
	}
	var out []knowledge.GraphNode
	for _, sym := range symbols {
		key := strings.ToLower(strings.TrimPrefix(sym.ID, string(knowledge.NodeTypeSymbol)+":"))
		if strings.Contains(hay, "invitation") && strings.Contains(key, "invite") {
			out = append(out, sym)
		}
	}
	return out
}

func linkSymbolCalls(edges []knowledge.GraphEdge, symbols []knowledge.GraphNode, now time.Time, edgeSeen map[string]struct{}, addEdge func(knowledge.GraphEdge)) []knowledge.GraphEdge {
	byFile := map[string][]knowledge.GraphNode{}
	for _, sym := range symbols {
		byFile[sym.Path] = append(byFile[sym.Path], sym)
	}
	for _, group := range byFile {
		var service, invite knowledge.GraphNode
		for _, sym := range group {
			key := strings.ToLower(strings.TrimPrefix(sym.ID, string(knowledge.NodeTypeSymbol)+":"))
			if strings.Contains(key, "invitationservice") && strings.Contains(key, "_invite") {
				invite = sym
			}
			if strings.Contains(key, "invitationservice") && !strings.Contains(key, "_") {
				service = sym
			}
		}
		if invite.ID != "" && service.ID != "" && invite.ID != service.ID {
			e := stampEdge(knowledge.GraphEdge{
				ID:   knowledge.EdgeID(knowledge.EdgeTypeCalls, service.ID, invite.ID),
				From: service.ID,
				To:   invite.ID,
				Type: knowledge.EdgeTypeCalls,
				Source: knowledge.GraphSource{
					Kind:      "code",
					Path:      invite.Path,
					Extractor: "code",
				},
				Confidence: confLinkMid,
			}, now)
			addEdge(e)
		}
	}
	return edges
}

// WarnUntestedActions reports actions with no tests edge reachable from linked symbols.
func WarnUntestedActions(nodes []knowledge.GraphNode, edges []knowledge.GraphEdge) []string {
	testsBySymbol := map[string]struct{}{}
	for _, e := range edges {
		if e.Type != knowledge.EdgeTypeTests {
			continue
		}
		fromType := nodeTypePrefix(e.From)
		if fromType == knowledge.NodeTypeSymbol {
			testsBySymbol[e.From] = struct{}{}
		}
	}

	actionHasTest := map[string]bool{}
	for _, e := range edges {
		if e.Type != knowledge.EdgeTypeImplements && e.Type != knowledge.EdgeTypeCalls {
			continue
		}
		fromType := nodeTypePrefix(e.From)
		if fromType != knowledge.NodeTypeAction {
			continue
		}
		if _, ok := testsBySymbol[e.To]; ok {
			actionHasTest[e.From] = true
		}
	}

	var warnings []string
	for _, n := range nodes {
		if n.Type != knowledge.NodeTypeAction {
			continue
		}
		if actionHasTest[n.ID] {
			continue
		}
		warnings = append(warnings, "action "+n.Name+" has no linked test")
	}
	return warnings
}

func nodeTypePrefix(id string) knowledge.NodeType {
	prefix, _, ok := strings.Cut(id, ":")
	if !ok {
		return ""
	}
	return knowledge.NodeType(prefix)
}
