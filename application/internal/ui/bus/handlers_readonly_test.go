package bus

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/stretchr/testify/require"
)

type mockRuntimeStore struct {
	status    runtime.DaemonStatus
	statusErr error
	events    []runtime.RuntimeEvent
	eventsErr error
	sessions  []runtime.Session
	metrics   runtime.MetricsSnapshot
	closeErr  error
}

func (m *mockRuntimeStore) Close() error { return m.closeErr }

func (m *mockRuntimeStore) Status() (runtime.DaemonStatus, error) {
	return m.status, m.statusErr
}

func (m *mockRuntimeStore) ListEvents(limit int) ([]runtime.RuntimeEvent, error) {
	if m.eventsErr != nil {
		return nil, m.eventsErr
	}
	if limit <= 0 || limit >= len(m.events) {
		return m.events, nil
	}
	return m.events[:limit], nil
}

func (m *mockRuntimeStore) ListSessions() ([]runtime.Session, error) {
	return m.sessions, nil
}

func (m *mockRuntimeStore) CollectMetrics() (runtime.MetricsSnapshot, error) {
	return m.metrics, nil
}

type mockStateStore struct {
	runs       []sqlite.Run
	tasks      []sqlite.Task
	metric     *telemetry.RunMetric
	getRun     *sqlite.Run
	getRunErr  error
	listErr    error
	migrateErr error
	closeErr   error
}

func (m *mockStateStore) Close() error { return m.closeErr }

func (m *mockStateStore) Migrate() error { return m.migrateErr }

func (m *mockStateStore) ListRuns(limit int) ([]sqlite.Run, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	if limit <= 0 || limit >= len(m.runs) {
		return m.runs, nil
	}
	return m.runs[:limit], nil
}

func (m *mockStateStore) GetRun(id string) (*sqlite.Run, error) {
	if m.getRunErr != nil {
		return nil, m.getRunErr
	}
	if m.getRun != nil {
		return m.getRun, nil
	}
	for i := range m.runs {
		if m.runs[i].ID == id {
			r := m.runs[i]
			return &r, nil
		}
	}
	return nil, nil
}

func (m *mockStateStore) ListTasksByRun(runID string) ([]sqlite.Task, error) {
	var out []sqlite.Task
	for _, t := range m.tasks {
		if t.RunID == runID {
			out = append(out, t)
		}
	}
	return out, nil
}

func (m *mockStateStore) GetRunMetric(runID string) (*telemetry.RunMetric, error) {
	return m.metric, nil
}
func (m *mockStateStore) QuerySince(_ context.Context, _ time.Time) ([]telemetry.RunMetric, error) {
	return nil, nil
}

func (m *mockStateStore) SummarizeStepsSince(_ context.Context, _ time.Time) (telemetry.StepTotals, error) {
	return telemetry.StepTotals{}, nil
}

func (m *mockStateStore) QueryStepTokens(_ context.Context, _ time.Time) (telemetry.StepTokenTotals, error) {
	return telemetry.StepTokenTotals{}, nil
}

func (m *mockStateStore) QueryRunsBetween(_ context.Context, _, _ time.Time) ([]telemetry.RunMetric, error) {
	return nil, nil
}

func (m *mockStateStore) SummarizeStepsBetween(_ context.Context, _, _ time.Time) (telemetry.StepTotals, error) {
	return telemetry.StepTotals{}, nil
}

func (m *mockStateStore) QueryStepTokensBetween(_ context.Context, _, _ time.Time) (telemetry.StepTokenTotals, error) {
	return telemetry.StepTokenTotals{}, nil
}

func TestQueryBusRuntimeStatusOpenFailureReturnsWarning(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return nil, errors.New("runtime unavailable")
		},
	})

	rsAny, err := qb.Query(context.Background(), GetRuntimeStatusQuery{})
	require.NoError(t, err)

	rs, ok := rsAny.(RuntimeStatusResult)
	require.True(t, ok)
	require.Equal(t, "runtime unavailable", rs.Warning)
	require.Zero(t, rs.Status.Sessions)
}

