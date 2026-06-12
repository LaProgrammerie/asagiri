package app

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/require"
)

func TestPaletteFilterMatchesExplorerEntries(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark"},
		InitialScreen: ScreenMission,
	})
	m.showPalette = true
	m.refreshPaletteEntries()

	cases := []struct {
		query string
		id    string
	}{
		{query: "graph explorer", id: "nav.graph"},
		{query: "flow explorer", id: "nav.flow"},
		{query: "knowledge", id: "nav.knowledge"},
		{query: "trust explorer", id: "nav.trust"},
		{query: "asa explain", id: "nav.explain"},
		{query: "flow open onboarding", id: "flow.open-onboarding"},
	}

	for _, tc := range cases {
		m.paletteQuery = tc.query
		entries := m.filteredPaletteEntries()
		require.Len(t, entries, 1, tc.query)
		require.Equal(t, tc.id, entries[0].ID, tc.query)
	}
}

func TestPaletteFilterMatchesTitleAndCLI(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark"},
		InitialScreen: ScreenMission,
	})
	m.showPalette = true
	m.refreshPaletteEntries()

	m.paletteQuery = "verify trust"
	require.Len(t, m.filteredPaletteEntries(), 1)
	require.Equal(t, "cmd.verify-trust", m.filteredPaletteEntries()[0].ID)

	m.paletteQuery = "asa dashboard"
	require.Len(t, m.filteredPaletteEntries(), 1)
	require.Equal(t, "nav.dashboard", m.filteredPaletteEntries()[0].ID)
}

func TestPaletteFilterNoResults(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark"},
		InitialScreen: ScreenMission,
	})
	m.showPalette = true
	m.refreshPaletteEntries()
	m.paletteQuery = "zzzz-not-found"

	require.Empty(t, m.filteredPaletteEntries())
	require.Contains(t, m.renderPalette(), "no results")
}

func TestPaletteCursorWraps(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark"},
		InitialScreen: ScreenMission,
	})
	m.showPalette = true
	m.refreshPaletteEntries()
	m.paletteQuery = "nav"
	entries := m.filteredPaletteEntries()
	require.Greater(t, len(entries), 1)

	m.paletteCursor = 0
	m.movePaletteCursor(-1)
	require.Equal(t, len(entries)-1, m.paletteCursor)

	m.movePaletteCursor(1)
	require.Equal(t, 0, m.paletteCursor)
}

func TestPaletteBackspaceUpdatesQuery(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config:        config.UIConfig{Theme: "asagiri-dark"},
		InitialScreen: ScreenMission,
	})
	m = updateWithKey(t, m, tea.KeyCtrlP)
	m = updateWithRunes(t, m, "dash")
	require.Equal(t, "dash", m.paletteQuery)

	next, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = next.(model)
	require.Equal(t, "das", m.paletteQuery)
	require.Equal(t, 0, m.paletteCursor)
}

func TestSafetyConfirmationCancel(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config: config.UIConfig{
			Theme:                     "asagiri-dark",
			ConfirmDestructiveActions: true,
		},
		InitialScreen: ScreenMission,
	})

	m = updateWithKey(t, m, tea.KeyCtrlP)
	m = updateWithRunes(t, m, "rollback")
	m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyEnter})
	require.NotNil(t, m.confirmation)

	m = updateWithRunes(t, m, "n")
	require.Nil(t, m.confirmation)
	require.Equal(t, "action cancelled", m.lastCommandResult)
}

func TestSafetyConfirmationSkipsWhenDisabled(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config: config.UIConfig{
			Theme:                     "asagiri-dark",
			ConfirmDestructiveActions: false,
		},
		InitialScreen: ScreenMission,
	})

	m.commandBus = bus.NewCommandBus(bus.Deps{
		RepoRoot: t.TempDir(),
		GraphRollback: func(_ context.Context, _ bus.Deps, cmd bus.GraphRollbackCommand) (bus.CommandResult, error) {
			return bus.CommandResult{
				Accepted:      true,
				Message:       "graph graph-001 rolled back (1 nodes)",
				CLIEquivalent: cmd.CLIEquivalent(),
			}, nil
		},
	})

	m = updateWithKey(t, m, tea.KeyCtrlP)
	m = updateWithRunes(t, m, "rollback")
	m = applyUpdate(t, m, tea.KeyMsg{Type: tea.KeyEnter})

	require.Nil(t, m.confirmation)
	require.Contains(t, m.lastCommandResult, "rolled back")
}

func TestViewShowsSafetyConfirmationDetails(t *testing.T) {
	t.Parallel()

	m := newModel(t.Context(), Options{
		Config: config.UIConfig{
			Theme:              "asagiri-dark",
			ShowCLIEquivalents: true,
		},
		InitialScreen: ScreenMission,
	})
	m.width = 140
	m.height = 32
	m.confirmation = &safetyConfirmation{
		Title: "You are about to rollback graph-001.",
		Impact: []string{
			"2 worktrees will be impacted",
			"4 generated reports may become stale",
		},
		CLIEquivalent:  "asa graph rollback graph-001",
		RollbackPolicy: "Rollback can be replayed manually after confirmation.",
		ActionID:       "safe.graph-rollback",
	}

	got := stripANSI(m.View())
	require.Contains(t, got, "Safety confirmation")
	require.Contains(t, got, "You are about to rollback graph-001.")
	require.Contains(t, got, "Impacted:")
	require.Contains(t, got, "2 worktrees will be impacted")
	require.Contains(t, got, "CLI equivalent:")
	require.Contains(t, got, "asa graph rollback graph-001")
	require.Contains(t, got, "Rollback:")
	require.Contains(t, got, "Proceed? [y/N]")
}
