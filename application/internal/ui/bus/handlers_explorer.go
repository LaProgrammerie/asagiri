package bus

import (
	"context"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/replay"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

func (b *queryBus) handleGetGraphView(_ context.Context, q GetGraphViewQuery) (QueryResult, error) {
	graph, warning := latestFlowGraph(b.deps.RepoRoot, q.FlowID)
	if graph == nil {
		return GraphViewResult{
			FlowID:  q.FlowID,
			View:    normalizeGraphView(q.View),
			Nodes:   []GraphNodeSummary{},
			Warning: warning,
		}, nil
	}
	if q.GraphID != "" && graph.ID != q.GraphID {
		loaded, loadWarning := loadGraphByID(b.deps.RepoRoot, q.GraphID)
		if loaded != nil {
			graph = loaded
			warning = loadWarning
		}
	}
	view := normalizeGraphView(q.View)
	nodes, groups := projectGraphView(*graph, view)
	return GraphViewResult{
		GraphID: graph.ID,
		FlowID:  graph.Flow,
		View:    view,
		Nodes:   nodes,
		Groups:  groups,
		Warning: warning,
	}, nil
}

func (b *queryBus) handleGetGraphNodeDetail(_ context.Context, q GetGraphNodeDetailQuery) (QueryResult, error) {
	graphID := strings.TrimSpace(q.GraphID)
	nodeID := strings.TrimSpace(q.NodeID)
	if nodeID == "" {
		return GraphNodeDetail{}, fmt.Errorf("graph node detail: node id required")
	}
	graph, _ := loadGraphByID(b.deps.RepoRoot, graphID)
	if graph == nil {
		graph, _ = latestFlowGraph(b.deps.RepoRoot, "")
	}
	if graph == nil {
		return GraphNodeDetail{NodeID: nodeID, CLIEquivalent: "asa graph status <graph-id>"}, nil
	}
	var node *executiongraph.GraphNode
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == nodeID {
			node = &graph.Nodes[i]
			break
		}
	}
	if node == nil {
		return GraphNodeDetail{
			GraphID:       graph.ID,
			NodeID:        nodeID,
			Title:         nodeID,
			CLIEquivalent: fmt.Sprintf("asa graph status %s", graph.ID),
		}, nil
	}
	deps, dependents, blockers := graphAdjacency(*graph, nodeID)
	return GraphNodeDetail{
		GraphID:       graph.ID,
		NodeID:        node.ID,
		Title:         emptyFallback(node.Title, node.ID),
		Status:        string(node.Status),
		Risk:          emptyFallback(string(node.Risk), "unknown"),
		Type:          string(node.Type),
		Dependencies:  deps,
		Dependents:    dependents,
		BlockedBy:     blockers,
		LogsHint:      fmt.Sprintf("runtime events for node %s", node.ID),
		CLIEquivalent: fmt.Sprintf("asa graph status %s", graph.ID),
	}, nil
}

func (b *queryBus) handleGetFlowStepDetail(_ context.Context, q GetFlowStepDetailQuery) (QueryResult, error) {
	flowAny, _ := b.handleGetFlowExplorer(context.Background(), GetFlowExplorerQuery{FlowID: q.FlowID})
	flowRes, _ := flowAny.(FlowExplorerResult)
	stepID := strings.TrimSpace(q.StepID)
	if stepID == "" && len(flowRes.Steps) > 0 {
		stepID = flowRes.Steps[0].ID
	}
	for _, step := range flowRes.Steps {
		if step.ID == stepID {
			return FlowStepDetailResult{FlowID: flowRes.FlowID, Step: step}, nil
		}
	}
	return FlowStepDetailResult{
		FlowID: flowRes.FlowID,
		Step: FlowStepDetail{
			ID:    stepID,
			Label: stepID,
		},
	}, nil
}