func TestQueryBusRuntimeStatusSuccess(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{
				status: runtime.DaemonStatus{
					Running:  true,
					Sessions: 3,
					DBPath:   "/tmp/runtime.sqlite",
				},
			}, nil
		},
	})

	rsAny, err := qb.Query(context.Background(), GetRuntimeStatusQuery{})
	require.NoError(t, err)

	rs, ok := rsAny.(RuntimeStatusResult)
	require.True(t, ok)
	require.Equal(t, 3, rs.Status.Sessions)
	require.True(t, rs.Status.Running)
	require.Empty(t, rs.Warning)
}

func TestQueryBusListRunsMapsSummaries(t *testing.T) {
	t.Parallel()

	created := time.Date(2026, 5, 29, 10, 0, 0, 0, time.UTC)
	updated := time.Date(2026, 5, 29, 11, 0, 0, 0, time.UTC)
	qb := NewQueryBus(Deps{
		StateOpen: func(string) (stateStore, error) {
			return &mockStateStore{
				runs: []sqlite.Run{{
					ID:        "run-42",
					Feature:   "spec-ui",
					Status:    sqlite.StatusRunning,
					CreatedAt: created,
					UpdatedAt: updated,
				}},
			}, nil
		},
	})

	runsAny, err := qb.Query(context.Background(), ListRunsQuery{Limit: 5})
	require.NoError(t, err)

	runs, ok := runsAny.(ListRunsResult)
	require.True(t, ok)
	require.Len(t, runs.Runs, 1)
	require.Equal(t, "run-42", runs.Runs[0].ID)
	require.Equal(t, "spec-ui", runs.Runs[0].Feature)
	require.Equal(t, sqlite.StatusRunning, runs.Runs[0].Status)
	require.Equal(t, created, runs.Runs[0].CreatedAt)
}

func TestQueryBusListRunsOpenFailureReturnsWarning(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{
		StateOpen: func(string) (stateStore, error) {
			return nil, errors.New("state unavailable")
		},
	})

	runsAny, err := qb.Query(context.Background(), ListRunsQuery{Limit: 5})
	require.NoError(t, err)

	runs, ok := runsAny.(ListRunsResult)
	require.True(t, ok)
	require.Empty(t, runs.Runs)
	require.Equal(t, "state unavailable", runs.Warning)
}

func TestQueryBusRecentEventsOpenFailureReturnsEmpty(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return nil, errors.New("runtime unavailable")
		},
	})

	eventsAny, err := qb.Query(context.Background(), GetRecentEventsQuery{Limit: 10})
	require.NoError(t, err)

	events, ok := eventsAny.(RecentEventsResult)
	require.True(t, ok)
	require.Empty(t, events.Events)
}

func TestQueryBusRecentEventsMapsSummaries(t *testing.T) {
	t.Parallel()

	when := time.Date(2026, 5, 29, 9, 30, 0, 0, time.UTC)
	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{
				events: []runtime.RuntimeEvent{{
					ID:        "ev-1",
					Type:      "runtime.started",
					Source:    "tests",
					SessionID: "sess-1",
					FlowID:    "flow-1",
					CreatedAt: when,
				}},
			}, nil
		},
	})

	eventsAny, err := qb.Query(context.Background(), GetRecentEventsQuery{Limit: 10})
	require.NoError(t, err)

	events, ok := eventsAny.(RecentEventsResult)
	require.True(t, ok)
	require.Len(t, events.Events, 1)
	require.Equal(t, "ev-1", events.Events[0].ID)
	require.Equal(t, "runtime.started", events.Events[0].Type)
	require.Equal(t, when, events.Events[0].CreatedAt)
}

func TestQueryBusUnsupportedQuery(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{})
	_, err := qb.Query(context.Background(), unsupportedQuery{})
	require.Error(t, err)
}

func TestQueryBusRespectsCancelledContext(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			t.Fatal("runtime open should not be called")
			return nil, nil
		},
	})

	_, err := qb.Query(ctx, GetRuntimeStatusQuery{})
	require.Error(t, err)
	require.ErrorIs(t, err, context.Canceled)
}

