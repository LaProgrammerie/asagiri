package bus

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestDispatchGraphRollback(t *testing.T) {
	repoRoot := t.TempDir()
	graph := executiongraph.ExecutionGraph{
		ID:        "graph-2026-05-29-abcdef03",
		Product:   "workspace-saas",
		Status:    executiongraph.GraphStatusRunning,
		CreatedAt: "2026-05-29T10:00:00Z",
		Strategy:  executiongraph.Strategy{MaxParallel: 1},
		Nodes: []executiongraph.GraphNode{
			{ID: "n1", Type: executiongraph.NodeTypeImplementation, Status: executiongraph.NodeStatusRunning, Risk: executiongraph.RiskLevelLow},
		},
	}
	repo := executiongraph.NewRepository(repoRoot)
	_, _, err := repo.Save(graph)
	require.NoError(t, err)

	cb := NewCommandBus(Deps{RepoRoot: repoRoot})
	res, err := cb.Dispatch(context.Background(), GraphRollbackCommand{GraphID: graph.ID})
	require.NoError(t, err)
	require.True(t, res.Accepted)
	require.Equal(t, "asa graph rollback "+graph.ID, res.CLIEquivalent)
}

func TestGetGraphRollbackImpactQuery(t *testing.T) {
	repoRoot := t.TempDir()
	graph := executiongraph.ExecutionGraph{
		ID:        "graph-2026-05-29-abcdef04",
		Product:   "workspace-saas",
		Status:    executiongraph.GraphStatusRunning,
		CreatedAt: "2026-05-29T10:00:00Z",
		Strategy:  executiongraph.Strategy{MaxParallel: 1},
		Nodes: []executiongraph.GraphNode{
			{ID: "n1", Type: executiongraph.NodeTypeImplementation, Status: executiongraph.NodeStatusRunning, Risk: executiongraph.RiskLevelHigh, RollbackStrategy: executiongraph.RollbackStrategyWorktreeReset},
		},
		Rollback: &executiongraph.RollbackPlan{Strategy: executiongraph.RollbackStrategyWorktreeReset, PreserveReports: true},
	}
	repo := executiongraph.NewRepository(repoRoot)
	_, _, err := repo.Save(graph)
	require.NoError(t, err)

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	raw, err := qb.Query(context.Background(), GetGraphRollbackImpactQuery{GraphID: graph.ID})
	require.NoError(t, err)
	impact, ok := raw.(GraphRollbackImpactResult)
	require.True(t, ok)
	require.Contains(t, impact.Title, graph.ID)
	require.NotEmpty(t, impact.ImpactLines)
	require.True(t, impact.CanRollback)
}

func TestGetPaletteEntriesIncludesBuildKnowledge(t *testing.T) {
	qb := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	raw, err := qb.Query(context.Background(), GetPaletteEntriesQuery{Screen: "mission", Limit: 200})
	require.NoError(t, err)
	res, ok := raw.(PaletteEntriesResult)
	require.True(t, ok)
	found := false
	for _, entry := range res.Entries {
		if entry.ActionID == "cmd.build-knowledge" {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestDispatchExportEvents(t *testing.T) {
	repoRoot := t.TempDir()
	store, err := runtime.Open(repoRoot)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	_, err = store.EmitEvent("graph.generated", "tests", "", "", nil)
	require.NoError(t, err)

	cb := NewCommandBus(Deps{RepoRoot: repoRoot})
	res, err := cb.Dispatch(context.Background(), ExportEventsCommand{TypeFilter: "graph"})
	require.NoError(t, err)
	require.True(t, res.Accepted)
	require.Contains(t, res.Message, "exported")
}

func TestDispatchBuildKnowledgeGraphDryRun(t *testing.T) {
	cb := NewCommandBus(Deps{
		RepoRoot: t.TempDir(),
		DryRun:   true,
		Config:   &config.Config{},
	})
	res, err := cb.Dispatch(context.Background(), BuildKnowledgeGraphCommand{Incremental: true})
	require.NoError(t, err)
	require.True(t, res.Accepted)
	require.Contains(t, res.Message, "dry-run")
	require.Equal(t, "asa knowledge build", res.CLIEquivalent)
}

func TestDispatchReplayRunRequiresRunID(t *testing.T) {
	cb := NewCommandBus(Deps{RepoRoot: t.TempDir(), Config: &config.Config{}})
	_, err := cb.Dispatch(context.Background(), ReplayRunCommand{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "replay id required")
}

func TestCommandBusReplayAndKnowledgeCLIEquivalent(t *testing.T) {
	cb := NewCommandBus(Deps{
		BuildKnowledgeGraph: func(_ context.Context, _ Deps, cmd BuildKnowledgeGraphCommand) (CommandResult, error) {
			return CommandResult{Accepted: true, CLIEquivalent: cmd.CLIEquivalent()}, nil
		},
		ReplayRun: func(_ context.Context, _ Deps, cmd ReplayRunCommand) (CommandResult, error) {
			return CommandResult{Accepted: true, CLIEquivalent: cmd.CLIEquivalent()}, nil
		},
	})
	kRes, err := cb.Dispatch(context.Background(), BuildKnowledgeGraphCommand{})
	require.NoError(t, err)
	require.Equal(t, "asa knowledge build", kRes.CLIEquivalent)

	rRes, err := cb.Dispatch(context.Background(), ReplayRunCommand{RunID: "replay-001"})
	require.NoError(t, err)
	require.Equal(t, "asa replay run replay-001", rRes.CLIEquivalent)
}
