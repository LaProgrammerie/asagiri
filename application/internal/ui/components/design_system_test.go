package components

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

func TestProgressBarRatio(t *testing.T) {
	got := ProgressBar(0.5, 10)
	require.Contains(t, stripANSI(got), "█████░░░░░")
}

func stripANSI(v string) string {
	var out []rune
	in := false
	for _, r := range v {
		if r == '\x1b' {
			in = true
			continue
		}
		if in {
			if r == 'm' {
				in = false
			}
			continue
		}
		out = append(out, r)
	}
	return string(out)
}

func TestSparklineEmpty(t *testing.T) {
	require.Equal(t, "········", Sparkline(nil, 8))
}

func TestVirtualListWindow(t *testing.T) {
	items := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	win := SliceWindow(len(items), 5, 4)
	slice := VisibleSlice(items, win)
	require.Len(t, slice, 4)
	require.Equal(t, 4, slice[0])
}

func TestRiskCardHighRiskNodes(t *testing.T) {
	body := RiskCard(
		bus.TrustExplorerResult{ResidualRisk: "medium"},
		bus.GraphExplorerResult{Nodes: []bus.GraphNodeSummary{
			{ID: "n1", Title: "blocked step", Risk: "high", Status: "blocked"},
		}},
	)
	require.Contains(t, body, "Residual: medium")
	require.Contains(t, body, "blocked step")
}

func TestRenderTabsActive(t *testing.T) {
	out := RenderTabs(TabsViewModel{Labels: []string{"A", "B"}, Active: 1, Theme: theme.Default()})
	require.Contains(t, out, "B")
}

func TestRenderToast(t *testing.T) {
	out := RenderToast(ToastSuccess, "saved", theme.Default())
	require.Contains(t, out, "saved")
}

func TestStatusGlyphRunningAnimated(t *testing.T) {
	require.Equal(t, "⠋", StatusGlyph(StateRunning, true))
	require.Equal(t, "◐", StatusGlyph(StateRunning, false))
}

func TestLoadingShimmerRunning(t *testing.T) {
	require.Equal(t, "░ ", LoadingShimmer(true, 0, StateRunning))
	require.Equal(t, "▒ ", LoadingShimmer(true, 1, StateWaiting))
	require.Equal(t, "", LoadingShimmer(true, 0, StateSuccess))
	require.Equal(t, "", LoadingShimmer(false, 0, StateRunning))
}