func (b *queryBus) handleGetKnowledgeMatchDetail(ctx context.Context, q GetKnowledgeMatchDetailQuery) (QueryResult, error) {
	matchID := strings.TrimSpace(q.MatchID)
	if matchID == "" {
		return KnowledgeMatchDetail{}, fmt.Errorf("knowledge match detail: match id required")
	}
	store, err := knowledge.OpenStoreIfExists(b.deps.RepoRoot)
	if err != nil {
		return KnowledgeMatchDetail{
			MatchID: matchID,
			Name:    matchID,
		}, nil
	}
	defer store.Close()
	graph, err := store.LoadGraph(ctx)
	if err != nil {
		return KnowledgeMatchDetail{MatchID: matchID, Name: matchID}, nil
	}
	var node *knowledge.GraphNode
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == matchID {
			node = &graph.Nodes[i]
			break
		}
	}
	if node == nil {
		return KnowledgeMatchDetail{
			MatchID:       matchID,
			Name:          matchID,
			CLIEquivalent: `asa knowledge query "` + matchID + `"`,
		}, nil
	}
	flows, apis, tests, events := relatedKnowledgeEntities(graph, matchID)
	return KnowledgeMatchDetail{
		MatchID:       node.ID,
		Name:          emptyFallback(node.Name, node.ID),
		Type:          string(node.Type),
		Path:          emptyFallback(node.Path, "-"),
		RelatedFlows:  flows,
		RelatedAPIs:   apis,
		RelatedTests:  tests,
		RelatedEvents: events,
		CLIEquivalent: `asa knowledge query "` + node.ID + `"`,
	}, nil
}

func (b *queryBus) handleGetTrustDimensionDetail(_ context.Context, q GetTrustDimensionDetailQuery) (QueryResult, error) {
	label := strings.TrimSpace(q.Label)
	report, _ := latestTrustReport(b.deps.RepoRoot)
	if report == nil {
		return TrustDimensionDetail{
			Label:        label,
			GateStatus:   "unknown",
			ResidualRisk: "unknown",
		}, nil
	}
	kind := trustCheckTypeForLabel(label)
	findings := findingsForCheckType(report.Checks, kind)
	evidence := evidenceForCheckType(report.Checks, kind)
	checks := checkSummariesForType(report.Checks, kind)
	return TrustDimensionDetail{
		Label:         label,
		Score:         trustScoreForLabel(*report, label),
		Findings:      findings,
		Evidence:      evidence,
		Checks:        checks,
		GateStatus:    string(report.Gate.Status),
		GateReason:    report.Gate.Reason,
		ResidualRisk:  string(report.ResidualRisk),
		CLIEquivalent: "asa verify trust <flow>",
	}, nil
}

func (b *queryBus) handleGetReplayEventDetail(_ context.Context, q GetReplayEventDetailQuery) (QueryResult, error) {
	replayID, _ := b.resolveReplayID(q.ReplayID)
	if replayID == "" {
		return ReplayEventDetail{ReplayID: q.ReplayID, Index: q.Index}, nil
	}
	pkgAny, _ := b.handleGetReplayPackage(context.Background(), GetReplayPackageQuery{ReplayID: replayID, Limit: 500})
	pkgRes, _ := pkgAny.(ReplayPackageResult)
	if q.Index < 0 || q.Index >= len(pkgRes.Timeline) {
		return ReplayEventDetail{
			ReplayID:      replayID,
			Index:         q.Index,
			CLIEquivalent: "asa replay open " + replayID,
		}, nil
	}
	ev := pkgRes.Timeline[q.Index]
	artifactPath := ""
	if ev.Artifact != "" {
		artifactPath = filepath.Join(replay.RelDir, replayID, ev.Artifact)
	}
	return ReplayEventDetail{
		ReplayID:      replayID,
		Index:         q.Index,
		Type:          ev.Type,
		Time:          ev.Time,
		Artifact:      ev.Artifact,
		ArtifactPath:  artifactPath,
		CLIEquivalent: "asa replay open " + replayID,
	}, nil
}

func (b *queryBus) handleGetReplayCompare(_ context.Context, q GetReplayCompareQuery) (QueryResult, error) {
	replayA := strings.TrimSpace(q.ReplayA)
	replayB := strings.TrimSpace(q.ReplayB)
	if replayA == "" || replayB == "" {
		return ReplayCompareResult{
			ReplayA:       replayA,
			ReplayB:       replayB,
			Warning:       "both replay ids required",
			CLIEquivalent: "asa replay compare <replay-a> <replay-b>",
		}, nil
	}
	cmp, err := replay.NewComparator(b.deps.RepoRoot).Compare(context.Background(), replayA, replayB)
	if err != nil {
		return ReplayCompareResult{
			ReplayA:       replayA,
			ReplayB:       replayB,
			Warning:       err.Error(),
			CLIEquivalent: "asa replay compare " + replayA + " " + replayB,
		}, nil
	}
	summary := []string{
		fmt.Sprintf("cost delta: %.4f EUR", cmp.CostDelta),
	}
	if cmp.DurationDelta != "" {
		summary = append(summary, "duration delta: "+cmp.DurationDelta)
	}
	divs := cmp.Differences
	if len(divs) == 0 {
		divs = replay.ExplainDivergences(cmp.Divergences)
	}
	if len(divs) == 0 {
		divs = []string{"no divergence detected"}
	}
	return ReplayCompareResult{
		ReplayA:       replayA,
		ReplayB:       replayB,
		Summary:       summary,
		Divergences:   divs,
		CLIEquivalent: "asa replay compare " + replayA + " " + replayB,
	}, nil
}

