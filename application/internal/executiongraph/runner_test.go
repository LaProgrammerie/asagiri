package executiongraph

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

func TestDefaultRunnerRunCompletesWithEventsAndCheckpoints(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	runGitInit(t, repo)

	planner := Planner{
		RepoRoot: repo,
		Inferer:  DefaultDependencyInferer{},
		Now:      func() time.Time { return time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC) },
	}
	graph, err := planner.Build(t.Context(), GraphPlanRequest{
		Product:        "minimal-product",
		Flow:           "workspace-onboarding",
		IncludeReviews: true,
	})
	require.NoError(t, err)

	sched := DefaultScheduler{}
	schedule, err := sched.Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.NoError(t, err)

	repoObj := NewRepository(repo)
	_, err = repoObj.SaveAll(graph, &schedule)
	require.NoError(t, err)

	runner := NewRunner(repo)
	result, err := runner.Run(t.Context(), graph, schedule, RunOptions{})
	require.NoError(t, err)
	require.Equal(t, GraphStatusCompleted, result.Status)

	loaded, err := repoObj.Load(graph.ID)
	require.NoError(t, err)
	for _, n := range loaded.Nodes {
		if n.Type == NodeTypeInvestigation || n.Type == NodeTypeImplementation {
			require.Equal(t, NodeStatusSucceeded, n.Status)
		}
	}

	eventsPath := filepath.Join(repo, ".asagiri", "graphs", graph.ID, "events.jsonl")
	raw, err := os.ReadFile(eventsPath)
	require.NoError(t, err)
	body := string(raw)
	require.Contains(t, body, runtime.EventGraphStarted)
	require.Contains(t, body, runtime.EventGraphNodeCompleted)
	require.Contains(t, body, runtime.EventGraphCompleted)

	_, found, err := repoObj.LoadLatestCheckpoint(graph.ID)
	require.NoError(t, err)
	require.True(t, found)
}

func TestDefaultRunnerResumeFromCheckpoint(t *testing.T) {
	repo := writeMinimalPlanningFixture(t)
	runGitInit(t, repo)

	graph := ExecutionGraph{
		ID:        "graph-2026-05-29-resume01",
		Product:   "minimal-product",
		Flow:      "workspace-onboarding",
		Status:    GraphStatusPaused,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Strategy:  Strategy{MaxParallel: 2, StopOnRisk: RiskLevelHigh},
		Nodes: []GraphNode{
			{ID: "investigate-onboarding", Type: NodeTypeInvestigation, Title: "Investigate", Agent: "local", Risk: RiskLevelLow, Status: NodeStatusSucceeded},
			{ID: "implement-click-get-started", Type: NodeTypeImplementation, Title: "Implement A", Agent: "cursor", Risk: RiskLevelMedium, Status: NodeStatusSucceeded},
			{ID: "implement-invite-member", Type: NodeTypeImplementation, Title: "Implement B", Agent: "cursor", Risk: RiskLevelHigh, Status: NodeStatusPending},
			{ID: "verify-onboarding-flow", Type: NodeTypeValidation, Title: "Verify", Agent: "local", Risk: RiskLevelMedium, Status: NodeStatusPending},
		},
		Edges: []GraphEdge{
			{From: "investigate-onboarding", To: "implement-click-get-started", Type: EdgeTypeProducesContextFor},
			{From: "implement-click-get-started", To: "implement-invite-member", Type: EdgeTypeMustRunAfter},
			{From: "implement-invite-member", To: "verify-onboarding-flow", Type: EdgeTypeMustRunAfter},
		},
		Checkpoints: []Checkpoint{
			{After: "implement-click-get-started"},
			{After: "implement-invite-member"},
		},
	}
	require.NoError(t, graph.Validate())

	repoObj := NewRepository(repo)
	sched := DefaultScheduler{}
	schedule, err := sched.Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.NoError(t, err)
	_, err = repoObj.SaveAll(graph, &schedule)
	require.NoError(t, err)

	gitRef, _ := CaptureGitState(repo)
	_, err = repoObj.SaveCheckpoint(graph.ID, CheckpointState{
		AfterNode: "implement-click-get-started",
		GitRef:    gitRef,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	})
	require.NoError(t, err)

	runner := NewRunner(repo)
	result, err := runner.Resume(t.Context(), graph.ID, RunOptions{})
	require.NoError(t, err)
	require.Equal(t, GraphStatusCompleted, result.Status)

	loaded, err := repoObj.Load(graph.ID)
	require.NoError(t, err)
	require.Equal(t, NodeStatusSucceeded, loaded.Nodes[nodeIndex(loaded.Nodes, "implement-invite-member")].Status)
}

