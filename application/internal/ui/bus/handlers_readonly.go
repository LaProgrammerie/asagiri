package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"github.com/LaProgrammerie/asagiri/application/internal/replay"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"gopkg.in/yaml.v3"
)

func (b *queryBus) handleGetRuntimeStatus(ctx context.Context, _ GetRuntimeStatusQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store, err := b.deps.RuntimeOpen(b.deps.RepoRoot)
	if err != nil {
		return RuntimeStatusResult{
			Status:  runtime.DaemonStatus{},
			Warning: err.Error(),
		}, nil
	}
	defer func() { _ = store.Close() }()

	status, err := store.Status()
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return RuntimeStatusResult{Status: status}, nil
}

func (b *queryBus) handleListRuns(ctx context.Context, q ListRunsQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store, err := b.deps.StateOpen(b.deps.StateDBPath)
	if err != nil {
		return ListRunsResult{
			Runs:    []RunSummary{},
			Warning: err.Error(),
		}, nil
	}
	defer func() { _ = store.Close() }()

	if err := store.Migrate(); err != nil {
		return ListRunsResult{
			Runs:    []RunSummary{},
			Warning: err.Error(),
		}, nil
	}

	rows, err := store.ListRuns(q.Limit)
	if err != nil {
		return ListRunsResult{
			Runs:    []RunSummary{},
			Warning: err.Error(),
		}, nil
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]RunSummary, 0, len(rows))
	for _, run := range rows {
		out = append(out, RunSummary{
			ID:        run.ID,
			Feature:   run.Feature,
			Status:    run.Status,
			CreatedAt: run.CreatedAt,
			UpdatedAt: run.UpdatedAt,
		})
	}
	return ListRunsResult{Runs: out}, nil
}

func (b *queryBus) handleGetRecentEvents(ctx context.Context, q GetRecentEventsQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store, err := b.deps.RuntimeOpen(b.deps.RepoRoot)
	if err != nil {
		return RecentEventsResult{}, nil
	}
	defer func() { _ = store.Close() }()

	rows, err := store.ListEvents(q.Limit)
	if err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	out := make([]EventSummary, 0, len(rows))
	for _, ev := range rows {
		out = append(out, EventSummary{
			ID:        ev.ID,
			Type:      ev.Type,
			Source:    ev.Source,
			SessionID: ev.SessionID,
			FlowID:    ev.FlowID,
			CreatedAt: ev.CreatedAt,
			Payload:   ev.Payload,
		})
	}
	return RecentEventsResult{Events: out}, nil
}

func (b *queryBus) handleGetTrustSummary(_ context.Context, _ GetTrustSummaryQuery) (QueryResult, error) {
	report, warning := latestTrustReport(b.deps.RepoRoot)
	if report == nil {
		return TrustSummaryResult{
			Dimensions: []TrustDimensionScore{},
			Warning:    warning,
		}, nil
	}
	generatedAt, _ := time.Parse(time.RFC3339Nano, report.GeneratedAt)
	dimensions := []TrustDimensionScore{
		{Label: "Architecture", Score: report.Confidence.Architecture},
		{Label: "Security", Score: report.Confidence.Security},
		{Label: "Observability", Score: report.Confidence.Observability},
		{Label: "Regression", Score: report.Confidence.Regression},
	}
	return TrustSummaryResult{
		Overall:     report.Confidence.Overall,
		Dimensions:  dimensions,
		GeneratedAt: generatedAt,
		Warning:     warning,
	}, nil
}

func (b *queryBus) handleListActiveAgents(ctx context.Context, q ListActiveAgentsQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store, err := b.deps.RuntimeOpen(b.deps.RepoRoot)
	if err != nil {
		return ActiveAgentsResult{
			Agents:  []ActiveAgentSummary{},
			Warning: err.Error(),
		}, nil
	}
	defer func() { _ = store.Close() }()
	limit := q.Limit
	if limit <= 0 {
		limit = 200
	}
	events, err := store.ListEvents(limit)
	if err != nil {
		return ActiveAgentsResult{
			Agents:  []ActiveAgentSummary{},
			Warning: err.Error(),
		}, nil
	}

	agentsByKey := map[string]ActiveAgentSummary{}
	for _, ev := range events {
		if !strings.HasPrefix(ev.Type, "agent.") {
			continue
		}
		role := payloadString(ev.Payload, "role")
		agentRef := payloadString(ev.Payload, "agent_ref")
		if agentRef == "" {
			agentRef = payloadString(ev.Payload, "agent")
		}
		key := role + "|" + agentRef
		if key == "|" {
			key = ev.Type + "|" + ev.Source
		}
		existing, ok := agentsByKey[key]
		if ok && existing.UpdatedAt.After(ev.CreatedAt) {
			continue
		}
		agentsByKey[key] = ActiveAgentSummary{
			Role:      role,
			AgentRef:  agentRef,
			Status:    agentStatusFromEvent(ev.Type),
			FlowID:    ev.FlowID,
			UpdatedAt: ev.CreatedAt,
		}
	}

	agents := make([]ActiveAgentSummary, 0, len(agentsByKey))
	for _, ag := range agentsByKey {
		agents = append(agents, ag)
	}
	sort.Slice(agents, func(i, j int) bool {
		if agents[i].Status != agents[j].Status {
			return agents[i].Status == "running"
		}
		return agents[i].UpdatedAt.After(agents[j].UpdatedAt)
	})
	return ActiveAgentsResult{Agents: agents}, nil
}

func (b *queryBus) handleGetFlowGraph(_ context.Context, q GetFlowGraphQuery) (QueryResult, error) {
	graph, warning := latestFlowGraph(b.deps.RepoRoot, q.FlowID)
	if graph == nil {
		return FlowGraphResult{
			FlowID:  q.FlowID,
			Steps:   []FlowGraphStep{},
			Warning: warning,
		}, nil
	}
	steps := make([]FlowGraphStep, 0, len(graph.Nodes))
	for _, node := range graph.SortedNodes() {
		steps = append(steps, FlowGraphStep{
			ID:     node.ID,
			Label:  node.Title,
			Status: string(node.Status),
		})
	}
	return FlowGraphResult{
		FlowID:  graph.Flow,
		Steps:   steps,
		Warning: warning,
	}, nil
}