func TestQueryBusGetTrustSummaryLoadsLatestReport(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	trustDir := filepath.Join(repoRoot, ".asagiri", "trust", "trust-1")
	require.NoError(t, os.MkdirAll(trustDir, 0o755))
	report := map[string]any{
		"trust_id":     "trust-1",
		"generated_at": time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
		"confidence": map[string]any{
			"architecture":  0.82,
			"security":      0.71,
			"observability": 0.63,
			"regression":    0.78,
			"overall":       0.76,
		},
	}
	body, err := json.Marshal(report)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(trustDir, "report.json"), body, 0o644))

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	gotAny, err := qb.Query(context.Background(), GetTrustSummaryQuery{})
	require.NoError(t, err)

	got := gotAny.(TrustSummaryResult)
	require.InDelta(t, 0.76, got.Overall, 0.0001)
	require.Len(t, got.Dimensions, 4)
	require.Empty(t, got.Warning)
}

func TestQueryBusListActiveAgentsFromRuntimeEvents(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{
				events: []runtime.RuntimeEvent{
					{
						Type:      runtime.EventAgentStarted,
						FlowID:    "onboarding",
						CreatedAt: now,
						Payload: map[string]any{
							"role":      "implementer",
							"agent_ref": "cursor",
						},
					},
					{
						Type:      runtime.EventAgentCompleted,
						FlowID:    "onboarding",
						CreatedAt: now.Add(2 * time.Minute),
						Payload: map[string]any{
							"role":      "investigator",
							"agent_ref": "codex",
						},
					},
				},
			}, nil
		},
	})

	gotAny, err := qb.Query(context.Background(), ListActiveAgentsQuery{Limit: 20})
	require.NoError(t, err)
	got := gotAny.(ActiveAgentsResult)
	require.Len(t, got.Agents, 2)
	require.NotEmpty(t, got.Agents[0].Status)
}

func TestQueryBusMissionControlSnapshotWorkspaceFromRepoRoot(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repoRoot, "my-workspace"), 0o755))

	qb := NewQueryBus(Deps{
		RepoRoot: repoRoot,
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{}, nil
		},
		StateOpen: func(string) (stateStore, error) {
			return &mockStateStore{}, nil
		},
	})

	gotAny, err := qb.Query(context.Background(), GetMissionControlSnapshotQuery{})
	require.NoError(t, err)
	got := gotAny.(MissionControlSnapshotResult)
	require.Equal(t, filepath.Base(repoRoot), got.Workspace)
}

func TestQueryBusMissionControlSnapshotCollectsWarnings(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return nil, errors.New("runtime unavailable")
		},
		StateOpen: func(string) (stateStore, error) {
			return nil, errors.New("state unavailable")
		},
	})

	gotAny, err := qb.Query(context.Background(), GetMissionControlSnapshotQuery{})
	require.NoError(t, err)
	got := gotAny.(MissionControlSnapshotResult)
	require.Contains(t, got.Warnings, "runtime unavailable")
	require.Contains(t, got.Warnings, "state unavailable")
	require.Equal(t, "inactive", got.SessionStatus)
}

func TestQueryBusMissionControlSnapshotAggregatesData(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	statePath := filepath.Join(repoRoot, ".asagiri", "state.sqlite")
	require.NoError(t, os.MkdirAll(filepath.Dir(statePath), 0o755))

	st, err := sqlite.Open(statePath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = st.Close() })
	require.NoError(t, st.Migrate())
	require.NoError(t, st.CreateRun(&sqlite.Run{
		ID:      "run-1",
		Feature: "spec-ui",
		Status:  sqlite.StatusRunning,
	}))

	qb := NewQueryBus(Deps{
		RepoRoot:    repoRoot,
		StateDBPath: statePath,
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{
				status: runtime.DaemonStatus{
					Running:      true,
					Sessions:     2,
					FlowsActive:  1,
					QueuedEvents: 1,
				},
				events: []runtime.RuntimeEvent{
					{
						ID:        "ev-1",
						Type:      runtime.EventAgentStarted,
						FlowID:    "onboarding",
						CreatedAt: time.Now().UTC(),
						Payload: map[string]any{
							"role":      "implementer",
							"agent_ref": "cursor",
							"cost_eur":  0.42,
						},
					},
				},
			}, nil
		},
	})

	gotAny, err := qb.Query(context.Background(), GetMissionControlSnapshotQuery{})
	require.NoError(t, err)
	got := gotAny.(MissionControlSnapshotResult)
	require.Equal(t, "active", got.SessionStatus)
	require.NotEmpty(t, got.Runs)
	require.NotEmpty(t, got.Events)
	require.NotEmpty(t, got.ActiveAgents)
	require.GreaterOrEqual(t, got.CostTodayEUR, 0.42)
}

