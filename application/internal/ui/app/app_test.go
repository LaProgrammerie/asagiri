package app

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestRunQuitsOnQ(t *testing.T) {
	var out bytes.Buffer
	err := Run(context.Background(), Options{
		In:            bytes.NewBufferString("q"),
		Out:           &out,
		Config:        config.UIConfig{Theme: "asagiri-dark", Animations: true, RefreshIntervalMs: 200},
		InitialScreen: ScreenMission,
	})
	if err != nil {
		t.Fatalf("run: %v", err)
	}
}

func TestViewMissionFrameGolden(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			Animations:         true,
			CompactThreshold:   100,
			ShowCLIEquivalents: true,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 140
	m.height = 32
	m.now = func() time.Time { return fixed }
	m.snapshot = bus.MissionControlSnapshotResult{
		Workspace:     "workspace-saas",
		Branch:        "onboarding-v2",
		SessionStatus: "active",
		Runtime: bus.RuntimeStatusResult{
			Status: runtime.DaemonStatus{
				Running:      true,
				Sessions:     3,
				QueuedEvents: 2,
				FlowsActive:  1,
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
		Flow: bus.FlowGraphResult{
			FlowID: "onboarding",
			Steps: []bus.FlowGraphStep{
				{ID: "create_workspace", Label: "create_workspace", Status: "succeeded"},
				{ID: "invite_member", Label: "invite_member", Status: "running"},
				{ID: "accept_invite", Label: "accept_invite", Status: "pending"},
			},
		},
		ActiveAgents: []bus.ActiveAgentSummary{
			{Role: "investigator", AgentRef: "codex", Status: "done"},
			{Role: "implementer", AgentRef: "cursor", Status: "running"},
		},
		CostTodayEUR: 0.42,
		CostMonthEUR: 6.10,
		UpdatedAt:    fixed,
	}
	m.snapshot.Runs = []bus.RunSummary{
		{ID: "run-1", Feature: "spec-ui", Status: "running"},
		{ID: "run-2", Feature: "specv3", Status: "completed"},
	}
	m.snapshot.Events = []bus.EventSummary{
		{Type: "runtime.started", CreatedAt: fixed.Add(-2 * time.Minute)},
		{Type: "session.created", CreatedAt: fixed.Add(-1 * time.Minute)},
	}

	got := stripANSI(m.View())
	golden := filepath.Join("testdata", "mission_frame.txt")
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

func TestViewDashboardFrameGolden(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			Animations:         true,
			CompactThreshold:   100,
			ShowCLIEquivalents: true,
		},
		InitialScreen: ScreenDashboard,
	})
	m.width = 140
	m.height = 32
	m.now = func() time.Time { return fixed }
	m.snapshot = bus.MissionControlSnapshotResult{
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
			{Role: "investigator", AgentRef: "codex", Status: "done"},
			{Role: "implementer", AgentRef: "cursor", Status: "running"},
		},
		Flow: bus.FlowGraphResult{
			FlowID: "onboarding",
			Steps: []bus.FlowGraphStep{
				{ID: "create_workspace", Label: "create_workspace", Status: "succeeded"},
				{ID: "invite_member", Label: "invite_member", Status: "running"},
			},
		},
		Runs: []bus.RunSummary{
			{ID: "run-1", Feature: "spec-ui", Status: "running"},
			{ID: "run-2", Feature: "specv3", Status: "completed"},
		},
		Events: []bus.EventSummary{
			{Type: "investigation.completed", CreatedAt: fixed.Add(-2 * time.Minute)},
			{Type: "graph.generated", CreatedAt: fixed.Add(-1 * time.Minute)},
		},
		CostTodayEUR: 0.42,
		CostMonthEUR: 6.10,
		UpdatedAt:    fixed,
	}

	got := stripANSI(m.View())
	golden := filepath.Join("testdata", "dashboard_frame.txt")
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

func TestViewReplayFrameGolden(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			Animations:         true,
			CompactThreshold:   100,
			ShowCLIEquivalents: true,
		},
		InitialScreen: ScreenReplay,
		ReplayID:      "replay-001",
	})
	m.width = 140
	m.height = 32
	m.now = func() time.Time { return fixed }
	m.snapshot = bus.MissionControlSnapshotResult{
		Replay: bus.ReplayPackageResult{
			ReplayID:   "replay-001",
			CreatedAt:  fixed.Add(-5 * time.Minute),
			RepoBranch: "feature/spec-ui",
			RepoCommit: "abcdef1234567890",
			Mode:       "simulation",
			Artifacts:  []string{"graph/execution-graph.json", "runtime/events.jsonl"},
			Timeline: []bus.ReplayTimelineEvent{
				{Time: fixed.Add(-4 * time.Minute), Type: "investigation.started"},
				{Time: fixed.Add(-3 * time.Minute), Type: "graph.generated", Artifact: "graph/execution-graph.json"},
			},
		},
	}

	got := stripANSI(m.View())
	golden := filepath.Join("testdata", "replay_frame.txt")
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