func (b *queryBus) handleGetFlowExplorer(_ context.Context, q GetFlowExplorerQuery) (QueryResult, error) {
	graph, warning := latestFlowGraph(b.deps.RepoRoot, q.FlowID)
	if graph == nil {
		return FlowExplorerResult{
			FlowID:  q.FlowID,
			Steps:   []FlowStepDetail{},
			Warning: warning,
		}, nil
	}
	steps := make([]FlowStepDetail, 0, len(graph.Nodes))
	for _, node := range graph.SortedNodes() {
		tests := []string{}
		if node.Task != "" {
			tests = append(tests, node.Task+"Test")
		}
		metrics := []string{}
		if node.Task != "" {
			metrics = append(metrics, node.Task+"_success_rate")
		}
		steps = append(steps, FlowStepDetail{
			ID:         node.ID,
			Label:      node.Title,
			Status:     string(node.Status),
			API:        "n/a",
			Service:    emptyFallback(node.Agent, "n/a"),
			Event:      "n/a",
			Tests:      tests,
			Metrics:    metrics,
			TrustScore: trustScoreFromRisk(node.Risk),
			Risk:       emptyFallback(string(node.Risk), "unknown"),
		})
	}
	selected := ""
	if len(steps) > 0 {
		selected = steps[0].ID
	}
	return FlowExplorerResult{
		FlowID:   graph.Flow,
		Steps:    steps,
		Selected: selected,
		Warning:  warning,
	}, nil
}

func (b *queryBus) handleGetGraphExplorer(_ context.Context, q GetGraphExplorerQuery) (QueryResult, error) {
	graph, warning := latestFlowGraph(b.deps.RepoRoot, q.FlowID)
	if graph == nil {
		return GraphExplorerResult{
			FlowID:  q.FlowID,
			Nodes:   []GraphNodeSummary{},
			Warning: warning,
		}, nil
	}
	blockers := map[string][]string{}
	for _, edge := range graph.Edges {
		blockers[edge.To] = append(blockers[edge.To], edge.From)
	}
	nodes := make([]GraphNodeSummary, 0, len(graph.Nodes))
	for _, node := range graph.SortedNodes() {
		nodes = append(nodes, GraphNodeSummary{
			ID:            node.ID,
			Title:         node.Title,
			Type:          string(node.Type),
			Status:        string(node.Status),
			Risk:          emptyFallback(string(node.Risk), "unknown"),
			BlockedBy:     blockers[node.ID],
			CLIEquivalent: fmt.Sprintf("asa graph status %s", graph.ID),
		})
	}
	return GraphExplorerResult{
		GraphID: graph.ID,
		Product: graph.Product,
		FlowID:  graph.Flow,
		Status:  string(graph.Status),
		Nodes:   nodes,
		Warning: warning,
	}, nil
}

func (b *queryBus) handleSearchKnowledge(ctx context.Context, q SearchKnowledgeQuery) (QueryResult, error) {
	limit := q.Limit
	if limit <= 0 {
		limit = 8
	}
	store, err := knowledge.OpenStoreIfExists(b.deps.RepoRoot)
	if err != nil {
		return KnowledgeSearchResult{
			Query:   q.Query,
			Matches: []KnowledgeMatch{},
			Warning: "knowledge graph unavailable",
		}, nil
	}
	defer func() { _ = store.Close() }()
	graph, err := store.LoadGraph(ctx)
	if err != nil {
		return KnowledgeSearchResult{
			Query:   q.Query,
			Matches: []KnowledgeMatch{},
			Warning: "knowledge graph unreadable",
		}, nil
	}
	query := strings.ToLower(strings.TrimSpace(q.Query))
	matches := make([]KnowledgeMatch, 0, limit)
	if query == "" {
		return KnowledgeSearchResult{Query: q.Query, Matches: matches}, nil
	}
	for _, node := range graph.Nodes {
		if len(matches) >= limit {
			break
		}
		idText := strings.ToLower(node.ID)
		nameText := strings.ToLower(node.Name)
		pathText := strings.ToLower(node.Path)
		if !strings.Contains(idText, query) && !strings.Contains(nameText, query) && !strings.Contains(pathText, query) {
			continue
		}
		matches = append(matches, KnowledgeMatch{
			ID:            node.ID,
			Type:          string(node.Type),
			Name:          emptyFallback(node.Name, "-"),
			Path:          emptyFallback(node.Path, "-"),
			Score:         knowledgeScore(node, query),
			CLIEquivalent: fmt.Sprintf(`asa knowledge query "%s"`, query),
		})
	}
	sort.Slice(matches, func(i, j int) bool { return matches[i].Score > matches[j].Score })
	return KnowledgeSearchResult{
		Query:   q.Query,
		Matches: matches,
	}, nil
}

func (b *queryBus) handleGetTrustExplorer(_ context.Context, _ GetTrustExplorerQuery) (QueryResult, error) {
	report, warning := latestTrustReport(b.deps.RepoRoot)
	if report == nil {
		return TrustExplorerResult{
			Dimensions: []TrustEvidenceDimension{},
			Warnings:   []string{},
			Warning:    warning,
		}, nil
	}
	dimensions := []TrustEvidenceDimension{
		{
			Label:         "Architecture",
			Score:         report.Confidence.Architecture,
			Findings:      findingsForCheckType(report.Checks, trust.CheckArchitecture),
			Evidence:      evidenceForCheckType(report.Checks, trust.CheckArchitecture),
			CLIEquivalent: "asa verify trust <flow>",
		},
		{
			Label:         "Implementation",
			Score:         report.Confidence.Implementation,
			Findings:      findingsForCheckType(report.Checks, trust.CheckStaticAnalysis),
			Evidence:      evidenceForCheckType(report.Checks, trust.CheckStaticAnalysis),
			CLIEquivalent: "asa verify trust <flow>",
		},
		{
			Label:         "Security",
			Score:         report.Confidence.Security,
			Findings:      findingsForCheckType(report.Checks, trust.CheckSecurity),
			Evidence:      evidenceForCheckType(report.Checks, trust.CheckSecurity),
			CLIEquivalent: "asa verify trust <flow> --strict",
		},
		{
			Label:         "Observability",
			Score:         report.Confidence.Observability,
			Findings:      findingsForCheckType(report.Checks, trust.CheckObservability),
			Evidence:      evidenceForCheckType(report.Checks, trust.CheckObservability),
			CLIEquivalent: "asa verify trust <flow>",
		},
		{
			Label:         "Regression",
			Score:         report.Confidence.Regression,
			Findings:      findingsForCheckType(report.Checks, trust.CheckBackwardCompatibility),
			Evidence:      evidenceForCheckType(report.Checks, trust.CheckBackwardCompatibility),
			CLIEquivalent: "asa verify trust <flow>",
		},
	}
	return TrustExplorerResult{
		Overall:      report.Confidence.Overall,
		ResidualRisk: string(report.ResidualRisk),
		GateStatus:   string(report.Gate.Status),
		GateReason:   report.Gate.Reason,
		Dimensions:   dimensions,
		Warnings:     append([]string(nil), report.Warnings...),
		Warning:      warning,
	}, nil
}

