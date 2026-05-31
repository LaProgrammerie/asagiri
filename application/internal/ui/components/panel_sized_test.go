package components

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/charmbracelet/lipgloss"
	"github.com/stretchr/testify/require"
)

func TestPanelSizedFootprint(t *testing.T) {
	th := theme.Default()
	got := PanelSized("Runtime", "agents: 3\nsessions: 1", 40, 8, th)

	require.Equal(t, 40, lipgloss.Width(got), "rendered width must match requested w")
	require.Equal(t, 8, lipgloss.Height(got), "rendered height must match requested h")
	require.Contains(t, stripANSI(got), "Runtime")
	require.Contains(t, stripANSI(got), "agents: 3")
}

func TestPanelSizedClampsBody(t *testing.T) {
	th := theme.Default()
	var body strings.Builder
	for i := 0; i < 40; i++ {
		body.WriteString("line\n")
	}
	got := PanelSized("Events", body.String(), 30, 6, th)

	require.Equal(t, 6, lipgloss.Height(got), "tall body must be clamped to h rows")
	require.Equal(t, 30, lipgloss.Width(got))
}

func TestPanelSizedTruncatesTitle(t *testing.T) {
	th := theme.Default()
	got := PanelSized("A very long panel title that should be truncated", "", 16, 4, th)

	require.Equal(t, 16, lipgloss.Width(got))
	require.Contains(t, stripANSI(got), "…")
}

func TestRenderNavRailActiveAndBadge(t *testing.T) {
	th := theme.Default()
	items := []NavItem{
		{Icon: "◆", Label: "Mission", Active: true},
		{Icon: "▶", Label: "Runs", Badge: "3"},
	}
	got := RenderNavRail("NAVIGATION", items, 24, 10, th)
	plain := stripANSI(got)

	require.Equal(t, 10, lipgloss.Height(got))
	require.Contains(t, plain, "Mission")
	require.Contains(t, plain, "Runs")
	require.Contains(t, plain, "3")
	require.Contains(t, plain, "NAVIGATION")
}

func TestRenderTopBottomBars(t *testing.T) {
	th := theme.Default()
	top := RenderTopBar("ASAGIRI", "12:00:00", 60, th)
	bottom := RenderBottomBar("TIP", "online", 60, th)

	require.Equal(t, 60, lipgloss.Width(top))
	require.Equal(t, 60, lipgloss.Width(bottom))
	require.Contains(t, stripANSI(top), "ASAGIRI")
	require.Contains(t, stripANSI(top), "12:00:00")
	require.Contains(t, stripANSI(bottom), "TIP")
	require.Contains(t, stripANSI(bottom), "online")
}