func TestViewPrototypeFrameGolden(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			Animations:         true,
			CompactThreshold:   100,
			ShowCLIEquivalents: true,
		},
		InitialScreen:    ScreenPrototype,
		PrototypeProduct: "workspace-saas",
	})
	m.width = 140
	m.height = 32
	m.snapshot = bus.MissionControlSnapshotResult{
		Prototype: bus.PrototypePipelineResult{
			Product:        "workspace-saas",
			WireframeTitle: "workspace saas",
			WireframePath:  ".asagiri/products/workspace-saas/prototype/src/App.tsx",
			PipelineStage:  "flow",
			StagesDone:     []string{"wireframe", "journey", "flow"},
			Flow:           "workspace-onboarding",
			FlowExtraction: []bus.PrototypeFlowStep{
				{
					FlowID:   "workspace-onboarding",
					StepID:   "step-1",
					Action:   "click_get_started",
					Screen:   "landing",
					Next:     "signup",
					Contract: "POST /api/workspaces",
					Trust:    "pending",
					Metric:   "onboarding_completion_rate",
				},
			},
			SuggestedActions: []string{
				"asa contracts extract workspace-saas",
				"asa architecture derive workspace-saas",
			},
		},
	}

	got := stripANSI(m.View())
	golden := filepath.Join("testdata", "prototype_frame.txt")
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

func TestSnapshotUpdateIgnoresUpdatedAt(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	base := bus.MissionControlSnapshotResult{
		Workspace: "workspace-saas",
		UpdatedAt: fixed,
	}
	next := base
	next.UpdatedAt = fixed.Add(time.Second)

	require.True(t, snapshotContentEqual(base, next))
}

func TestSnapshotUpdateDetectsContentChange(t *testing.T) {
	fixed := time.Date(2026, 5, 29, 12, 34, 56, 0, time.UTC)
	base := bus.MissionControlSnapshotResult{
		Workspace: "workspace-saas",
		UpdatedAt: fixed,
	}
	changed := base
	changed.Workspace = "other-workspace"
	changed.UpdatedAt = fixed.Add(time.Second)

	require.False(t, snapshotContentEqual(base, changed))
}

func TestSnapshotMsgClearsErrorAndShowsWarning(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark", Animations: true},
		InitialScreen: ScreenMission,
	})
	m.lastError = "previous query failure"

	next, _ := m.Update(snapshotMsg{
		result: bus.MissionControlSnapshotResult{
			Warnings: []string{"runtime unavailable"},
		},
	})
	m = next.(model)

	require.Empty(t, m.lastError)
	require.Equal(t, "runtime unavailable", m.lastWarning)
}

func TestSnapshotMsgClearsWarningOnCleanSnapshot(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark", Animations: true},
		InitialScreen: ScreenMission,
	})
	m.lastWarning = "stale warning"

	next, _ := m.Update(snapshotMsg{
		result: bus.MissionControlSnapshotResult{},
	})
	m = next.(model)

	require.Empty(t, m.lastWarning)
}

func TestViewShowsWarningNotErrorInFooter(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:            "asagiri-dark",
			Animations:       true,
			CompactThreshold: 100,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 140
	m.height = 32
	m.lastWarning = "trust report unavailable"

	got := stripANSI(m.View())
	require.Contains(t, got, "warn: trust report unavailable")
	require.NotContains(t, got, "error: trust report unavailable")
}

func stripANSI(v string) string {
	return regexp.MustCompile(`\x1b\[[0-9;]*m`).ReplaceAllString(v, "")
}

func TestGlobalKeybindingsNavigateScreens(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:      "asagiri-dark",
			Animations: true,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 140
	m.height = 32

	m = updateWithKey(t, m, tea.KeyCtrlD)
	require.Equal(t, ScreenDashboard, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlG)
	require.Equal(t, ScreenGraph, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlF)
	require.Equal(t, ScreenFlow, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlL)
	require.Equal(t, ScreenLogs, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlE)
	require.Equal(t, ScreenExplain, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlR)
	require.Equal(t, ScreenReplay, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlK)
	require.Equal(t, ScreenKnowledge, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlT)
	require.Equal(t, ScreenTrust, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlM)
	require.Equal(t, ScreenMission, m.router.Current())
}

func TestPaletteOpensAndNavigates(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:      "asagiri-dark",
			Animations: true,
		},
		InitialScreen: ScreenMission,
	})

	m = updateWithKey(t, m, tea.KeyCtrlP)
	require.True(t, m.showPalette)

	m = updateWithRunes(t, m, "dash")
	m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	require.False(t, m.showPalette)
	require.Equal(t, ScreenDashboard, m.router.Current())
}

func TestSafetyConfirmationOnDestructivePaletteAction(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:                     "asagiri-dark",
			Animations:                true,
			ConfirmDestructiveActions: true,
		},
		InitialScreen: ScreenMission,
	})

	m = updateWithKey(t, m, tea.KeyCtrlP)
	m = updateWithRunes(t, m, "rollback")
	m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, m.confirmation)

	m = updateWithRunes(t, m, "y")
	require.Nil(t, m.confirmation)
	require.Contains(t, m.lastCommandResult, "confirmed")
}

