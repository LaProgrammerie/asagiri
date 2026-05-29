package app

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestIntegrationMissionDashboardPaletteVerifyTrust(t *testing.T) {
	repoRoot := t.TempDir()
	recorder := &recordingCommandBus{
		result: bus.CommandResult{
			Accepted:      true,
			Message:       "trust verified",
			CLIEquivalent: "asa verify trust onboarding",
		},
	}
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			ShowCLIEquivalents: true,
			Mouse:              true,
		},
		InitialScreen: ScreenMission,
		CommandBus:    recorder,
		QueryBus:      bus.NewQueryBus(bus.Deps{RepoRoot: repoRoot}),
	})
	m.width = 140
	m.height = 40

	m = applyUpdate(t, m, snapshotMsg{
		result: bus.MissionControlSnapshotResult{
			Workspace: "workspace-saas",
			Runtime: bus.RuntimeStatusResult{
				Status: runtime.DaemonStatus{Running: true},
			},
		},
	})
	require.Equal(t, ScreenMission, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlD)
	require.Equal(t, ScreenDashboard, m.router.Current())

	m = updateWithKey(t, m, tea.KeyCtrlP)
	require.True(t, m.showPalette)

	m = updateWithRunes(t, m, "verify")
	entries := m.filteredPaletteEntries()
	require.NotEmpty(t, entries)

	m = updateWithKey(t, m, tea.KeyEnter)
	require.Contains(t, recorder.commands, "VerifyTrust")
	require.Contains(t, m.lastCommandResult, "asa verify trust")
}

func TestIntegrationExplainUsesFocusContext(t *testing.T) {
	repoRoot := t.TempDir()
	qb := bus.NewQueryBus(bus.Deps{RepoRoot: repoRoot})
	m := newModel(context.Background(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark"},
		InitialScreen: ScreenGraph,
		QueryBus:      qb,
	})
	m.width = 120
	m.height = 32
	m.openExplainForFocus(bus.FocusKindGraphNode, "invite_member", "onboarding")
	require.Equal(t, ScreenExplain, m.router.Current())
	require.Equal(t, bus.FocusKindGraphNode, m.state.FocusContext.Kind)

	raw, err := qb.Query(context.Background(), bus.GetExplainQuery{
		Subject: "invite_member",
		Context: m.explainContext(),
	})
	require.NoError(t, err)
	explain, ok := raw.(bus.ExplainResult)
	require.True(t, ok)
	require.Contains(t, explain.Question, "blocked")
	require.NotEmpty(t, explain.SupportedQuestions)
}
