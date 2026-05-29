package mission

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func TestRenderGolden(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	vm := ViewModel{
		RuntimeStatus: "running (3 sessions)",
		Runs: []bus.RunSummary{
			{ID: "run-1", Feature: "spec-ui", Status: "running"},
			{ID: "run-2", Feature: "specv3", Status: "completed"},
		},
		Events: []bus.EventSummary{
			{Type: "runtime.started", CreatedAt: fixed.Add(-2 * time.Minute)},
			{Type: "session.created", CreatedAt: fixed.Add(-1 * time.Minute)},
		},
		QueuedEvents: 2,
		Now:          fixed,
	}

	got := Render(vm)
	golden := filepath.Join("testdata", "mission_control.txt")

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

func TestRenderEmptyState(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 8, 0, 0, 0, time.UTC)
	got := Render(ViewModel{
		RuntimeStatus: "stopped",
		Now:           fixed,
	})

	require.Contains(t, got, "Runtime: stopped")
	require.Contains(t, got, "- none\n")
	require.Contains(t, got, fixed.Format(time.RFC3339))
}

func TestRenderShowsWarning(t *testing.T) {
	got := Render(ViewModel{
		RuntimeStatus: "degraded",
		Warning:       "runtime unavailable",
		Now:           time.Date(2026, 5, 29, 8, 0, 0, 0, time.UTC),
	})

	require.Contains(t, got, "Warnings:")
	require.Contains(t, got, "runtime unavailable")
}

func TestRenderHeaderShowsWorkspaceBranchSession(t *testing.T) {
	got := RenderHeader(ViewModel{
		Workspace:     "workspace-saas",
		Branch:        "onboarding-v2",
		RuntimeStatus: "running",
		SessionStatus: "active",
	})

	require.Contains(t, got, "Workspace: workspace-saas")
	require.Contains(t, got, "Branch: onboarding-v2")
	require.Contains(t, got, "Runtime: running")
	require.Contains(t, got, "Session: active")
}

func TestRenderShowsTrustDimensionsAndOverall(t *testing.T) {
	got := RenderTrustPane(ViewModel{
		Trust: bus.TrustSummaryResult{
			Overall: 0.76,
			Dimensions: []bus.TrustDimensionScore{
				{Label: "Architecture", Score: 0.82},
				{Label: "Security", Score: 0.71},
			},
		},
	})

	require.Contains(t, got, "Architecture")
	require.Contains(t, got, "82%")
	require.Contains(t, got, "Security")
	require.Contains(t, got, "71%")
	require.Contains(t, got, "Overall")
	require.Contains(t, got, "76%")
}

func TestRenderActiveFlowShowsStepsWithGlyphs(t *testing.T) {
	got := RenderActiveFlowPane(ViewModel{
		Flow: bus.FlowGraphResult{
			FlowID: "onboarding",
			Steps: []bus.FlowGraphStep{
				{ID: "create_workspace", Label: "create_workspace", Status: "succeeded"},
				{ID: "invite_member", Label: "invite_member", Status: "running"},
				{ID: "accept_invite", Label: "accept_invite", Status: "pending"},
			},
		},
	})

	require.Contains(t, got, "onboarding")
	require.Contains(t, got, "✓ create_workspace")
	require.Contains(t, got, "⠋ invite_member")
	require.Contains(t, got, "○ accept_invite")
}

func TestRenderAgentTheatreShowsAgents(t *testing.T) {
	got := RenderAgentTheatrePane(ViewModel{
		ActiveAgents: []bus.ActiveAgentSummary{
			{Role: "investigator", AgentRef: "codex", Status: "done"},
			{Role: "implementer", AgentRef: "cursor", Status: "running"},
		},
	})

	require.Contains(t, got, "investigator ✓ codex")
	require.Contains(t, got, "implementer ⠋ cursor")
}

func TestRenderEventsLimitedToFive(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 8, 0, 0, 0, time.UTC)
	events := make([]bus.EventSummary, 0, 7)
	for i := 0; i < 7; i++ {
		events = append(events, bus.EventSummary{
			Type:      "event." + string(rune('a'+i)),
			CreatedAt: fixed.Add(time.Duration(i) * time.Minute),
		})
	}

	got := RenderEventsPane(ViewModel{Events: events})
	require.NotContains(t, got, "event.g")
	require.Contains(t, got, "event.a")
	require.Contains(t, got, "event.e")
}

func TestRenderRuntimeShowsCostsAndRuns(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 8, 0, 0, 0, time.UTC)
	got := RenderRuntimeRuns(ViewModel{
		SessionStatus: "active",
		ActiveAgents: []bus.ActiveAgentSummary{
			{Role: "implementer", Status: "running"},
		},
		CostTodayEUR: 0.42,
		CostMonthEUR: 6.10,
		Runs: []bus.RunSummary{
			{ID: "run-1", Feature: "spec-ui", Status: "running"},
		},
		Now: fixed,
	})

	require.Contains(t, got, "Agents: 1")
	require.Contains(t, got, "Sessions: 1")
	require.Contains(t, got, "Cost today: €0.42")
	require.Contains(t, got, "Cost month: €6.10")
	require.Contains(t, got, "run-1  spec-ui  running")
}

func TestRenderRuntimeUsesQueuedEventsNotEventCount(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 8, 0, 0, 0, time.UTC)
	got := RenderRuntimeRuns(ViewModel{
		Events: []bus.EventSummary{
			{Type: "runtime.started", CreatedAt: fixed},
			{Type: "session.created", CreatedAt: fixed.Add(time.Minute)},
		},
		QueuedEvents: 5,
		Now:          fixed,
	})

	require.Contains(t, got, "Queue: 5")
	require.NotContains(t, got, "Queue: 2")
}
