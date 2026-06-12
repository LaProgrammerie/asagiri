package app

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func newRunsModel() model {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{Theme: "asagiri-dark", CompactThreshold: 100},
	})
	m.width = 140
	m.height = 32
	m.router.Set(ScreenRuns)
	m.snapshot.Runs = []bus.RunSummary{
		{ID: "run-1", Feature: "cockpit", Status: "running"},
		{ID: "run-2", Feature: "spec-ui", Status: "completed"},
	}
	return m
}

func TestRunsDrillDownNavigates(t *testing.T) {
	cases := map[string]string{
		"t": ScreenTrust,
		"g": ScreenGraph,
		"r": ScreenReplay,
	}
	for key, want := range cases {
		m := newRunsModel()
		next, _ := m.updateRunsExplorerKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune(key)}, key)
		nm := next.(*model)
		require.Equal(t, want, nm.router.Current(), "key %q must navigate to %s", key, want)
	}
}

func TestRunsEnterOpensSelectedRun(t *testing.T) {
	m := newRunsModel()
	next, _ := m.updateRunsExplorerKey(tea.KeyMsg{Type: tea.KeyEnter}, "enter")
	nm := next.(*model)
	require.Contains(t, nm.lastCommandResult, "run-1")
}

func TestRunsListSelectionMovesCursor(t *testing.T) {
	m := newRunsModel()
	next, _ := m.updateRunsExplorerKey(tea.KeyMsg{Type: tea.KeyDown}, "down")
	nm := next.(*model)
	require.Equal(t, "run-2", nm.runsExplorer.SelectedRunID(nm.snapshot.Runs))
}

func TestRunsScreenIsExplorerInput(t *testing.T) {
	m := newRunsModel()
	require.True(t, m.explorerInputActive())
}
