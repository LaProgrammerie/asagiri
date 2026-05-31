package app

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/stretchr/testify/require"
)

func newRailModel(width int) model {
	m := newModel(context.Background(), Options{
		Config: config.UIConfig{Theme: "asagiri-dark", CompactThreshold: 100},
	})
	m.width = width
	m.height = 32
	return m
}

func TestNavRailActiveFollowsRouter(t *testing.T) {
	m := newRailModel(140)
	m.router.Set(ScreenTrust)
	for _, it := range m.navRailItems() {
		if it.Label == "Trust" {
			require.True(t, it.Active, "Trust must be active")
		} else {
			require.False(t, it.Active, "%s must not be active", it.Label)
		}
	}
}

func TestNavRailCollapsesWhenCompact(t *testing.T) {
	m := newRailModel(80) // below compact threshold
	require.Equal(t, 0, m.railWidth())

	m.width = 140
	require.Equal(t, navRailWidth, m.railWidth())

	m.showPalette = true
	require.Equal(t, 0, m.railWidth(), "rail hides while palette is open")
}

func TestNavRailStateBadges(t *testing.T) {
	m := newRailModel(140)
	m.snapshot.Runs = []bus.RunSummary{{Status: "running"}, {Status: "completed"}, {Status: "running"}}
	m.snapshot.ActiveAgents = []bus.ActiveAgentSummary{{}, {}}
	m.snapshot.Trust = bus.TrustSummaryResult{Overall: 0.3}
	m.snapshot.Runtime.Status.QueuedEvents = 4

	badges := map[string]string{}
	for _, it := range m.navRailItems() {
		badges[it.Label] = it.Badge
	}
	require.Equal(t, "2", badges["Runs"])
	require.Equal(t, "2", badges["Agents"])
	require.Equal(t, "!", badges["Trust"])
	require.Equal(t, "4", badges["Logs"])
	require.Empty(t, badges["Mission"])
}

func TestRailScreenAtMapsClickToScreen(t *testing.T) {
	m := newRailModel(140) // Mission default → tab bar present
	y := m.railFirstItemY()

	require.Equal(t, ScreenMission, m.railScreenAt(2, y))
	require.Equal(t, ScreenDashboard, m.railScreenAt(2, y+1))
	require.Equal(t, ScreenRuns, m.railScreenAt(2, y+2))
	require.Equal(t, "", m.railScreenAt(2, y-1), "above first entry")
	require.Equal(t, "", m.railScreenAt(navRailWidth+5, y), "click outside rail x range")

	m.width = 80 // collapsed
	require.Equal(t, "", m.railScreenAt(2, y))
}