func (b *queryBus) handleGetExplain(ctx context.Context, q GetExplainQuery) (QueryResult, error) {
	question := explainQuestionForContext(q.Context, q.Subject)
	subject := strings.TrimSpace(q.Subject)
	if subject == "" {
		subject = strings.TrimSpace(q.Context.Focus.Subject)
	}
	if subject == "" {
		subject = "current decision"
	}

	graphAny, _ := b.handleGetGraphExplorer(ctx, GetGraphExplorerQuery{FlowID: q.Context.Focus.Detail})
	graphRes, _ := graphAny.(GraphExplorerResult)
	trustAny, _ := b.handleGetTrustExplorer(ctx, GetTrustExplorerQuery{})
	trustRes, _ := trustAny.(TrustExplorerResult)
	knowledgeAny, _ := b.handleSearchKnowledge(ctx, SearchKnowledgeQuery{Query: subject, Limit: 3})
	knowledgeRes, _ := knowledgeAny.(KnowledgeSearchResult)

	reasons := explainReasonsForQuestion(question, graphRes, trustRes, q.Context)
	evidence := explainEvidenceForQuestion(question, graphRes, trustRes, knowledgeRes, q.Context)
	alternatives := explainAlternativesForQuestion(question, q.Context)

	return ExplainResult{
		Subject:            subject,
		Question:           question,
		SupportedQuestions: explainSupportedQuestions(),
		Reasons:            reasons,
		Evidence:           evidence,
		Source:             explainSourceForContext(q.Context),
		Alternatives:       alternatives,
		CLIEquivalent:      fmt.Sprintf(`asa explain --subject "%s"`, question),
	}, nil
}

func (b *queryBus) handleGetRecommendedActions(ctx context.Context, q GetRecommendedActionsQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	runtimeAny, _ := b.handleGetRuntimeStatus(ctx, GetRuntimeStatusQuery{})
	runtimeRes, _ := runtimeAny.(RuntimeStatusResult)
	trustAny, _ := b.handleGetTrustSummary(ctx, GetTrustSummaryQuery{})
	trustRes, _ := trustAny.(TrustSummaryResult)
	graphAny, _ := b.handleGetGraphExplorer(ctx, GetGraphExplorerQuery(q))
	graphRes, _ := graphAny.(GraphExplorerResult)
	flowAny, _ := b.handleGetFlowGraph(ctx, GetFlowGraphQuery{FlowID: q.FlowID})
	flowRes, _ := flowAny.(FlowGraphResult)
	prototypeAny, _ := b.handleGetPrototypePipeline(ctx, GetPrototypePipelineQuery{Limit: 1})
	prototypeRes, _ := prototypeAny.(PrototypePipelineResult)

	actions := make([]RecommendedAction, 0, 8)
	if !runtimeRes.Status.Running {
		actions = append(actions, RecommendedAction{
			ID:            "rec.start-work",
			Title:         "Start work",
			Description:   "Runtime is stopped — dispatch a new workflow run",
			Priority:      1,
			CLIEquivalent: `asa work "<intent>"`,
			ActionID:      "cmd.start-work",
		})
	}
	if runtimeRes.Status.QueuedEvents >= 3 {
		actions = append(actions, RecommendedAction{
			ID:            "rec.export-events",
			Title:         "Export queued events",
			Description:   fmt.Sprintf("Queue has %d pending events", runtimeRes.Status.QueuedEvents),
			Priority:      2,
			CLIEquivalent: "asa runtime events --export",
			ActionID:      "cmd.export-events",
		})
	}
	if trustRes.Overall > 0 && trustRes.Overall < 0.75 {
		target := emptyFallback(flowRes.FlowID, "onboarding")
		actions = append(actions, RecommendedAction{
			ID:            "rec.verify-trust",
			Title:         "Verify trust",
			Description:   fmt.Sprintf("Overall trust is %.0f%% — run verification", trustRes.Overall*100),
			Priority:      2,
			CLIEquivalent: "asa verify trust " + target,
			ActionID:      "cmd.verify-trust",
		})
	}
	for _, dim := range trustRes.Dimensions {
		if strings.EqualFold(dim.Label, "Security") && dim.Score > 0 && dim.Score < 0.75 {
			actions = append(actions, RecommendedAction{
				ID:            "rec.explain-security",
				Title:         "Explain security confidence",
				Description:   fmt.Sprintf("Security confidence is %.0f%%", dim.Score*100),
				Priority:      2,
				CLIEquivalent: `asa explain --subject "Why is security confidence low?"`,
				ActionID:      "nav.explain",
			})
			break
		}
	}
	for _, node := range graphRes.Nodes {
		if node.Status == "blocked" || len(node.BlockedBy) > 0 {
			actions = append(actions, RecommendedAction{
				ID:            "rec.graph-resume",
				Title:         "Resume blocked graph",
				Description:   fmt.Sprintf("Node %s is blocked", emptyFallback(node.Title, node.ID)),
				Priority:      1,
				CLIEquivalent: "asa graph resume " + emptyFallback(graphRes.GraphID, "<graph-id>"),
				ActionID:      "ctx.graph-resume",
			})
			actions = append(actions, RecommendedAction{
				ID:            "rec.explain-blocked",
				Title:         "Explain blocked node",
				Description:   "Review why execution is waiting on dependencies",
				Priority:      3,
				CLIEquivalent: `asa explain --subject "Why is this node blocked?"`,
				ActionID:      "nav.explain",
			})
			break
		}
	}
	for _, step := range flowRes.Steps {
		if step.Status == "failed" || step.Status == "blocked" {
			actions = append(actions, RecommendedAction{
				ID:            "rec.investigate-flow",
				Title:         "Investigate flow failure",
				Description:   fmt.Sprintf("Step %s is %s", emptyFallback(step.Label, step.ID), step.Status),
				Priority:      1,
				CLIEquivalent: `asa investigate "` + emptyFallback(step.Label, step.ID) + `"`,
				ActionID:      "cmd.run-investigation",
			})
			break
		}
	}
	if prototypeRes.Product == "" {
		actions = append(actions, RecommendedAction{
			ID:            "rec.prototype-create",
			Title:         "Create prototype",
			Description:   "No product prototype found — start product pipeline",
			Priority:      4,
			CLIEquivalent: `asa prototype create "<intent>"`,
			ActionID:      "cmd.prototype-create",
		})
	} else if len(prototypeRes.SuggestedActions) > 0 {
		actions = append(actions, RecommendedAction{
			ID:            "rec.prototype-next",
			Title:         "Advance prototype pipeline",
			Description:   prototypeRes.SuggestedActions[0],
			Priority:      3,
			CLIEquivalent: prototypeRes.SuggestedActions[0],
			ActionID:      "cmd.prototype-pipeline",
		})
	}
	if len(actions) == 0 {
		actions = append(actions, RecommendedAction{
			ID:            "rec.dashboard",
			Title:         "Open dashboard",
			Description:   "Review live widgets for runtime, trust, and costs",
			Priority:      5,
			CLIEquivalent: "asa dashboard",
			ActionID:      "nav.dashboard",
		})
	}
	sort.Slice(actions, func(i, j int) bool {
		if actions[i].Priority != actions[j].Priority {
			return actions[i].Priority < actions[j].Priority
		}
		return actions[i].Title < actions[j].Title
	})
	if len(actions) > 6 {
		actions = actions[:6]
	}
	return RecommendedActionsResult{Actions: actions}, nil
}

