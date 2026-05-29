package executiongraph

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"gopkg.in/yaml.v3"
)

const productsRel = ".asagiri/products"

var cycleFormingEdgeTypes = map[EdgeType]bool{
	EdgeTypeRequires:          true,
	EdgeTypeMustRunAfter:      true,
	EdgeTypeBlocks:            true,
	EdgeTypeRollbackDependsOn: true,
}

// DefaultDependencyInferer infers graph edges from tasks, flows, contracts, and files (spec §10).
type DefaultDependencyInferer struct {
	LoadBundle func(repoRoot, productID string) (analysis.Bundle, error)
	ReadFile   func(path string) ([]byte, error)
}

func (d DefaultDependencyInferer) Infer(_ context.Context, input DependencyInput) ([]GraphEdge, error) {
	if input.Product == "" {
		return nil, fmt.Errorf("dependency infer: product required")
	}
	loadBundle := d.LoadBundle
	if loadBundle == nil {
		loadBundle = analysis.LoadBundle
	}
	readFile := d.ReadFile
	if readFile == nil {
		readFile = os.ReadFile
	}

	var bundle analysis.Bundle
	var bundleErr error
	if input.RepoRoot != "" {
		bundle, bundleErr = loadBundle(input.RepoRoot, input.Product)
	}

	bindings := sortedBindings(input.TaskBindings)
	nodeSet := nodeIDSet(input.Nodes)
	edges := make([]GraphEdge, 0)

	edges = append(edges, inferFlowContractChain(bindings, bundle, bundleErr)...)
	edges = append(edges, inferAPIUsage(bindings, bundle, bundleErr)...)
	edges = append(edges, inferSharedFileConflicts(bindings)...)
	edges = append(edges, inferContractBlocking(input, bindings, nodeSet, readFile)...)
	edges = append(edges, inferSecurityReviewChain(input, bindings, nodeSet)...)
	edges = append(edges, inferBackwardCompat(input, bindings, nodeSet, bundle, bundleErr)...)

	return dedupeEdges(edges), nil
}

// DetectCycles returns ErrCycleDetected when cycle-forming edges contain a directed cycle.
func DetectCycles(nodes []GraphNode, edges []GraphEdge) error {
	ids := make([]string, 0, len(nodes))
	for _, n := range nodes {
		ids = append(ids, n.ID)
	}
	cycle, ok := findCycle(ids, edges)
	if !ok {
		return nil
	}
	return fmt.Errorf("%w: %s", ErrCycleDetected, strings.Join(cycle, " -> "))
}

func inferFlowContractChain(bindings []TaskBinding, bundle analysis.Bundle, bundleErr error) []GraphEdge {
	if len(bindings) < 2 {
		return nil
	}
	edges := make([]GraphEdge, 0)
	for i := 1; i < len(bindings); i++ {
		prev := bindings[i-1]
		curr := bindings[i]
		if prev.NodeID == "" || curr.NodeID == "" {
			continue
		}
		prevRef := strings.TrimSpace(prev.ContractRef)
		currRef := strings.TrimSpace(curr.ContractRef)
		if currRef == "" {
			continue
		}
		if prevRef != "" && !strings.HasPrefix(prevRef, "TODO:") {
			edges = append(edges, GraphEdge{
				From:   prev.NodeID,
				To:     curr.NodeID,
				Type:   EdgeTypeRequires,
				Reason: "flow step order with contract dependency",
			})
			continue
		}
		if bundleErr == nil && graphHasRoute(bundle.Graphs["api"], currRef) && prevRef != "" {
			edges = append(edges, GraphEdge{
				From:   prev.NodeID,
				To:     curr.NodeID,
				Type:   EdgeTypeRequires,
				Reason: "later step consumes API from earlier step",
			})
		}
	}
	return edges
}

func inferAPIUsage(bindings []TaskBinding, bundle analysis.Bundle, bundleErr error) []GraphEdge {
	if bundleErr != nil {
		return nil
	}
	apiGraph := bundle.Graphs["api"]
	if len(apiGraph.Nodes) == 0 {
		return nil
	}

	producer := map[string]string{}
	for _, b := range bindings {
		ref := strings.TrimSpace(b.ContractRef)
		if ref == "" || strings.HasPrefix(ref, "TODO:") {
			continue
		}
		if graphHasRoute(apiGraph, ref) {
			producer[normalizeAPIRef(ref)] = b.NodeID
		}
	}

	edges := make([]GraphEdge, 0)
	for _, b := range bindings {
		ref := strings.TrimSpace(b.ContractRef)
		if ref == "" || strings.HasPrefix(ref, "TODO:") {
			continue
		}
		key := normalizeAPIRef(ref)
		if from, ok := producer[key]; ok && from != b.NodeID {
			edges = append(edges, GraphEdge{
				From:   from,
				To:     b.NodeID,
				Type:   EdgeTypeRequires,
				Reason: fmt.Sprintf("task uses API %q produced by upstream task", ref),
			})
		}
	}
	return edges
}