func normalizeGraphView(view GraphViewMode) GraphViewMode {
	v := GraphViewMode(strings.TrimSpace(string(view)))
	for _, candidate := range GraphViewModes {
		if candidate == v {
			return candidate
		}
	}
	return GraphViewTimeline
}

func loadGraphByID(repoRoot, graphID string) (*executiongraph.ExecutionGraph, string) {
	id := strings.TrimSpace(graphID)
	if id == "" {
		return nil, ""
	}
	repo := executiongraph.NewRepository(repoRoot)
	graph, err := repo.Load(id)
	if err != nil {
		return nil, err.Error()
	}
	return &graph, ""
}

func projectGraphView(graph executiongraph.ExecutionGraph, view GraphViewMode) ([]GraphNodeSummary, []string) {
	blockers := map[string][]string{}
	for _, edge := range graph.Edges {
		blockers[edge.To] = append(blockers[edge.To], edge.From)
	}
	all := make([]GraphNodeSummary, 0, len(graph.Nodes))
	for _, node := range graph.SortedNodes() {
		all = append(all, GraphNodeSummary{
			ID:            node.ID,
			Title:         node.Title,
			Type:          string(node.Type),
			Status:        string(node.Status),
			Risk:          emptyFallback(string(node.Risk), "unknown"),
			BlockedBy:     blockers[node.ID],
			CLIEquivalent: fmt.Sprintf("asa graph status %s", graph.ID),
		})
	}
	switch view {
	case GraphViewBlocked:
		filtered := make([]GraphNodeSummary, 0, len(all))
		for _, node := range all {
			if node.Status == "blocked" || len(node.BlockedBy) > 0 {
				filtered = append(filtered, node)
			}
		}
		return filtered, nil
	case GraphViewCriticalPath:
		path := longestDependencyPath(graph)
		if len(path) == 0 {
			return all, nil
		}
		set := map[string]struct{}{}
		for _, id := range path {
			set[id] = struct{}{}
		}
		filtered := make([]GraphNodeSummary, 0, len(path))
		for _, node := range all {
			if _, ok := set[node.ID]; ok {
				filtered = append(filtered, node)
			}
		}
		return filtered, []string{strings.Join(path, " → ")}
	case GraphViewParallelGroups:
		groups := parallelGroups(graph)
		groupLines := make([]string, 0, len(groups))
		filtered := make([]GraphNodeSummary, 0, len(all))
		for i, group := range groups {
			groupLines = append(groupLines, fmt.Sprintf("group %d: %s", i+1, strings.Join(group, ", ")))
			set := map[string]struct{}{}
			for _, id := range group {
				set[id] = struct{}{}
			}
			for _, node := range all {
				if _, ok := set[node.ID]; ok {
					filtered = append(filtered, node)
				}
			}
		}
		return filtered, groupLines
	case GraphViewDependency:
		return all, dependencyLines(graph)
	default:
		return all, nil
	}
}

func graphAdjacency(graph executiongraph.ExecutionGraph, nodeID string) (deps, dependents, blockers []string) {
	for _, edge := range graph.Edges {
		if edge.To == nodeID {
			deps = append(deps, edge.From)
		}
		if edge.From == nodeID {
			dependents = append(dependents, edge.To)
		}
	}
	sort.Strings(deps)
	sort.Strings(dependents)
	for _, dep := range deps {
		for _, node := range graph.Nodes {
			if node.ID == dep && (node.Status == "blocked" || node.Status == "failed") {
				blockers = append(blockers, dep)
			}
		}
	}
	return deps, dependents, blockers
}

func dependencyLines(graph executiongraph.ExecutionGraph) []string {
	lines := make([]string, 0, len(graph.Edges))
	for _, edge := range graph.SortedEdges() {
		lines = append(lines, edge.From+" → "+edge.To)
	}
	return lines
}

