package executiongraph

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestEstimateNodeByTypeAndRisk(t *testing.T) {
	low := EstimateNode(GraphNode{Type: NodeTypeInvestigation, Risk: RiskLevelLow})
	require.Equal(t, 0.0, low.Cost)
	require.Equal(t, 30*time.Second, low.Duration)

	medium := EstimateNode(GraphNode{Type: NodeTypeImplementation, Risk: RiskLevelMedium})
	require.Equal(t, 0.08, medium.Cost)
	require.Equal(t, 4*time.Minute, medium.Duration)

	high := EstimateNode(GraphNode{Type: NodeTypeImplementation, Risk: RiskLevelHigh})
	require.Equal(t, 0.12, high.Cost)
	require.Equal(t, 6*time.Minute, high.Duration)
}

func TestApplyEstimatesFillsNodeFields(t *testing.T) {
	nodes := ApplyEstimates([]GraphNode{{
		ID:   "implement-x",
		Type: NodeTypeImplementation,
		Risk: RiskLevelHigh,
	}})
	require.Equal(t, 0.12, nodes[0].EstimatedCost)
	require.Equal(t, "6m", nodes[0].EstimatedDuration)
}

func TestEstimateGraphAggregatesParallelGroups(t *testing.T) {
	graph := ExecutionGraph{
		Strategy: Strategy{Budget: 1.00},
		Nodes: []GraphNode{
			{ID: "a", Type: NodeTypeInvestigation, Risk: RiskLevelLow, Agent: "local"},
			{ID: "b", Type: NodeTypeImplementation, Risk: RiskLevelMedium, Agent: "cursor"},
			{ID: "c", Type: NodeTypeImplementation, Risk: RiskLevelMedium, Agent: "cursor"},
		},
	}
	schedule := &ExecutionSchedule{
		ParallelGroups: [][]string{
			{"a"},
			{"b", "c"},
		},
	}

	est := EstimateGraph(graph, schedule)
	require.Equal(t, 3, est.Nodes)
	require.Equal(t, 2, est.ParallelGroups)
	require.Equal(t, "5m", est.EstimatedDuration)
	require.InDelta(t, 0.16, est.EstimatedCost, 0.001)
	require.Equal(t, RiskLevelMedium, est.HighestRisk)
	require.Equal(t, budgetStatusOK, est.BudgetStatus)
	require.InDelta(t, 0.16, est.ByAgent["cursor"], 0.001)
}

func TestEstimateGraphBudgetExceeded(t *testing.T) {
	graph := ExecutionGraph{
		Strategy: Strategy{Budget: 0.05},
		Nodes: []GraphNode{
			{ID: "a", Type: NodeTypeImplementation, Risk: RiskLevelHigh, Agent: "claude"},
		},
	}
	est := EstimateGraph(graph, nil)
	require.Equal(t, budgetStatusExceeded, est.BudgetStatus)
}

func TestFormatEstimateSummary(t *testing.T) {
	summary := FormatEstimateSummary(GraphEstimate{
		Nodes:             3,
		ParallelGroups:    2,
		EstimatedDuration: "5m",
		EstimatedCost:     0.16,
		HighestRisk:       RiskLevelMedium,
		BudgetStatus:      budgetStatusOK,
	})
	require.Contains(t, summary, "Nodes:              3")
	require.Contains(t, summary, "Estimated cost:     €0.16")
	require.Contains(t, summary, "Budget status:      OK")
}