func inferSharedFileConflicts(bindings []TaskBinding) []GraphEdge {
	edges := make([]GraphEdge, 0)
	for i := 0; i < len(bindings); i++ {
		for j := i + 1; j < len(bindings); j++ {
			a := bindings[i]
			b := bindings[j]
			if a.NodeID == "" || b.NodeID == "" {
				continue
			}
			if !scopePathsOverlap(a.ScopePaths, b.ScopePaths) {
				continue
			}
			first, second := a, b
			if b.StepIndex >= 0 && (a.StepIndex < 0 || b.StepIndex < a.StepIndex) {
				first, second = b, a
			} else if a.StepIndex == b.StepIndex && a.NodeID > b.NodeID {
				first, second = b, a
			}
			edges = append(edges, GraphEdge{
				From:   first.NodeID,
				To:     second.NodeID,
				Type:   EdgeTypeMustRunAfter,
				Reason: "tasks share touched file scope; no parallel execution",
			})
		}
	}
	return edges
}

func inferContractBlocking(input DependencyInput, bindings []TaskBinding, nodeSet map[string]struct{}, readFile func(string) ([]byte, error)) []GraphEdge {
	deriveID := "derive-contracts"
	if _, ok := nodeSet[deriveID]; !ok {
		return nil
	}

	needsContract := false
	for _, b := range bindings {
		ref := strings.TrimSpace(b.ContractRef)
		if strings.HasPrefix(ref, "TODO:") {
			needsContract = true
		}
		if b.Sensitive && !permissionCoversAction(input, b.Action, readFile) {
			needsContract = true
		}
	}
	if !needsContract {
		return nil
	}

	edges := make([]GraphEdge, 0)
	for _, b := range bindings {
		if b.NodeID == "" || b.NodeID == deriveID {
			continue
		}
		ref := strings.TrimSpace(b.ContractRef)
		if strings.HasPrefix(ref, "TODO:") || (b.Sensitive && !permissionCoversAction(input, b.Action, readFile)) {
			edges = append(edges, GraphEdge{
				From:   deriveID,
				To:     b.NodeID,
				Type:   EdgeTypeBlocks,
				Reason: "implementation blocked until contracts or permissions are published",
			})
		}
	}
	return edges
}

func inferSecurityReviewChain(input DependencyInput, bindings []TaskBinding, nodeSet map[string]struct{}) []GraphEdge {
	reviewID := "security-review"
	trustID := "trust-gate"
	if _, hasReview := nodeSet[reviewID]; !hasReview {
		return nil
	}

	sensitiveActions := sensitiveActionSet(input)
	edges := make([]GraphEdge, 0)
	for _, b := range bindings {
		if b.NodeID == "" {
			continue
		}
		if !b.Sensitive && !sensitiveActions[b.Action] {
			continue
		}
		edges = append(edges, GraphEdge{
			From:   b.NodeID,
			To:     reviewID,
			Type:   EdgeTypeValidates,
			Reason: fmt.Sprintf("%q is a sensitive action", b.Action),
		})
		if _, ok := nodeSet[trustID]; ok {
			edges = append(edges, GraphEdge{
				From:   reviewID,
				To:     trustID,
				Type:   EdgeTypeRequires,
				Reason: "security review required before trust verification",
			})
		}
	}
	return edges
}

func inferBackwardCompat(input DependencyInput, bindings []TaskBinding, nodeSet map[string]struct{}, bundle analysis.Bundle, bundleErr error) []GraphEdge {
	verifyID := "verify-contracts"
	if _, ok := nodeSet[verifyID]; !ok {
		return nil
	}
	if bundleErr != nil {
		return nil
	}
	apiGraph := bundle.Graphs["api"]
	deriveID := "derive-contracts"

	edges := make([]GraphEdge, 0)
	for _, b := range bindings {
		ref := strings.TrimSpace(b.ContractRef)
		if ref == "" || strings.HasPrefix(ref, "TODO:") {
			continue
		}
		if !graphHasRoute(apiGraph, ref) {
			continue
		}
		if _, ok := nodeSet[deriveID]; ok {
			edges = append(edges, GraphEdge{
				From:   deriveID,
				To:     verifyID,
				Type:   EdgeTypeValidates,
				Reason: "public API contract requires compatibility validation",
			})
		}
		if b.NodeID != "" && b.NodeID != verifyID {
			edges = append(edges, GraphEdge{
				From:   verifyID,
				To:     b.NodeID,
				Type:   EdgeTypeRequires,
				Reason: fmt.Sprintf("public API change %q requires backward compatibility check", ref),
			})
		}
	}
	return dedupeEdges(edges)
}

func permissionCoversAction(input DependencyInput, action string, readFile func(string) ([]byte, error)) bool {
	if input.RepoRoot == "" || action == "" {
		return false
	}
	path := filepath.Join(productDir(input.RepoRoot, input.Product), "contracts", "permissions.yaml")
	raw, err := readFile(path)
	if err != nil {
		return false
	}
	var doc struct {
		Roles map[string]struct {
			Permissions []string `yaml:"permissions"`
		} `yaml:"roles"`
	}
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return false
	}
	normAction := normalizeAction(action)
	for _, role := range doc.Roles {
		for _, perm := range role.Permissions {
			if permissionMatchesAction(perm, normAction) {
				return true
			}
		}
	}
	return false
}