func TestQueryBusFlowAndGraphExplorer(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	repo := executiongraph.NewRepository(repoRoot)
	_, _, err := repo.Save(executiongraph.ExecutionGraph{
		ID:        "graph-2026-05-29-001",
		Product:   "workspace-saas",
		Flow:      "onboarding",
		Status:    executiongraph.GraphStatusRunning,
		CreatedAt: time.Date(2026, 5, 29, 11, 0, 0, 0, time.UTC).Format(time.RFC3339),
		Strategy: executiongraph.Strategy{
			MaxParallel: 1,
		},
		Nodes: []executiongraph.GraphNode{
			{ID: "investigate", Type: executiongraph.NodeTypeInvestigation, Title: "investigate", Status: executiongraph.NodeStatusSucceeded, Risk: executiongraph.RiskLevelLow},
			{ID: "implement", Type: executiongraph.NodeTypeImplementation, Title: "implement", Status: executiongraph.NodeStatusBlocked, Risk: executiongraph.RiskLevelHigh},
		},
		Edges: []executiongraph.GraphEdge{
			{From: "investigate", To: "implement", Type: executiongraph.EdgeTypeRequires},
		},
	})
	require.NoError(t, err)

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	flowAny, err := qb.Query(context.Background(), GetFlowExplorerQuery{FlowID: "onboarding"})
	require.NoError(t, err)
	flow := flowAny.(FlowExplorerResult)
	require.Equal(t, "onboarding", flow.FlowID)
	require.NotEmpty(t, flow.Steps)

	graphAny, err := qb.Query(context.Background(), GetGraphExplorerQuery{FlowID: "onboarding"})
	require.NoError(t, err)
	graph := graphAny.(GraphExplorerResult)
	require.Equal(t, "graph-2026-05-29-001", graph.GraphID)
	require.Len(t, graph.Nodes, 2)
	blockedBy := map[string][]string{}
	for _, node := range graph.Nodes {
		blockedBy[node.ID] = node.BlockedBy
	}
	require.Contains(t, blockedBy["implement"], "investigate")
}

func TestQueryBusExplorerEmptyWhenNoGraph(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{RepoRoot: t.TempDir()})

	flowAny, err := qb.Query(context.Background(), GetFlowExplorerQuery{FlowID: "onboarding"})
	require.NoError(t, err)
	flow := flowAny.(FlowExplorerResult)
	require.Equal(t, "onboarding", flow.FlowID)
	require.Empty(t, flow.Steps)
	require.Equal(t, "flow graph unavailable", flow.Warning)

	graphAny, err := qb.Query(context.Background(), GetGraphExplorerQuery{FlowID: "onboarding"})
	require.NoError(t, err)
	graph := graphAny.(GraphExplorerResult)
	require.Equal(t, "onboarding", graph.FlowID)
	require.Empty(t, graph.Nodes)
	require.Equal(t, "flow graph unavailable", graph.Warning)
}

func TestQueryBusSearchKnowledgeUnavailable(t *testing.T) {
	t.Parallel()

	qb := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	gotAny, err := qb.Query(context.Background(), SearchKnowledgeQuery{Query: "invite_member"})
	require.NoError(t, err)
	got := gotAny.(KnowledgeSearchResult)
	require.Equal(t, "invite_member", got.Query)
	require.Empty(t, got.Matches)
	require.Equal(t, "knowledge graph unavailable", got.Warning)
}