func explainSupportedQuestions() []string {
	return []string{
		"Why was review required?",
		"Why is this node blocked?",
		"Why is security confidence low?",
		"Why was this agent selected?",
		"Why is this flow high risk?",
		"Why did Asagiri insert investigation?",
	}
}

func explainQuestionForContext(ctx ExplainContext, subject string) string {
	if q := strings.TrimSpace(ctx.Question); q != "" {
		return q
	}
	switch ctx.Focus.Kind {
	case FocusKindGraphNode:
		return "Why is this node blocked?"
	case FocusKindFlowStep:
		return "Why is this flow high risk?"
	case FocusKindTrustDimension:
		label := strings.TrimSpace(ctx.Focus.Subject)
		if strings.EqualFold(label, "Security") {
			return "Why is security confidence low?"
		}
		if label != "" {
			return "Why is " + strings.ToLower(label) + " confidence low?"
		}
		return "Why is security confidence low?"
	case FocusKindAgent:
		return "Why was this agent selected?"
	case FocusKindReplayEvent:
		return "Why did Asagiri insert investigation?"
	default:
		if s := strings.TrimSpace(subject); s != "" && strings.Contains(strings.ToLower(s), "review") {
			return "Why was review required?"
		}
		return "Why was review required?"
	}
}

func explainReasonsForQuestion(question string, graph GraphExplorerResult, trust TrustExplorerResult, ctx ExplainContext) []string {
	q := strings.ToLower(question)
	reasons := make([]string, 0, 4)
	switch {
	case strings.Contains(q, "blocked"):
		for _, node := range graph.Nodes {
			if node.Status == "blocked" || len(node.BlockedBy) > 0 {
				reasons = append(reasons, fmt.Sprintf("Node %s is blocked by %s", emptyFallback(node.Title, node.ID), strings.Join(node.BlockedBy, ", ")))
				break
			}
		}
		if len(reasons) == 0 {
			reasons = append(reasons, fmt.Sprintf("Graph status is %s for flow %s", emptyFallback(graph.Status, "unknown"), emptyFallback(graph.FlowID, "-")))
		}
	case strings.Contains(q, "security"):
		for _, dim := range trust.Dimensions {
			if strings.EqualFold(dim.Label, "Security") {
				reasons = append(reasons, fmt.Sprintf("Security confidence is %.0f%%", dim.Score*100))
				if len(dim.Findings) > 0 {
					reasons = append(reasons, dim.Findings[0])
				}
				break
			}
		}
	case strings.Contains(q, "agent"):
		if ctx.Focus.Subject != "" {
			reasons = append(reasons, "Agent "+ctx.Focus.Subject+" matched routing policy for the active step")
		} else {
			reasons = append(reasons, "Agent selection follows coordination policy and role assignment")
		}
	case strings.Contains(q, "investigation"):
		reasons = append(reasons, "Investigation inserted when trust or dependency gates require more evidence")
	case strings.Contains(q, "risk"):
		reasons = append(reasons, "Flow risk aggregates sensitive steps, unresolved contracts, and trust findings")
	default:
		reasons = append(reasons, fmt.Sprintf("Graph status is %s for flow %s", emptyFallback(graph.Status, "unknown"), emptyFallback(graph.FlowID, "-")))
		reasons = append(reasons, fmt.Sprintf("Trust overall is %.0f%%", trust.Overall*100))
	}
	if len(reasons) == 0 {
		reasons = append(reasons, "Decision driven by runtime state, trust gates, and graph dependencies")
	}
	return reasons
}

func explainEvidenceForQuestion(question string, graph GraphExplorerResult, trust TrustExplorerResult, knowledge KnowledgeSearchResult, ctx ExplainContext) []string {
	q := strings.ToLower(question)
	evidence := make([]string, 0, 4)
	if ctx.Focus.Subject != "" {
		evidence = append(evidence, "Focus: "+string(ctx.Focus.Kind)+" "+ctx.Focus.Subject)
	}
	if strings.Contains(q, "blocked") {
		for _, node := range graph.Nodes {
			if node.Status == "blocked" || len(node.BlockedBy) > 0 {
				evidence = append(evidence, "Node: "+emptyFallback(node.Title, node.ID))
				if len(node.BlockedBy) > 0 {
					evidence = append(evidence, "Blocked by: "+strings.Join(node.BlockedBy, ", "))
				}
				break
			}
		}
	}
	for _, dim := range trust.Dimensions {
		if len(dim.Evidence) == 0 && len(dim.Findings) == 0 {
			continue
		}
		if strings.Contains(q, "security") && !strings.EqualFold(dim.Label, "Security") {
			continue
		}
		if len(dim.Evidence) > 0 {
			evidence = append(evidence, dim.Label+": "+dim.Evidence[0])
		} else if len(dim.Findings) > 0 {
			evidence = append(evidence, dim.Label+": "+dim.Findings[0])
		}
		if len(evidence) >= 3 {
			break
		}
	}
	if len(graph.Nodes) > 0 && len(evidence) < 3 {
		evidence = append(evidence, "Node: "+emptyFallback(graph.Nodes[0].Title, graph.Nodes[0].ID))
	}
	if len(knowledge.Matches) > 0 {
		evidence = append(evidence, "Knowledge: "+knowledge.Matches[0].Name)
	}
	if len(evidence) == 0 {
		evidence = append(evidence, "No additional evidence found")
	}
	return evidence
}

func explainAlternativesForQuestion(question string, ctx ExplainContext) []string {
	q := strings.ToLower(question)
	switch {
	case strings.Contains(q, "blocked"):
		return []string{"asa graph", "asa graph resume <graph-id>", "asa logs"}
	case strings.Contains(q, "security"), strings.Contains(q, "confidence"):
		return []string{"asa trust", "asa verify trust <flow>", "asa knowledge query"}
	case strings.Contains(q, "agent"):
		return []string{"asa agents watch", "asa coordination status"}
	case strings.Contains(q, "investigation"):
		return []string{"asa investigate", "asa replay open <replay-id>"}
	case strings.Contains(q, "risk"):
		return []string{"asa flow", "asa impact analyze"}
	default:
		return []string{"asa graph", "asa trust", "asa knowledge query"}
	}
}

func explainSourceForContext(ctx ExplainContext) string {
	if ctx.Focus.Kind != "" {
		return "query-bus read-only (" + string(ctx.Focus.Kind) + ")"
	}
	return "query-bus read-only"
}

