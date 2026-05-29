package executiongraph

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestScheduleSimpleLinear(t *testing.T) {
	graph := loadScheduleFixture(t, "simple-linear")
	sched, err := (DefaultScheduler{}).Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.NoError(t, err)

	require.Equal(t, graph.ID, sched.GraphID)
	require.Equal(t, [][]string{
		{"investigate-onboarding"},
		{"implement-workspace-create"},
		{"verify-onboarding-flow"},
	}, sched.ParallelGroups)

	require.Len(t, sched.Blocked, 2)
	require.Equal(t, BlockedNode{
		NodeID:  "implement-workspace-create",
		WaitFor: []string{"investigate-onboarding"},
	}, sched.Blocked[0])
	require.Equal(t, BlockedNode{
		NodeID:  "verify-onboarding-flow",
		WaitFor: []string{"implement-workspace-create"},
	}, sched.Blocked[1])
}

func TestScheduleParallelIndependent(t *testing.T) {
	graph := loadScheduleFixture(t, "parallel-independent")
	sched, err := (DefaultScheduler{}).Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.NoError(t, err)

	require.Equal(t, [][]string{
		{"investigate-onboarding"},
		{"implement-ui-onboarding", "implement-workspace-create"},
		{"verify-onboarding-flow"},
	}, sched.ParallelGroups)

	var parallelBlocked *BlockedNode
	for i := range sched.Blocked {
		if sched.Blocked[i].NodeID == "verify-onboarding-flow" {
			parallelBlocked = &sched.Blocked[i]
			break
		}
	}
	require.NotNil(t, parallelBlocked)
	require.ElementsMatch(t,
		[]string{"implement-ui-onboarding", "implement-workspace-create"},
		parallelBlocked.WaitFor,
	)
}

func TestScheduleFileConflictBlocksParallelism(t *testing.T) {
	graph := ExecutionGraph{
		ID:      "graph-2026-05-29-aabbccdd",
		Product: "workspace-saas",
		Status:  GraphStatusPlanned,
		Strategy: Strategy{
			MaxParallel: 2,
		},
		Nodes: []GraphNode{
			{ID: "investigate-onboarding", Type: NodeTypeInvestigation},
			{ID: "implement-a", Type: NodeTypeImplementation},
			{ID: "implement-b", Type: NodeTypeImplementation},
			{ID: "implement-c", Type: NodeTypeImplementation},
		},
		Edges: []GraphEdge{
			{From: "investigate-onboarding", To: "implement-a", Type: EdgeTypeProducesContextFor},
			{From: "investigate-onboarding", To: "implement-b", Type: EdgeTypeProducesContextFor},
			{From: "investigate-onboarding", To: "implement-c", Type: EdgeTypeProducesContextFor},
			{From: "implement-a", To: "implement-b", Type: EdgeTypeMustRunAfter, Reason: "tasks share touched file scope; no parallel execution"},
		},
	}
	require.NoError(t, graph.Validate())

	sched, err := (DefaultScheduler{}).Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.NoError(t, err)

	require.Equal(t, [][]string{
		{"investigate-onboarding"},
		{"implement-a", "implement-c"},
		{"implement-b"},
	}, sched.ParallelGroups)

	var bBlocked *BlockedNode
	for i := range sched.Blocked {
		if sched.Blocked[i].NodeID == "implement-b" {
			bBlocked = &sched.Blocked[i]
			break
		}
	}
	require.NotNil(t, bBlocked)
	require.Equal(t, []string{"implement-a"}, bBlocked.WaitFor)
}

func TestScheduleCIModeForcesMaxParallelOne(t *testing.T) {
	graph := loadScheduleFixture(t, "parallel-independent")
	sched, err := (DefaultScheduler{}).Schedule(t.Context(), ScheduleRequest{
		Graph:  graph,
		CIMode: true,
	})
	require.NoError(t, err)

	require.Equal(t, [][]string{
		{"investigate-onboarding"},
		{"implement-ui-onboarding"},
		{"implement-workspace-create"},
		{"verify-onboarding-flow"},
	}, sched.ParallelGroups)
}

func TestScheduleRespectsGraphMaxParallel(t *testing.T) {
	graph := loadScheduleFixture(t, "parallel-independent")
	require.Equal(t, 2, graph.Strategy.MaxParallel)

	sched, err := (DefaultScheduler{}).Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.NoError(t, err)
	require.Len(t, sched.ParallelGroups[1], 2)
}

func TestScheduleDetectsCycle(t *testing.T) {
	graph := ExecutionGraph{
		ID:      "graph-2026-05-29-ccddeeff",
		Product: "p",
		Status:  GraphStatusPlanned,
		Strategy: Strategy{
			MaxParallel: 1,
		},
		Nodes: []GraphNode{
			{ID: "a", Type: NodeTypeImplementation},
			{ID: "b", Type: NodeTypeImplementation},
		},
		Edges: []GraphEdge{
			{From: "a", To: "b", Type: EdgeTypeRequires},
			{From: "b", To: "a", Type: EdgeTypeRequires},
		},
	}
	require.NoError(t, graph.Validate())

	_, err := (DefaultScheduler{}).Schedule(t.Context(), ScheduleRequest{Graph: graph})
	require.Error(t, err)
	require.ErrorIs(t, err, ErrCycleDetected)
}

func loadScheduleFixture(t *testing.T, scenario string) ExecutionGraph {
	t.Helper()
	graph, err := ParseYAML(readGoldenFixture(t, scenario))
	require.NoError(t, err)
	require.NoError(t, graph.Validate())
	return graph
}
