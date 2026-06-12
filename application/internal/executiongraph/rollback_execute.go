package executiongraph

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// RollbackImpact summarizes safety UX data before a graph rollback (spec-ui §30).
type RollbackImpact struct {
	GraphID           string
	Title             string
	ImpactedWorktrees int
	StaleReports      int
	ActiveSessions    int
	ImpactLines       []string
	RollbackPossible  bool
	RollbackPolicy    string
	CLIEquivalent     string
	HighRiskUncovered int
}

// GraphRollbackResult is returned after rollback execution.
type GraphRollbackResult struct {
	GraphID         string
	Status          GraphStatus
	NodesRolledBack int
	DryRun          bool
}

// AssessRollbackImpact loads the graph and derives rollback safety metadata.
func AssessRollbackImpact(repoRoot, graphID string) (RollbackImpact, error) {
	graphID = strings.TrimSpace(graphID)
	if graphID == "" {
		return RollbackImpact{}, fmt.Errorf("graph rollback: id required")
	}
	repo := NewRepository(repoRoot)
	graph, err := repo.Load(graphID)
	if err != nil {
		return RollbackImpact{}, err
	}

	assessment := PlanRollback(graph, nil)
	worktrees := 0
	for _, strategy := range assessment.NodeStrategies {
		if strategy == RollbackStrategyWorktreeReset {
			worktrees++
		}
	}
	reports := countGraphArtifacts(filepath.Join(repoRoot, ".asagiri", "graphs", graph.ID))
	sessions := countActiveSessions(repoRoot)

	impact := RollbackImpact{
		GraphID:           graph.ID,
		Title:             fmt.Sprintf("You are about to rollback %s.", graph.ID),
		ImpactedWorktrees: worktrees,
		StaleReports:      reports,
		ActiveSessions:    sessions,
		RollbackPossible:  graph.Status != GraphStatusRolledBack,
		CLIEquivalent:     "asa graph rollback " + graph.ID,
		HighRiskUncovered: len(assessment.MissingStrategy),
	}
	if graph.Rollback != nil && graph.Rollback.Strategy != "" {
		impact.RollbackPolicy = fmt.Sprintf("Default strategy: %s (preserve reports: %t).", graph.Rollback.Strategy, graph.Rollback.PreserveReports)
	} else if assessment.GraphDefault.Strategy != "" {
		impact.RollbackPolicy = fmt.Sprintf("Default strategy: %s (preserve reports: %t).", assessment.GraphDefault.Strategy, assessment.GraphDefault.PreserveReports)
	} else {
		impact.RollbackPolicy = "Rollback can be replayed manually after confirmation."
	}
	if impact.ImpactedWorktrees > 0 {
		impact.ImpactLines = append(impact.ImpactLines, fmt.Sprintf("%d worktree(s) will be impacted", impact.ImpactedWorktrees))
	}
	if impact.StaleReports > 0 {
		impact.ImpactLines = append(impact.ImpactLines, fmt.Sprintf("%d generated report(s) may become stale", impact.StaleReports))
	}
	if impact.ActiveSessions > 0 {
		impact.ImpactLines = append(impact.ImpactLines, fmt.Sprintf("%d active session(s) will require a manual refresh", impact.ActiveSessions))
	}
	if impact.HighRiskUncovered > 0 {
		impact.ImpactLines = append(impact.ImpactLines, fmt.Sprintf("%d high-risk node(s) lack an explicit rollback strategy", impact.HighRiskUncovered))
	}
	if len(impact.ImpactLines) == 0 {
		impact.ImpactLines = []string{"No additional artefacts detected beyond graph state."}
	}
	if !impact.RollbackPossible {
		impact.RollbackPolicy = "Graph is already rolled back."
	}
	return impact, nil
}

// ExecuteGraphRollback marks the graph and eligible nodes as rolled back and persists the graph.
func ExecuteGraphRollback(ctx context.Context, repoRoot, graphID string, dryRun bool) (GraphRollbackResult, error) {
	if err := ctx.Err(); err != nil {
		return GraphRollbackResult{}, err
	}
	graphID = strings.TrimSpace(graphID)
	if graphID == "" {
		return GraphRollbackResult{}, fmt.Errorf("graph rollback: id required")
	}
	repo := NewRepository(repoRoot)
	graph, err := repo.Load(graphID)
	if err != nil {
		return GraphRollbackResult{}, err
	}
	if graph.Status == GraphStatusRolledBack {
		return GraphRollbackResult{
			GraphID: graph.ID,
			Status:  graph.Status,
			DryRun:  dryRun,
		}, nil
	}

	rolled := 0
	for i := range graph.Nodes {
		switch graph.Nodes[i].Status {
		case NodeStatusRunning, NodeStatusSucceeded, NodeStatusFailed, NodeStatusBlocked, NodeStatusReady:
			graph.Nodes[i].Status = NodeStatusRolledBack
			rolled++
		}
	}
	graph.Status = GraphStatusRolledBack

	result := GraphRollbackResult{
		GraphID:         graph.ID,
		Status:          graph.Status,
		NodesRolledBack: rolled,
		DryRun:          dryRun,
	}
	if dryRun {
		return result, nil
	}
	if _, _, err := repo.Save(graph); err != nil {
		return GraphRollbackResult{}, err
	}
	emitGraphRollbackEvent(repoRoot, graph.ID, rolled)
	return result, nil
}

func countGraphArtifacts(dir string) int {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0
	}
	count := 0
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		switch name {
		case "execution-graph.yaml", "execution-graph.json":
			continue
		default:
			count++
		}
	}
	return count
}

func countActiveSessions(repoRoot string) int {
	store, err := runtime.Open(repoRoot)
	if err != nil {
		return 0
	}
	defer func() { _ = store.Close() }()
	sessions, err := store.ListSessions()
	if err != nil {
		return 0
	}
	active := 0
	for _, s := range sessions {
		if s.Status == runtime.SessionActive {
			active++
		}
	}
	if active == 0 && len(sessions) > 0 {
		return len(sessions)
	}
	return active
}

func emitGraphRollbackEvent(repoRoot, graphID string, nodes int) {
	store, err := runtime.Open(repoRoot)
	if err != nil {
		return
	}
	defer func() { _ = store.Close() }()
	emitter := &runtime.GraphEmitter{Store: store}
	_ = emitter.Emit(runtime.EventGraphBlocked, graphID, "", map[string]any{
		"action": "rollback",
		"nodes":  nodes,
	})
}