func (b *queryBus) handleGetAgentTheatre(ctx context.Context, q GetAgentTheatreQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	store, err := b.deps.RuntimeOpen(b.deps.RepoRoot)
	if err != nil {
		return AgentTheatreResult{
			Agents:  []AgentCard{},
			Warning: err.Error(),
		}, nil
	}
	defer func() { _ = store.Close() }()

	limit := firstPositive(q.Limit, 400)
	events, err := store.ListEvents(limit)
	if err != nil {
		return AgentTheatreResult{
			Agents:  []AgentCard{},
			Warning: err.Error(),
		}, nil
	}
	sort.Slice(events, func(i, j int) bool { return events[i].CreatedAt.Before(events[j].CreatedAt) })

	startedAtByKey := map[string]time.Time{}
	agentsByKey := map[string]AgentCard{}
	for _, ev := range events {
		if !strings.HasPrefix(ev.Type, "agent.") {
			continue
		}
		role := emptyFallback(payloadString(ev.Payload, "role"), "agent")
		agentRef := payloadString(ev.Payload, "agent_ref")
		if agentRef == "" {
			agentRef = payloadString(ev.Payload, "agent")
		}
		if agentRef == "" {
			agentRef = "unknown"
		}
		key := role + "|" + agentRef
		if ev.Type == runtime.EventAgentStarted {
			startedAtByKey[key] = ev.CreatedAt
		}

		task := firstNonEmptyString(
			payloadString(ev.Payload, "task"),
			payloadString(ev.Payload, "step"),
			payloadString(ev.Payload, "title"),
		)
		lastOutput := firstNonEmptyString(
			payloadString(ev.Payload, "output"),
			payloadString(ev.Payload, "message"),
			payloadString(ev.Payload, "summary"),
		)
		filesActive := payloadInt(ev.Payload, "files_active", "files_scanned")
		tokensEstimated := payloadInt(ev.Payload, "tokens_estimated", "tokens")
		cost := payloadFloat(ev.Payload, "cost_eur", "cost", "total_cost_eur")
		confidence := payloadFloat(ev.Payload, "confidence", "score")
		durationMs := payloadFloat(ev.Payload, "duration_ms")
		duration := time.Duration(durationMs) * time.Millisecond
		if duration <= 0 {
			if startedAt, ok := startedAtByKey[key]; ok && ev.CreatedAt.After(startedAt) {
				duration = ev.CreatedAt.Sub(startedAt)
			}
		}

		card := AgentCard{
			Role:            role,
			AgentRef:        agentRef,
			Status:          agentStatusFromEvent(ev.Type),
			Task:            task,
			FilesActive:     filesActive,
			Hypothesis:      payloadString(ev.Payload, "hypothesis"),
			TokensEstimated: tokensEstimated,
			CostEUR:         cost,
			Duration:        duration,
			LastOutput:      lastOutput,
			Confidence:      confidence,
			UpdatedAt:       ev.CreatedAt,
		}
		if previous, ok := agentsByKey[key]; ok && previous.UpdatedAt.After(card.UpdatedAt) {
			continue
		}
		agentsByKey[key] = card
	}

	agents := make([]AgentCard, 0, len(agentsByKey))
	for _, card := range agentsByKey {
		agents = append(agents, card)
	}
	sort.Slice(agents, func(i, j int) bool {
		if agents[i].Status != agents[j].Status {
			return agents[i].Status == "running"
		}
		return agents[i].UpdatedAt.After(agents[j].UpdatedAt)
	})
	return AgentTheatreResult{Agents: agents}, nil
}

func (b *queryBus) handleGetReplayPackage(_ context.Context, q GetReplayPackageQuery) (QueryResult, error) {
	replayID, warning := b.resolveReplayID(q.ReplayID)
	if replayID == "" {
		return ReplayPackageResult{
			ReplayID:  replayID,
			Timeline:  []ReplayTimelineEvent{},
			Artifacts: []string{},
			Warnings:  compactWarnings(warning),
			Warning:   warning,
		}, nil
	}
	pkg, err := replay.LoadPackage(b.deps.RepoRoot, replayID)
	if err != nil {
		return ReplayPackageResult{
			ReplayID:  replayID,
			Timeline:  []ReplayTimelineEvent{},
			Artifacts: []string{},
			Warnings:  []string{err.Error()},
			Warning:   err.Error(),
		}, nil
	}

	limit := firstPositive(q.Limit, 120)
	timeline, timelineWarning := replayTimeline(pkg.Path, limit)
	createdAt := pkg.Manifest.CreatedAt
	artifacts := append([]string(nil), pkg.Manifest.Artifacts...)
	mode := strings.TrimSpace(pkg.Manifest.Runtime.RuntimeMode)
	if mode == "" {
		mode = "full"
	}
	warnings := compactWarnings(warning, timelineWarning)
	return ReplayPackageResult{
		ReplayID:   pkg.ID,
		CreatedAt:  createdAt,
		RepoBranch: pkg.Manifest.Repo.Branch,
		RepoCommit: pkg.Manifest.Repo.Commit,
		Mode:       mode,
		Artifacts:  artifacts,
		Timeline:   timeline,
		Warnings:   warnings,
		Warning:    firstNonEmptyString(warning, timelineWarning),
	}, nil
}

func (b *queryBus) handleGetPrototypePipeline(_ context.Context, q GetPrototypePipelineQuery) (QueryResult, error) {
	productID, warning := b.resolvePrototypeProduct(q.Product)
	if productID == "" {
		return PrototypePipelineResult{
			Product:          "",
			PipelineStage:    "wireframe",
			FlowExtraction:   []PrototypeFlowStep{},
			SuggestedActions: []string{"asa prototype create \"<intent>\""},
			Warnings:         compactWarnings(warning),
			Warning:          warning,
		}, nil
	}
	productRoot := filepath.Join(b.deps.RepoRoot, ".asagiri", "products", productID)
	prototypePath := filepath.Join(productRoot, "prototype", "src", "App.tsx")

	stage := "wireframe"
	done := make([]string, 0, 6)
	if fileExists(filepath.Join(productRoot, "prototype", "model.json")) {
		stage = "journey"
		done = append(done, "wireframe", "journey")
	}
	flowSteps := make([]PrototypeFlowStep, 0, firstPositive(q.Limit, 24))
	if steps, err := loadPrototypeFlowSteps(productRoot, firstPositive(q.Limit, 24)); err == nil {
		flowSteps = steps
		if len(flowSteps) > 0 {
			stage = "flow"
			done = appendIfMissing(done, "flow")
		}
	}
	if fileExists(filepath.Join(productRoot, "contracts", "api.openapi.yaml")) {
		stage = "contracts"
		done = appendIfMissing(done, "contracts")
	}
	if fileExists(filepath.Join(productRoot, "generated-specs", "tasks.md")) || fileExists(filepath.Join(productRoot, "tasks", productID+"-001.yaml")) {
		stage = "tasks"
		done = appendIfMissing(done, "tasks")
	}
	if latestGraphForProductFlow(b.deps.RepoRoot, flowSteps) != "" {
		stage = "execution-graph"
		done = appendIfMissing(done, "execution-graph")
	}

	suggested := prototypeSuggestedActions(stage, productID)
	flowID := ""
	if len(flowSteps) > 0 {
		flowID = flowSteps[0].FlowID
	}
	return PrototypePipelineResult{
		Product:          productID,
		WireframeTitle:   strings.ReplaceAll(productID, "-", " "),
		WireframePath:    prototypePath,
		PipelineStage:    stage,
		StagesDone:       done,
		Flow:             flowID,
		FlowExtraction:   flowSteps,
		SuggestedActions: suggested,
		Warnings:         compactWarnings(warning),
		Warning:          warning,
	}, nil
}

