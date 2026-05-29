package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestComputeSingleBounds(t *testing.T) {
	t.Parallel()

	e := NewEngine(100)
	got := e.Compute(Single, 120, 40)

	require.Equal(t, Single, got.Kind)
	require.Len(t, got.Panes, 1)
	require.Equal(t, PaneBounds{ID: PaneMain, X: 0, Y: 0, Width: 120, Height: 40}, got.Panes[0])

	main, ok := got.Primary()
	require.True(t, ok)
	require.Equal(t, got.Panes[0], main)

	_, ok = got.Secondary()
	require.False(t, ok)
}

func TestComputeSplitHorizontalBounds(t *testing.T) {
	t.Parallel()

	e := NewEngine(100)
	got := e.Compute(SplitHorizontal, 120, 40)

	require.Equal(t, SplitHorizontal, got.Kind)
	require.Len(t, got.Panes, 2)

	main := got.Panes[0]
	side := got.Panes[1]
	require.Equal(t, PaneMain, main.ID)
	require.Equal(t, PaneSide, side.ID)
	require.Equal(t, 0, main.X)
	require.Equal(t, 0, main.Y)
	require.Equal(t, 120, main.Width)
	require.Equal(t, 20, main.Height)
	require.Equal(t, 0, side.X)
	require.Equal(t, 20, side.Y)
	require.Equal(t, 120, side.Width)
	require.Equal(t, 20, side.Height)
	require.Equal(t, 40, main.Height+side.Height)
}

func TestComputeSplitVerticalBounds(t *testing.T) {
	t.Parallel()

	e := NewEngine(100)
	got := e.Compute(SplitVertical, 140, 40)

	require.Equal(t, SplitVertical, got.Kind)
	require.Len(t, got.Panes, 2)

	main := got.Panes[0]
	side := got.Panes[1]
	require.Equal(t, PaneMain, main.ID)
	require.Equal(t, PaneSide, side.ID)
	require.Equal(t, 0, main.X)
	require.Equal(t, 0, main.Y)
	require.Equal(t, 70, main.Width)
	require.Equal(t, 40, main.Height)
	require.Equal(t, 70, side.X)
	require.Equal(t, 0, side.Y)
	require.Equal(t, 70, side.Width)
	require.Equal(t, 40, side.Height)
	require.Equal(t, 140, main.Width+side.Width)
}

func TestComputeZeroDimensionsReturnsEmptySingle(t *testing.T) {
	t.Parallel()

	e := NewEngine(100)
	got := e.Compute(SplitVertical, 0, 40)
	require.Equal(t, Single, got.Kind)
	require.Empty(t, got.Panes)
}

func TestComputeCompactForcesSingle(t *testing.T) {
	e := NewEngine(100)
	got := e.Compute(SplitVertical, 80, 24)
	if got.Kind != Single {
		t.Fatalf("kind: got %q want single", got.Kind)
	}
	if len(got.Panes) != 1 {
		t.Fatalf("panes: got %d", len(got.Panes))
	}
}

func TestComputeCompactKeepsHorizontalSplit(t *testing.T) {
	t.Parallel()

	e := NewEngine(100)
	got := e.Compute(SplitHorizontal, 80, 24)

	require.Equal(t, SplitHorizontal, got.Kind)
	require.Len(t, got.Panes, 2)
}
