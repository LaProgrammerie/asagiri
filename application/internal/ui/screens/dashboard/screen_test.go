package dashboard

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

func TestWidgetsRender(t *testing.T) {
	snapshot := fixtureSnapshot()
	widgets := []Widget{
		RuntimeWidget(snapshot, true),
		AgentWidget(snapshot, true),
		TrustWidget(snapshot, true),
		CostWidget(snapshot, true),
		FlowWidget(snapshot, true),
		EventWidget(snapshot, true),
		ProgressWidget(snapshot, true),
	}
	for _, widget := range widgets {
		require.NotEmpty(t, widget.Title())
		require.NotEmpty(t, widget.View())
	}
}

func TestRenderCompactLayout(t *testing.T) {
	vm := ViewModel{
		Snapshot: fixtureSnapshot(),
		Theme:    theme.Default(),
		Width:    80,
		Animated: true,
	}
	got := Render(vm)
	require.NotContains(t, got, "╮╭", "compact layout must stack panels vertically")
	require.Contains(t, got, "Runtime")
	require.NotContains(t, got, "Agents")
}

func TestRenderWideLayout(t *testing.T) {
	vm := ViewModel{
		Snapshot: fixtureSnapshot(),
		Theme:    theme.Default(),
		Width:    140,
		Animated: true,
	}
	got := Render(vm)
	require.Contains(t, got, "╮╭", "wide layout must place panels side by side")
}

func TestRenderGolden(t *testing.T) {
	vm := ViewModel{
		Snapshot: fixtureSnapshot(),
		Theme:    theme.Default(),
		Width:    140,
		Animated: true,
	}
	got := Render(vm)
	golden := filepath.Join("testdata", "dashboard.txt")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
	}

	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", golden)
	}
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}

func fixtureSnapshot() bus.MissionControlSnapshotResult {
	now := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	return bus.MissionControlSnapshotResult{
		Runtime: bus.RuntimeStatusResult{
			Status: runtime.DaemonStatus{
				Running:      true,
				Sessions:     3,
				FlowsActive:  1,
				QueuedEvents: 2,
			},
		},
		Trust: bus.TrustSummaryResult{
			Overall: 0.76,
			Dimensions: []bus.TrustDimensionScore{
				{Label: "Architecture", Score: 0.82},
				{Label: "Security", Score: 0.71},
				{Label: "Observability", Score: 0.63},
				{Label: "Regression", Score: 0.78},
			},
		},
		ActiveAgents: []bus.ActiveAgentSummary{
			{Role: "investigator", AgentRef: "codex", Status: "done", UpdatedAt: now},
			{Role: "implementer", AgentRef: "cursor", Status: "running", UpdatedAt: now.Add(30 * time.Second)},
		},
		Flow: bus.FlowGraphResult{
			FlowID: "onboarding",
			Steps: []bus.FlowGraphStep{
				{ID: "create_workspace", Label: "create_workspace", Status: "succeeded"},
				{ID: "invite_member", Label: "invite_member", Status: "running"},
				{ID: "accept_invite", Label: "accept_invite", Status: "pending"},
			},
		},
		Events: []bus.EventSummary{
			{Type: "investigation.completed", CreatedAt: now.Add(-2 * time.Minute)},
			{Type: "graph.generated", CreatedAt: now.Add(-1 * time.Minute)},
		},
		Runs: []bus.RunSummary{
			{ID: "run-1", Feature: "spec-ui", Status: "running"},
			{ID: "run-2", Feature: "specv3", Status: "completed"},
		},
		CostTodayEUR: 0.42,
		CostMonthEUR: 6.10,
	}
}