func (b *queryBus) handleGetMissionControlSnapshot(ctx context.Context, q GetMissionControlSnapshotQuery) (QueryResult, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	runtimeAny, err := b.handleGetRuntimeStatus(ctx, GetRuntimeStatusQuery{})
	if err != nil {
		return nil, err
	}
	runtimeRes, _ := runtimeAny.(RuntimeStatusResult)

	runsAny, err := b.handleListRuns(ctx, ListRunsQuery{Limit: firstPositive(q.RunsLimit, 8)})
	if err != nil {
		return nil, err
	}
	runsRes, _ := runsAny.(ListRunsResult)

	eventsAny, err := b.handleGetRecentEvents(ctx, GetRecentEventsQuery{Limit: firstPositive(q.EventsLimit, 12)})
	if err != nil {
		return nil, err
	}
	eventsRes, _ := eventsAny.(RecentEventsResult)

	trustAny, err := b.handleGetTrustSummary(ctx, GetTrustSummaryQuery{})
	if err != nil {
		return nil, err
	}
	trustRes, _ := trustAny.(TrustSummaryResult)

	agentsAny, err := b.handleListActiveAgents(ctx, ListActiveAgentsQuery{Limit: firstPositive(q.AgentsLimit, 200)})
	if err != nil {
		return nil, err
	}
	agentsRes, _ := agentsAny.(ActiveAgentsResult)

	flowID := strings.TrimSpace(q.FlowID)
	if flowID == "" {
		for _, ev := range eventsRes.Events {
			if ev.FlowID != "" {
				flowID = ev.FlowID
				break
			}
		}
	}
	flowAny, err := b.handleGetFlowGraph(ctx, GetFlowGraphQuery{FlowID: flowID})
	if err != nil {
		return nil, err
	}
	flowRes, _ := flowAny.(FlowGraphResult)
	flowExplorerAny, err := b.handleGetFlowExplorer(ctx, GetFlowExplorerQuery{FlowID: flowID})
	if err != nil {
		return nil, err
	}
	flowExplorerRes, _ := flowExplorerAny.(FlowExplorerResult)
	graphExplorerAny, err := b.handleGetGraphExplorer(ctx, GetGraphExplorerQuery{FlowID: flowID})
	if err != nil {
		return nil, err
	}
	graphExplorerRes, _ := graphExplorerAny.(GraphExplorerResult)
	knowledgeAny, err := b.handleSearchKnowledge(ctx, SearchKnowledgeQuery{
		Query: firstNonEmptyString(q.Knowledge, flowID),
		Limit: 6,
	})
	if err != nil {
		return nil, err
	}
	knowledgeRes, _ := knowledgeAny.(KnowledgeSearchResult)
	trustExplorerAny, err := b.handleGetTrustExplorer(ctx, GetTrustExplorerQuery{})
	if err != nil {
		return nil, err
	}
	trustExplorerRes, _ := trustExplorerAny.(TrustExplorerResult)
	explainAny, err := b.handleGetExplain(ctx, GetExplainQuery{
		Subject: q.ExplainFor,
		Context: q.ExplainContext,
	})
	if err != nil {
		return nil, err
	}
	explainRes, _ := explainAny.(ExplainResult)
	agentTheatreAny, err := b.handleGetAgentTheatre(ctx, GetAgentTheatreQuery{Limit: firstPositive(q.AgentsLimit, 200)})
	if err != nil {
		return nil, err
	}
	agentTheatreRes, _ := agentTheatreAny.(AgentTheatreResult)
	replayAny, err := b.handleGetReplayPackage(ctx, GetReplayPackageQuery{ReplayID: q.ReplayID, Limit: firstPositive(q.EventsLimit, 60)})
	if err != nil {
		return nil, err
	}
	replayRes, _ := replayAny.(ReplayPackageResult)
	prototypeAny, err := b.handleGetPrototypePipeline(ctx, GetPrototypePipelineQuery{
		Product: q.PrototypeProduct,
		Limit:   firstPositive(q.PrototypeFlowLimit, 24),
	})
	if err != nil {
		return nil, err
	}
	prototypeRes, _ := prototypeAny.(PrototypePipelineResult)

	recommendedAny, err := b.handleGetRecommendedActions(ctx, GetRecommendedActionsQuery{FlowID: flowID})
	if err != nil {
		return nil, err
	}
	recommendedRes, _ := recommendedAny.(RecommendedActionsResult)

	readinessAny, _ := b.handleGetReadiness(ctx, GetReadinessQuery{})
	readinessRes, _ := readinessAny.(ReadinessResult)
	if !readinessRes.Ready {
		prepended := []RecommendedAction{{
			ID:            "rec.complete-onboarding",
			Title:         "Complete onboarding",
			Description:   fmt.Sprintf("Project readiness %d%%", readinessRes.Score),
			Priority:      0,
			CLIEquivalent: "asa onboard --yes --non-interactive",
			ActionID:      "cmd.complete-onboarding",
		}}
		recommendedRes.Actions = append(prepended, recommendedRes.Actions...)
	}

	warnings := compactWarnings(
		runtimeRes.Warning,
		runsRes.Warning,
		trustRes.Warning,
		agentsRes.Warning,
		flowRes.Warning,
		flowExplorerRes.Warning,
		graphExplorerRes.Warning,
		knowledgeRes.Warning,
		trustExplorerRes.Warning,
		explainRes.Warning,
		agentTheatreRes.Warning,
		replayRes.Warning,
		prototypeRes.Warning,
	)
	costToday, costMonth := deriveCosts(eventsRes.Events)

	// Cost efficiency snapshot — 7-day window; gracefully degrades on errors.
	var efficiencySnap CostEfficiencySnapshot
	if effStore, errEff := b.deps.StateOpen(b.deps.StateDBPath); errEff == nil {
		_ = effStore.Migrate()
		sinceEff := time.Now().AddDate(0, 0, -7)
		if tokens, err := effStore.QueryStepTokens(ctx, sinceEff); err == nil {
			if steps, err := effStore.SummarizeStepsSince(ctx, sinceEff); err == nil {
				if runs, err := effStore.QuerySince(ctx, sinceEff); err == nil {
					tot := telemetry.SummarizeRuns(runs)
					actualCents := int64(0)
					for _, r := range runs {
						actualCents += r.ActualCostCents
					}
					cTot := telemetry.CostTotals{RunCount: tot.RunCount, ActualCostCents: actualCents}
					w := cost.BuildWindowReport("7d", cTot, tokens, steps, b.deps.Config)
					efficiencySnap = CostEfficiencySnapshot{
						ActualCostCents:       w.ActualCostCents,
						PremiumEquivCents:     w.Savings.PremiumEquivCents,
						SavingsCents:          w.Savings.SavingsCents,
						SavingsRate:           w.Savings.SavingsRate,
						LocalPct:              w.Savings.LocalPct(),
						CloudPct:              100 - w.Savings.LocalPct(),
						StrategyScore:         w.Strategy.Grade,
						EscalationRate:        w.Escalations.EscalationRate,
						LocalSteps:            w.Escalations.LocalSteps,
						PremiumEscalations:    w.Escalations.PremiumEscalations,
						TotalSteps:            w.Escalations.TotalSteps,
						PremiumReferenceModel: w.Savings.PremiumReferenceModel,
						Currency:              w.Savings.Currency,
					}
				}
			}
		}
		_ = effStore.Close()
	}
	workspace := filepath.Base(b.deps.RepoRoot)
	branch := "-"
	if len(runsRes.Runs) > 0 && runsRes.Runs[0].Feature != "" {
		branch = runsRes.Runs[0].Feature
	}
	sessionStatus := "inactive"
	if runtimeRes.Status.Sessions > 0 {
		sessionStatus = "active"
	}
	return MissionControlSnapshotResult{
		Workspace:          workspace,
		Branch:             branch,
		SessionStatus:      sessionStatus,
		Runtime:            runtimeRes,
		Trust:              trustRes,
		Runs:               runsRes.Runs,
		Events:             eventsRes.Events,
		ActiveAgents:       agentsRes.Agents,
		Flow:               flowRes,
		FlowExplorer:       flowExplorerRes,
		GraphExplorer:      graphExplorerRes,
		Knowledge:          knowledgeRes,
		TrustExplorer:      trustExplorerRes,
		Explain:            explainRes,
		AgentTheatre:       agentTheatreRes,
		Replay:             replayRes,
		Prototype:          prototypeRes,
		CostTodayEUR:       costToday,
		CostMonthEUR:       costMonth,
		CostEfficiency:     efficiencySnap,
		RecommendedActions: recommendedRes.Actions,
		Readiness:          readinessRes,
		UpdatedAt:          time.Now().UTC(),
		Warnings:           warnings,
	}, nil
}