func TestDefaultRunnerStrictTrustBlocksTrustGate(t *testing.T) {
	repo := t.TempDir()
	graph := ExecutionGraph{
		ID:        "graph-2026-05-29-trust01",
		Product:   "p",
		Flow:      "f",
		Status:    GraphStatusReady,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Strategy:  Strategy{MaxParallel: 1, StrictTrust: true},
		Nodes: []GraphNode{
			{ID: "trust-gate", Type: NodeTypeTrustVerification, Title: "Trust gate", Agent: "local", Risk: RiskLevelHigh},
		},
		Checkpoints: []Checkpoint{{After: "trust-gate"}},
	}
	require.NoError(t, graph.Validate())

	sched := ExecutionSchedule{
		GraphID:        graph.ID,
		ParallelGroups: [][]string{{"trust-gate"}},
	}
	repoObj := NewRepository(repo)
	_, err := repoObj.SaveAll(graph, &sched)
	require.NoError(t, err)

	runner := NewRunner(repo)
	gates := trust.NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {RequiredChecks: []string{"contracts"}},
		},
	})
	result, err := runner.Run(context.Background(), graph, sched, RunOptions{
		StrictTrust: true,
		Gates:       gates,
	})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrTrustGateBlocked)
	require.Equal(t, GraphStatusBlocked, result.Status)
}

func TestApplyTrustEnrichmentInsertsGateForPublicContract(t *testing.T) {
	nodes := []GraphNode{
		{ID: "implement-click-get-started", Type: NodeTypeImplementation, Title: "Implement", Agent: "cursor"},
	}
	bindings := []TaskBinding{
		{
			NodeID:      "implement-click-get-started",
			Action:      "click_get_started",
			ContractRef: "POST /api/workspaces",
		},
	}
	out, edges := ApplyTrustEnrichment(nodes, bindings, nil, TrustEnrichmentInput{})
	require.Len(t, out, 2)
	require.Equal(t, NodeTypeTrustVerification, out[1].Type)
	require.Equal(t, "trust-gate-click-get-started", out[1].ID)
	require.Len(t, edges, 1)
	require.Equal(t, "implement-click-get-started", edges[0].From)
}

func TestApplyInvestigationEnrichmentInsertsBeforeSensitiveTask(t *testing.T) {
	nodes := []GraphNode{
		{ID: "implement-invite-member", Type: NodeTypeImplementation, Title: "Invite", Agent: "cursor"},
	}
	bindings := []TaskBinding{
		{
			NodeID:      "implement-invite-member",
			Action:      "invite_member",
			ContractRef: "TODO:auth.signup",
			Sensitive:   true,
		},
	}
	flow := product.Flow{Business: product.FlowBusiness{Criticality: "high"}}
	out, edges := ApplyInvestigationEnrichment(nodes, bindings, nil, flow)
	require.GreaterOrEqual(t, len(out), 2)
	found := false
	for _, n := range out {
		if n.ID == "investigate-invite-member" {
			found = true
			require.Equal(t, NodeTypeInvestigation, n.Type)
		}
	}
	require.True(t, found)
	require.NotEmpty(t, edges)
}

func runGitInit(t *testing.T, repo string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(repo, 0o755))
}