func permissionMatchesAction(perm, action string) bool {
	perm = strings.ToLower(strings.TrimSpace(perm))
	action = strings.ToLower(strings.TrimSpace(action))
	if perm == "" || action == "" {
		return false
	}
	if strings.Contains(perm, action) || strings.Contains(action, perm) {
		return true
	}
	// invite_member -> workspace.invite or invite.*
	actionToken := strings.ReplaceAll(action, "_", ".")
	return strings.Contains(perm, actionToken) || strings.Contains(perm, strings.TrimSuffix(actionToken, ".member"))
}

func sensitiveActionSet(input DependencyInput) map[string]bool {
	out := map[string]bool{}
	if input.RepoRoot == "" || input.Flow == "" {
		return out
	}
	path := resolveFlowPath(productDir(input.RepoRoot, input.Product), input.Flow)
	raw, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	flow, err := product.ParseFlowYAML(raw)
	if err != nil {
		return out
	}
	for _, action := range flow.Security.SensitiveActions {
		out[action] = true
	}
	return out
}

func graphHasRoute(g analysis.Graph, ref string) bool {
	ref = strings.TrimSpace(ref)
	for _, n := range g.Nodes {
		if strings.Contains(n.Name, ref) || strings.Contains(n.ID, ref) {
			return true
		}
	}
	return false
}

func normalizeAPIRef(ref string) string {
	return strings.ToUpper(strings.TrimSpace(ref))
}

func normalizeAction(action string) string {
	return strings.ToLower(strings.TrimSpace(strings.ReplaceAll(action, "-", "_")))
}

func scopePathsOverlap(a, b []string) bool {
	if len(a) == 0 || len(b) == 0 {
		return false
	}
	for _, pa := range a {
		pa = strings.TrimSpace(pa)
		if pa == "" {
			continue
		}
		for _, pb := range b {
			pb = strings.TrimSpace(pb)
			if pb == "" {
				continue
			}
			if pathsConflict(pa, pb) {
				return true
			}
		}
	}
	return false
}

func pathsConflict(a, b string) bool {
	a = strings.TrimSuffix(a, "/**")
	b = strings.TrimSuffix(b, "/**")
	if a == b {
		return true
	}
	return strings.HasPrefix(a, b+"/") || strings.HasPrefix(b, a+"/")
}

func sortedBindings(bindings []TaskBinding) []TaskBinding {
	out := append([]TaskBinding(nil), bindings...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].StepIndex != out[j].StepIndex {
			return out[i].StepIndex < out[j].StepIndex
		}
		return out[i].NodeID < out[j].NodeID
	})
	return out
}

func nodeIDSet(nodes []GraphNode) map[string]struct{} {
	out := make(map[string]struct{}, len(nodes))
	for _, n := range nodes {
		out[n.ID] = struct{}{}
	}
	return out
}

func dedupeEdges(edges []GraphEdge) []GraphEdge {
	seen := make(map[string]struct{}, len(edges))
	out := make([]GraphEdge, 0, len(edges))
	for _, e := range edges {
		key := e.From + "\x00" + e.To + "\x00" + string(e.Type)
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, e)
	}
	return out
}

func findCycle(nodeIDs []string, edges []GraphEdge) ([]string, bool) {
	adj := make(map[string][]string, len(nodeIDs))
	known := make(map[string]struct{}, len(nodeIDs))
	for _, id := range nodeIDs {
		known[id] = struct{}{}
	}
	for _, e := range edges {
		if !cycleFormingEdgeTypes[e.Type] {
			continue
		}
		if _, ok := known[e.From]; !ok {
			continue
		}
		if _, ok := known[e.To]; !ok {
			continue
		}
		adj[e.From] = append(adj[e.From], e.To)
	}

	state := make(map[string]int, len(nodeIDs)) // 0=unseen,1=stack,2=done
	stack := make([]string, 0)
	var cycle []string

	var visit func(v string) bool
	visit = func(v string) bool {
		switch state[v] {
		case 1:
			for i, n := range stack {
				if n == v {
					cycle = append(append([]string{}, stack[i:]...), v)
					return true
				}
			}
			cycle = append(append([]string{}, stack...), v)
			return true
		case 2:
			return false
		}
		state[v] = 1
		stack = append(stack, v)
		for _, w := range adj[v] {
			if visit(w) {
				return true
			}
		}
		stack = stack[:len(stack)-1]
		state[v] = 2
		return false
	}

	for _, id := range nodeIDs {
		if state[id] == 0 && visit(id) {
			return cycle, true
		}
	}
	return nil, false
}

func productDir(repoRoot, productID string) string {
	return filepath.Join(repoRoot, productsRel, product.Slug(productID))
}

func resolveFlowPath(productDir, flowID string) string {
	if flowID == "" {
		return ""
	}
	candidates := []string{
		filepath.Join(productDir, "flows", flowID+".flow.yaml"),
		filepath.Join(productDir, "flows", flowID),
	}
	for _, p := range candidates {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return candidates[0]
}