func longestDependencyPath(graph executiongraph.ExecutionGraph) []string {
	incoming := map[string][]string{}
	for _, edge := range graph.Edges {
		incoming[edge.To] = append(incoming[edge.To], edge.From)
	}
	memo := map[string][]string{}
	var walk func(id string) []string
	walk = func(id string) []string {
		if cached, ok := memo[id]; ok {
			return cached
		}
		best := []string{id}
		for _, parent := range incoming[id] {
			path := append(walk(parent), id)
			if len(path) > len(best) {
				best = path
			}
		}
		memo[id] = best
		return best
	}
	longest := []string{}
	for _, node := range graph.Nodes {
		path := walk(node.ID)
		if len(path) > len(longest) {
			longest = path
		}
	}
	return longest
}

func parallelGroups(graph executiongraph.ExecutionGraph) [][]string {
	level := map[string]int{}
	for _, node := range graph.SortedNodes() {
		if _, ok := level[node.ID]; !ok {
			level[node.ID] = 0
		}
		for _, edge := range graph.Edges {
			if edge.From == node.ID {
				next := level[node.ID] + 1
				if level[edge.To] < next {
					level[edge.To] = next
				}
			}
		}
	}
	maxLevel := 0
	for _, v := range level {
		if v > maxLevel {
			maxLevel = v
		}
	}
	groups := make([][]string, maxLevel+1)
	for id, lv := range level {
		groups[lv] = append(groups[lv], id)
	}
	for i := range groups {
		sort.Strings(groups[i])
	}
	return groups
}

func relatedKnowledgeEntities(graph knowledge.KnowledgeGraph, nodeID string) (flows, apis, tests, events []string) {
	seen := map[string]struct{}{}
	for _, edge := range graph.Edges {
		peer := ""
		if edge.From == nodeID {
			peer = edge.To
		} else if edge.To == nodeID {
			peer = edge.From
		}
		if peer == "" {
			continue
		}
		if _, ok := seen[peer]; ok {
			continue
		}
		seen[peer] = struct{}{}
		peerNode := graphNodeByID(graph, peer)
		if peerNode == nil {
			continue
		}
		switch peerNode.Type {
		case knowledge.NodeTypeFlow, knowledge.NodeTypeFlowStep:
			flows = append(flows, peerNode.Name)
		case knowledge.NodeTypeAPIOperation:
			apis = append(apis, peerNode.Name)
		case knowledge.NodeTypeTest:
			tests = append(tests, peerNode.Name)
		case knowledge.NodeTypeEvent:
			events = append(events, peerNode.Name)
		}
	}
	sort.Strings(flows)
	sort.Strings(apis)
	sort.Strings(tests)
	sort.Strings(events)
	return flows, apis, tests, events
}

func graphNodeByID(graph knowledge.KnowledgeGraph, id string) *knowledge.GraphNode {
	for i := range graph.Nodes {
		if graph.Nodes[i].ID == id {
			return &graph.Nodes[i]
		}
	}
	return nil
}

func trustCheckTypeForLabel(label string) trust.CheckType {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "architecture":
		return trust.CheckArchitecture
	case "implementation":
		return trust.CheckStaticAnalysis
	case "security":
		return trust.CheckSecurity
	case "observability":
		return trust.CheckObservability
	case "regression":
		return trust.CheckBackwardCompatibility
	default:
		return trust.CheckArchitecture
	}
}

func trustScoreForLabel(report trust.TrustReport, label string) float64 {
	switch strings.ToLower(strings.TrimSpace(label)) {
	case "architecture":
		return report.Confidence.Architecture
	case "implementation":
		return report.Confidence.Implementation
	case "security":
		return report.Confidence.Security
	case "observability":
		return report.Confidence.Observability
	case "regression":
		return report.Confidence.Regression
	default:
		return report.Confidence.Overall
	}
}

func checkSummariesForType(checks []trust.VerificationCheck, kind trust.CheckType) []string {
	out := make([]string, 0, 4)
	for _, check := range checks {
		if check.Type != kind {
			continue
		}
		status := string(check.Status)
		if status == "" {
			status = "unknown"
		}
		out = append(out, fmt.Sprintf("%s (%s)", check.Type, status))
		if len(out) >= 4 {
			break
		}
	}
	return out
}