func TestQueryBusSearchKnowledgeMatchesNodes(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	store, err := knowledge.OpenStore(repoRoot)
	require.NoError(t, err)
	ctx := context.Background()
	require.NoError(t, store.UpsertNode(ctx, knowledge.GraphNode{
		ID:         "action:invite_member",
		Type:       knowledge.NodeTypeAction,
		Name:       "invite_member",
		Path:       "application/internal/invitations/service.go",
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "handlers_readonly_test"},
		Confidence: 0.96,
	}))
	require.NoError(t, store.UpsertNode(ctx, knowledge.GraphNode{
		ID:         "event:member.invited",
		Type:       knowledge.NodeTypeEvent,
		Name:       "member.invited",
		Source:     knowledge.GraphSource{Kind: "fixture", Path: "handlers_readonly_test"},
		Confidence: 0.92,
	}))
	require.NoError(t, store.Close())

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	gotAny, err := qb.Query(ctx, SearchKnowledgeQuery{Query: "invite_member", Limit: 5})
	require.NoError(t, err)
	got := gotAny.(KnowledgeSearchResult)
	require.Equal(t, "invite_member", got.Query)
	require.NotEmpty(t, got.Matches)
	require.Equal(t, "action:invite_member", got.Matches[0].ID)
	require.Contains(t, got.Matches[0].CLIEquivalent, `asa knowledge query "invite_member"`)
}

func TestQueryBusTrustExplorerAndExplain(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	trustDir := filepath.Join(repoRoot, ".asagiri", "trust", "trust-001")
	require.NoError(t, os.MkdirAll(trustDir, 0o755))
	report := map[string]any{
		"trust_id":      "trust-001",
		"generated_at":  time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC).Format(time.RFC3339Nano),
		"residual_risk": "medium",
		"gate": map[string]any{
			"status": "warn",
			"reason": "security confidence below target",
		},
		"confidence": map[string]any{
			"architecture":   0.82,
			"implementation": 0.67,
			"security":       0.71,
			"observability":  0.63,
			"regression":     0.78,
			"overall":        0.74,
		},
		"checks": []map[string]any{
			{
				"id":         "security-01",
				"name":       "security",
				"type":       "security",
				"status":     "warn",
				"confidence": 0.71,
				"findings": []map[string]any{
					{"severity": "warning", "category": "security.flow", "message": "no retry validation for invite_member"},
				},
				"evidence": []map[string]any{
					{"kind": "trace", "source": "invite_member", "summary": "security.flow check report"},
				},
			},
		},
	}
	body, err := json.Marshal(report)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(trustDir, "report.json"), body, 0o644))

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	trustAny, err := qb.Query(context.Background(), GetTrustExplorerQuery{})
	require.NoError(t, err)
	trustRes := trustAny.(TrustExplorerResult)
	require.InDelta(t, 0.74, trustRes.Overall, 0.001)
	require.Equal(t, "warn", trustRes.GateStatus)
	require.NotEmpty(t, trustRes.Dimensions)

	explainAny, err := qb.Query(context.Background(), GetExplainQuery{Subject: "security confidence"})
	require.NoError(t, err)
	explainRes := explainAny.(ExplainResult)
	require.Equal(t, "security confidence", explainRes.Subject)
	require.NotEmpty(t, explainRes.Reasons)
	require.NotEmpty(t, explainRes.Evidence)
}