func latestTrustReport(repoRoot string) (*trust.TrustReport, string) {
	trustDir := filepath.Join(repoRoot, ".asagiri", "trust")
	entries, err := os.ReadDir(trustDir)
	if err != nil {
		return nil, "trust report unavailable"
	}
	type candidate struct {
		path string
		ts   time.Time
	}
	candidates := make([]candidate, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		reportPath := filepath.Join(trustDir, e.Name(), "report.json")
		info, err := os.Stat(reportPath)
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{path: reportPath, ts: info.ModTime()})
	}
	if len(candidates) == 0 {
		return nil, "trust report unavailable"
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ts.After(candidates[j].ts) })
	body, err := os.ReadFile(candidates[0].path)
	if err != nil {
		return nil, "trust report unreadable"
	}
	var report trust.TrustReport
	if err := json.Unmarshal(body, &report); err != nil {
		return nil, "trust report invalid"
	}
	return &report, ""
}

func latestFlowGraph(repoRoot, flowID string) (*executiongraph.ExecutionGraph, string) {
	graphsRoot := filepath.Join(repoRoot, ".asagiri", "graphs")
	entries, err := os.ReadDir(graphsRoot)
	if err != nil {
		return nil, "flow graph unavailable"
	}
	type candidate struct {
		id   string
		path string
		ts   time.Time
	}
	candidates := make([]candidate, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		graphPath := filepath.Join(graphsRoot, e.Name(), "execution-graph.json")
		info, err := os.Stat(graphPath)
		if err != nil {
			continue
		}
		candidates = append(candidates, candidate{id: e.Name(), path: graphPath, ts: info.ModTime()})
	}
	if len(candidates) == 0 {
		return nil, "flow graph unavailable"
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ts.After(candidates[j].ts) })
	repo := executiongraph.NewRepository(repoRoot)
	for _, c := range candidates {
		graph, err := repo.Load(c.id)
		if err != nil {
			continue
		}
		if flowID != "" && graph.Flow != flowID {
			continue
		}
		return &graph, ""
	}
	return nil, "flow graph unavailable"
}

func payloadString(payload map[string]any, key string) string {
	if payload == nil {
		return ""
	}
	raw, ok := payload[key]
	if !ok || raw == nil {
		return ""
	}
	v, ok := raw.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(v)
}

func agentStatusFromEvent(eventType string) string {
	switch eventType {
	case runtime.EventAgentStarted:
		return "running"
	case runtime.EventAgentCompleted:
		return "done"
	case runtime.EventAgentFailed:
		return "failed"
	case runtime.EventAgentBlocked, runtime.EventAgentReviewRequested:
		return "waiting"
	case runtime.EventAgentReviewRejected:
		return "blocked"
	default:
		return "unknown"
	}
}

func compactWarnings(values ...string) []string {
	out := make([]string, 0, len(values))
	seen := map[string]struct{}{}
	for _, value := range values {
		v := strings.TrimSpace(value)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}

func deriveCosts(events []EventSummary) (today float64, month float64) {
	now := time.Now().UTC()
	dayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, time.UTC)
	for _, ev := range events {
		cost := payloadCost(ev.Payload)
		if cost <= 0 {
			continue
		}
		if ev.CreatedAt.After(monthStart) {
			month += cost
		}
		if ev.CreatedAt.After(dayStart) {
			today += cost
		}
	}
	return today, month
}

func payloadCost(payload map[string]any) float64 {
	keys := []string{"cost_eur", "cost", "total_cost_eur"}
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		}
	}
	return 0
}

func firstPositive(value int, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}

func trustScoreFromRisk(r executiongraph.RiskLevel) float64 {
	switch r {
	case executiongraph.RiskLevelLow:
		return 0.88
	case executiongraph.RiskLevelMedium:
		return 0.72
	case executiongraph.RiskLevelHigh:
		return 0.56
	case executiongraph.RiskLevelCritical:
		return 0.41
	default:
		return 0.65
	}
}

func findingsForCheckType(checks []trust.VerificationCheck, kind trust.CheckType) []string {
	out := make([]string, 0, 4)
	for _, check := range checks {
		if check.Type != kind {
			continue
		}
		for _, finding := range check.Findings {
			out = append(out, finding.Message)
			if len(out) >= 4 {
				return out
			}
		}
	}
	return out
}

