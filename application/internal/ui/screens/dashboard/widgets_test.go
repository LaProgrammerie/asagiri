package dashboard

import (
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRuntimeWidgetShowsStatus(t *testing.T) {
	w := RuntimeWidget(bus.MissionControlSnapshotResult{
		Runtime: bus.RuntimeStatusResult{
			Status: runtime.DaemonStatus{
				Running:      true,
				Sessions:     3,
				FlowsActive:  1,
				QueuedEvents: 2,
			},
		},
	}, true)

	require.Equal(t, "Runtime", w.Title())
	require.Equal(t, Size{Width: 32, Height: 5}, w.MinSize())
	view := w.View()
	require.Contains(t, view, "Status: running")
	require.Contains(t, view, "Sessions: 3")
	require.Contains(t, view, "Flows: 1")
	require.Contains(t, view, "Queue: 2")
}

func TestAgentWidgetEmptyState(t *testing.T) {
	w := AgentWidget(bus.MissionControlSnapshotResult{}, true)
	require.Equal(t, "No active agents", w.View())
}

func TestTrustWidgetEmptyState(t *testing.T) {
	w := TrustWidget(bus.MissionControlSnapshotResult{}, true)
	require.Equal(t, "No trust report", w.View())
}

func TestCostWidgetShowsCosts(t *testing.T) {
	w := CostWidget(bus.MissionControlSnapshotResult{
		CostTodayEUR: 0.42,
		CostMonthEUR: 6.10,
	}, true)
	view := w.View()
	require.Contains(t, view, "Today: €0.42")
	require.Contains(t, view, "Month: €6.10")
}

func TestFlowWidgetShowsSteps(t *testing.T) {
	w := FlowWidget(bus.MissionControlSnapshotResult{
		Flow: bus.FlowGraphResult{
			Steps: []bus.FlowGraphStep{
				{Label: "create_workspace", Status: "succeeded"},
				{Label: "invite_member", Status: "running"},
			},
		},
	}, true)
	view := w.View()
	require.Contains(t, view, "✓ create_workspace")
	require.Contains(t, view, "⠋ invite_member")
}

func TestFlowWidgetEmptyState(t *testing.T) {
	w := FlowWidget(bus.MissionControlSnapshotResult{}, true)
	require.Equal(t, "No active flow", w.View())
}

func TestEventWidgetEmptyState(t *testing.T) {
	w := EventWidget(bus.MissionControlSnapshotResult{}, true)
	require.Equal(t, "No events", w.View())
}

func TestEventWidgetShowsRecentEvents(t *testing.T) {
	now := time.Date(2026, 5, 29, 8, 14, 0, 0, time.UTC)
	w := EventWidget(bus.MissionControlSnapshotResult{
		Events: []bus.EventSummary{
			{Type: "investigation.completed", CreatedAt: now},
			{Type: "graph.generated", CreatedAt: now.Add(time.Minute)},
		},
	}, true)
	view := w.View()
	require.Contains(t, view, "08:14:00  investigation.completed")
	require.Contains(t, view, "08:15:00  graph.generated")
}

func TestProgressWidgetCompletedRatio(t *testing.T) {
	w := ProgressWidget(bus.MissionControlSnapshotResult{
		Runs: []bus.RunSummary{
			{ID: "run-1", Status: "completed"},
			{ID: "run-2", Status: "running"},
		},
	}, true)
	view := w.View()
	require.Contains(t, view, "Completed: 1/2")
	require.Contains(t, view, "50%")
}

func TestProgressWidgetEmptyState(t *testing.T) {
	w := ProgressWidget(bus.MissionControlSnapshotResult{}, true)
	require.Equal(t, "No runs", w.View())
}