func TestQueryBusAgentTheatreDetails(t *testing.T) {
	t.Parallel()

	now := time.Date(2026, 5, 29, 12, 0, 0, 0, time.UTC)
	qb := NewQueryBus(Deps{
		RuntimeOpen: func(string) (runtimeStore, error) {
			return &mockRuntimeStore{
				events: []runtime.RuntimeEvent{
					{
						Type:      runtime.EventAgentStarted,
						CreatedAt: now,
						Payload: map[string]any{
							"role":      "implementer",
							"agent_ref": "cursor",
							"task":      "Edit invitation service",
						},
					},
					{
						Type:      runtime.EventAgentCompleted,
						CreatedAt: now.Add(2 * time.Minute),
						Payload: map[string]any{
							"role":             "implementer",
							"agent_ref":        "cursor",
							"files_active":     12,
							"tokens_estimated": 4200,
							"cost_eur":         0.09,
							"confidence":       0.78,
							"output":           "retry implemented",
							"hypothesis":       "missing retry in API client",
						},
					},
				},
			}, nil
		},
	})

	gotAny, err := qb.Query(context.Background(), GetAgentTheatreQuery{Limit: 20})
	require.NoError(t, err)
	got := gotAny.(AgentTheatreResult)
	require.Len(t, got.Agents, 1)
	require.Equal(t, "implementer", got.Agents[0].Role)
	require.Equal(t, "cursor", got.Agents[0].AgentRef)
	require.Equal(t, 12, got.Agents[0].FilesActive)
	require.InDelta(t, 0.09, got.Agents[0].CostEUR, 0.0001)
	require.InDelta(t, 0.78, got.Agents[0].Confidence, 0.0001)
	require.Greater(t, got.Agents[0].Duration, time.Duration(0))
}

func TestQueryBusReplayPackage(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	replayID := "replay-2026-05-29-test001"
	replayDir := filepath.Join(repoRoot, ".asagiri", "replays", replayID)
	require.NoError(t, os.MkdirAll(filepath.Join(replayDir, "runtime"), 0o755))
	manifest := []byte("id: replay-2026-05-29-test001\ncreated_at: 2026-05-29T10:00:00Z\nrepo:\n  branch: feature/spec-ui\n  commit: abcdef123456\nruntime:\n  runtime_mode: simulation\nartifacts:\n  - graph/execution-graph.json\n")
	require.NoError(t, os.WriteFile(filepath.Join(replayDir, "replay.yaml"), manifest, 0o644))
	events := strings.Join([]string{
		`{"type":"investigation.started","created_at":"2026-05-29T10:12:00Z"}`,
		`{"type":"graph.generated","created_at":"2026-05-29T10:14:00Z","payload":{"artifact":"graph/execution-graph.json"}}`,
	}, "\n")
	require.NoError(t, os.WriteFile(filepath.Join(replayDir, "runtime", "events.jsonl"), []byte(events), 0o644))

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	gotAny, err := qb.Query(context.Background(), GetReplayPackageQuery{ReplayID: replayID, Limit: 20})
	require.NoError(t, err)
	got := gotAny.(ReplayPackageResult)
	require.Equal(t, replayID, got.ReplayID)
	require.Equal(t, "simulation", got.Mode)
	require.Len(t, got.Timeline, 2)
	require.Equal(t, "graph.generated", got.Timeline[1].Type)
}

func TestQueryBusPrototypePipeline(t *testing.T) {
	t.Parallel()

	repoRoot := t.TempDir()
	productRoot := filepath.Join(repoRoot, ".asagiri", "products", "workspace-saas")
	require.NoError(t, os.MkdirAll(filepath.Join(productRoot, "prototype"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(productRoot, "flows"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(productRoot, "prototype", "model.json"), []byte(`{"product":"workspace-saas"}`), 0o644))
	flow := `id: workspace-onboarding
title: Workspace onboarding
entry_screen: landing
steps:
  - id: step-1
    screen: landing
    action: click_get_started
    next: signup
    contract_ref: POST /api/workspaces
metrics:
  - onboarding_completion_rate
`
	require.NoError(t, os.WriteFile(filepath.Join(productRoot, "flows", "workspace-onboarding.flow.yaml"), []byte(flow), 0o644))

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	gotAny, err := qb.Query(context.Background(), GetPrototypePipelineQuery{Product: "workspace-saas", Limit: 10})
	require.NoError(t, err)
	got := gotAny.(PrototypePipelineResult)
	require.Equal(t, "workspace-saas", got.Product)
	require.Equal(t, "flow", got.PipelineStage)
	require.NotEmpty(t, got.FlowExtraction)
	require.Equal(t, "workspace-onboarding", got.FlowExtraction[0].FlowID)
}

type unsupportedQuery struct{}

func (unsupportedQuery) Name() string { return "Unsupported" }