func evidenceForCheckType(checks []trust.VerificationCheck, kind trust.CheckType) []string {
	out := make([]string, 0, 4)
	for _, check := range checks {
		if check.Type != kind {
			continue
		}
		for _, evidence := range check.Evidence {
			value := strings.TrimSpace(evidence.Summary)
			if value == "" {
				value = strings.TrimSpace(evidence.Source)
			}
			if value == "" {
				continue
			}
			out = append(out, value)
			if len(out) >= 4 {
				return out
			}
		}
	}
	return out
}

func knowledgeScore(node knowledge.GraphNode, query string) float64 {
	score := 0.2
	if strings.Contains(strings.ToLower(node.ID), query) {
		score += 0.45
	}
	if strings.Contains(strings.ToLower(node.Name), query) {
		score += 0.35
	}
	if strings.Contains(strings.ToLower(node.Path), query) {
		score += 0.2
	}
	if score > 1 {
		return 1
	}
	return score
}

func emptyFallback(value string, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func payloadFloat(payload map[string]any, keys ...string) float64 {
	for _, key := range keys {
		raw, ok := payload[key]
		if !ok || raw == nil {
			continue
		}
		switch v := raw.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int32:
			return float64(v)
		case int64:
			return float64(v)
		case string:
			n, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
			if err == nil {
				return n
			}
		}
	}
	return 0
}

func payloadInt(payload map[string]any, keys ...string) int {
	return int(payloadFloat(payload, keys...))
}

func (b *queryBus) resolveReplayID(requested string) (string, string) {
	id := strings.TrimSpace(requested)
	if id != "" {
		return id, ""
	}
	replaysRoot := filepath.Join(b.deps.RepoRoot, replay.RelDir)
	entries, err := os.ReadDir(replaysRoot)
	if err != nil {
		return "", "replay package unavailable"
	}
	type replayCandidate struct {
		id string
		ts time.Time
	}
	candidates := make([]replayCandidate, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() || entry.Name() == replay.SnapshotsRelDir {
			continue
		}
		manifestPath := filepath.Join(replaysRoot, entry.Name(), replay.ManifestName)
		info, err := os.Stat(manifestPath)
		if err != nil {
			continue
		}
		candidates = append(candidates, replayCandidate{id: entry.Name(), ts: info.ModTime()})
	}
	if len(candidates) == 0 {
		return "", "replay package unavailable"
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ts.After(candidates[j].ts) })
	return candidates[0].id, ""
}

func replayTimeline(replayDir string, limit int) ([]ReplayTimelineEvent, string) {
	eventsPath := filepath.Join(replayDir, "runtime", "events.jsonl")
	body, err := replay.ReadMaybeCompressed(eventsPath)
	if err != nil {
		return []ReplayTimelineEvent{}, "replay timeline unavailable"
	}
	lines := strings.Split(strings.TrimSpace(string(body)), "\n")
	events := make([]ReplayTimelineEvent, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		var ev runtime.RuntimeEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			continue
		}
		events = append(events, ReplayTimelineEvent{
			Time:      ev.CreatedAt,
			Type:      ev.Type,
			Source:    ev.Source,
			SessionID: ev.SessionID,
			FlowID:    ev.FlowID,
			Artifact:  payloadString(ev.Payload, "artifact"),
		})
	}
	sort.Slice(events, func(i, j int) bool { return events[i].Time.Before(events[j].Time) })
	if limit > 0 && len(events) > limit {
		events = events[:limit]
	}
	return events, ""
}

func (b *queryBus) resolvePrototypeProduct(requested string) (string, string) {
	productID := strings.TrimSpace(requested)
	if productID != "" {
		return product.Slug(productID), ""
	}
	productsRoot := filepath.Join(b.deps.RepoRoot, ".asagiri", "products")
	entries, err := os.ReadDir(productsRoot)
	if err != nil {
		return "", "prototype product unavailable"
	}
	type productCandidate struct {
		id string
		ts time.Time
	}
	candidates := make([]productCandidate, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		protoModel := filepath.Join(productsRoot, entry.Name(), "prototype", "model.json")
		info, err := os.Stat(protoModel)
		if err != nil {
			continue
		}
		candidates = append(candidates, productCandidate{id: entry.Name(), ts: info.ModTime()})
	}
	if len(candidates) == 0 {
		return "", "prototype product unavailable"
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].ts.After(candidates[j].ts) })
	return candidates[0].id, ""
}

func loadPrototypeFlowSteps(productRoot string, limit int) ([]PrototypeFlowStep, error) {
	flowsDir := filepath.Join(productRoot, "flows")
	entries, err := os.ReadDir(flowsDir)
	if err != nil {
		return nil, err
	}
	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })
	steps := make([]PrototypeFlowStep, 0, limit)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".flow.yaml") {
			continue
		}
		body, err := os.ReadFile(filepath.Join(flowsDir, entry.Name()))
		if err != nil {
			continue
		}
		var flow product.Flow
		if err := yaml.Unmarshal(body, &flow); err != nil {
			continue
		}
		for _, step := range flow.Steps {
			trust := "pending"
			if step.Sensitive {
				trust = "review-required"
			}
			metric := ""
			if len(flow.Metrics) > 0 {
				metric = flow.Metrics[0]
			}
			steps = append(steps, PrototypeFlowStep{
				FlowID:    flow.ID,
				StepID:    step.ID,
				Action:    step.Action,
				Screen:    step.Screen,
				Next:      step.Next,
				Contract:  FormatContractRef(step.ContractRef),
				Trust:     trust,
				Metric:    metric,
				Sensitive: step.Sensitive,
			})
			if limit > 0 && len(steps) >= limit {
				return steps, nil
			}
		}
	}
	return steps, nil
}

func latestGraphForProductFlow(repoRoot string, steps []PrototypeFlowStep) string {
	if len(steps) == 0 {
		return ""
	}
	flowID := steps[0].FlowID
	graph, _ := latestFlowGraph(repoRoot, flowID)
	if graph == nil {
		return ""
	}
	return graph.ID
}

func prototypeSuggestedActions(stage, productID string) []string {
	switch stage {
	case "wireframe":
		return []string{`asa prototype create "<intent>"`}
	case "journey":
		return []string{
			"asa flows extract " + productID,
			"asa flows inspect " + productID,
		}
	case "flow":
		return []string{
			"asa contracts extract " + productID,
			"asa architecture derive " + productID,
		}
	case "contracts":
		return []string{
			"asa spec generate-from-product " + productID,
			"asa flows review " + productID,
		}
	case "tasks":
		return []string{
			"asa graph run " + productID + " --flow <flow-id> --dry-run",
			"asa graph visualize <graph-id> --format mermaid",
		}
	default:
		return []string{
			"asa graph status <graph-id>",
			"asa replay create --from-graph <graph-id>",
		}
	}
}

func appendIfMissing(values []string, value string) []string {
	for _, existing := range values {
		if existing == value {
			return values
		}
	}
	return append(values, value)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