func TestSafetyConfirmationEnterDoesNotConfirm(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:                     "asagiri-dark",
			Animations:                true,
			ConfirmDestructiveActions: true,
		},
		InitialScreen: ScreenMission,
	})

	m = updateWithKey(t, m, tea.KeyCtrlP)
	m = updateWithRunes(t, m, "rollback")
	m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, m.confirmation)

	m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, m.confirmation)
	require.NotContains(t, m.lastCommandResult, "confirmed")
}

func TestPaletteDispatchesCommand(t *testing.T) {
	cb := &recordingCommandBus{
		result: bus.CommandResult{
			Accepted:      true,
			Message:       "trust verification passed",
			CLIEquivalent: "asa verify trust onboarding",
		},
	}
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			Animations:         true,
			ShowCLIEquivalents: true,
		},
		InitialScreen: ScreenMission,
		CommandBus:    cb,
	})

	m = updateWithKey(t, m, tea.KeyCtrlP)
	m = updateWithRunes(t, m, "verify trust")
	next, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, cmd)
	require.Empty(t, cb.commands, "command bus must not run synchronously in Update")

	m = applyCmdChain(t, next.(model), cmd)

	require.Len(t, cb.commands, 1)
	require.Contains(t, m.lastCommandResult, "trust verification passed")
	require.Contains(t, m.lastCommandResult, "asa verify trust onboarding")
}

func TestTabFocusCyclesPanes(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:            "asagiri-dark",
			Animations:       true,
			CompactThreshold: 100,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 140
	m.height = 32

	require.Equal(t, "main", string(m.state.Focus))
	m = updateWithKey(t, m, tea.KeyTab)
	require.Equal(t, "side", string(m.state.Focus))
	m = updateWithKey(t, m, tea.KeyShiftTab)
	require.Equal(t, "main", string(m.state.Focus))
}

func TestHelpScreenShowsAccessibilityShortcuts(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:      "high-contrast",
			Animations: true,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 120
	m.height = 30

	m = updateWithRunes(t, m, "?")
	got := stripANSI(m.View())

	require.Contains(t, got, "Accessibility & key help")
	require.Contains(t, got, "high-contrast mode: active")
	require.Contains(t, got, "plain/json mode: available")
}

func TestNoAnimationModeUsesStaticGlyphs(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:      "asagiri-dark",
			Animations: false,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 140
	m.height = 32
	m.snapshot = bus.MissionControlSnapshotResult{
		ActiveAgents: []bus.ActiveAgentSummary{
			{Role: "implementer", AgentRef: "cursor", Status: "running"},
		},
		Flow: bus.FlowGraphResult{
			FlowID: "onboarding",
			Steps: []bus.FlowGraphStep{
				{Label: "invite_member", Status: "running"},
			},
		},
	}

	got := stripANSI(m.View())
	require.NotContains(t, got, "⠋")
	require.Contains(t, got, "• invite_member")
}

func TestMouseResizeAdjustsVerticalSplit(t *testing.T) {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:            "asagiri-dark",
			Animations:       true,
			Mouse:            true,
			CompactThreshold: 100,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 160
	m.height = 40
	initial := m.verticalSplit

	m = updateWithMouse(t, m, tea.MouseMsg{
		X:      80,
		Y:      10,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonLeft,
	})
	m = updateWithMouse(t, m, tea.MouseMsg{
		X:      110,
		Y:      10,
		Action: tea.MouseActionMotion,
		Button: tea.MouseButtonLeft,
	})
	m = updateWithMouse(t, m, tea.MouseMsg{
		X:      110,
		Y:      10,
		Action: tea.MouseActionRelease,
		Button: tea.MouseButtonLeft,
	})

	require.Greater(t, m.verticalSplit, initial)
	require.Contains(t, m.lastCommandResult, "pane resize:")
}

type recordingCommandBus struct {
	result   bus.CommandResult
	err      error
	commands []string
}

func (b *recordingCommandBus) Dispatch(_ context.Context, cmd bus.Command) (bus.CommandResult, error) {
	b.commands = append(b.commands, cmd.Name())
	if b.err != nil {
		return bus.CommandResult{}, b.err
	}
	return b.result, nil
}

func applyUpdate(t *testing.T, m model, msg tea.Msg) model {
	t.Helper()
	next, cmd := m.Update(msg)
	out, ok := next.(model)
	require.True(t, ok)
	return applyCmdChain(t, out, cmd)
}

func applyCmdChain(t *testing.T, m model, cmd tea.Cmd) model {
	t.Helper()
	for cmd != nil {
		next, nextCmd := m.Update(cmd())
		out, ok := next.(model)
		require.True(t, ok)
		m = out
		cmd = nextCmd
	}
	return m
}

func updateWithKey(t *testing.T, m model, key tea.KeyType) model {
	return applyUpdate(t, m, tea.KeyMsg{Type: key})
}

func updateWithRunes(t *testing.T, m model, chars string) model {
	t.Helper()
	for _, r := range chars {
		m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	return m
}

func updateWithMouse(t *testing.T, m model, msg tea.MouseMsg) model {
	return applyUpdate(t, m, msg)
}
