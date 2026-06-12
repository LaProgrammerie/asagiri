package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeGridLayout(t *testing.T) {
	t.Parallel()
	e := NewEngine(100)
	got := e.ComputeWithOpts(Grid, 120, 40, ComputeOpts{Grid: GridSpec{Columns: 2, Rows: 2}})
	require.Equal(t, Grid, got.Kind)
	require.Len(t, got.Panes, 4)
}

func TestComputeDashboardLayout(t *testing.T) {
	t.Parallel()
	e := NewEngine(100)
	got := e.ComputeWithOpts(Dashboard, 180, 30, ComputeOpts{Dashboard: DashboardSpec{Columns: 3}})
	require.Equal(t, Dashboard, got.Kind)
	require.Len(t, got.Panes, 9)
}

func TestComputeFocusLayout(t *testing.T) {
	t.Parallel()
	e := NewEngine(100)
	got := e.ComputeWithOpts(Focus, 100, 24, ComputeOpts{FocusPane: PaneMain})
	require.Equal(t, Focus, got.Kind)
	require.Len(t, got.Panes, 2)
	require.Equal(t, 75, got.Panes[0].Width)
}

func TestComputeFullscreenLayout(t *testing.T) {
	t.Parallel()
	e := NewEngine(100)
	got := e.Compute(Fullscreen, 80, 24)
	require.Equal(t, Fullscreen, got.Kind)
	require.Len(t, got.Panes, 1)
	require.Equal(t, 80, got.Panes[0].Width)
}

func TestDashboardColumns(t *testing.T) {
	t.Parallel()
	require.Equal(t, 1, DashboardColumns(80, 100))
	require.Equal(t, 2, DashboardColumns(140, 100))
	require.Equal(t, 3, DashboardColumns(200, 100))
}

func TestPanelSessionCollapse(t *testing.T) {
	t.Parallel()
	s := NewSession()
	require.False(t, s.IsCollapsed(PaneMain))
	s.ToggleCollapse(PaneMain)
	require.True(t, s.IsCollapsed(PaneMain))
}

func TestVisiblePanesSkipsCollapsed(t *testing.T) {
	t.Parallel()
	c := Computed{
		Panes: []PaneBounds{
			{ID: PaneMain, Width: 10, Height: 10},
			{ID: PaneSide, Width: 10, Height: 10},
		},
	}
	s := NewSession()
	s.ToggleCollapse(PaneSide)
	visible := c.VisiblePanes(s)
	require.Len(t, visible, 1)
	require.Equal(t, PaneMain, visible[0].ID)
}
