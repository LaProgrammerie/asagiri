package input

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsDoubleClick(t *testing.T) {
	now := time.Now()
	prev := now.Add(-200 * time.Millisecond)
	require.True(t, IsDoubleClick(prev, 10, 5, 10, 5, now))
	require.False(t, IsDoubleClick(prev, 10, 5, 20, 5, now))
	require.False(t, IsDoubleClick(prev.Add(-500*time.Millisecond), 10, 5, 10, 5, now))
}

func TestListRowFromY(t *testing.T) {
	require.Equal(t, -1, ListRowFromY(3, 8))
	require.Equal(t, 0, ListRowFromY(8, 8))
	require.Equal(t, 2, ListRowFromY(10, 8))
}

func TestContextMenuSelection(t *testing.T) {
	menu := OpenContextMenu(1, 2, []MenuItem{{ID: "a"}, {ID: "b"}, {ID: "c"}})
	require.NotNil(t, menu)
	MoveContextMenuSelection(menu, 1)
	require.Equal(t, 1, menu.Selected)
	item := SelectedMenuItem(menu)
	require.NotNil(t, item)
	require.Equal(t, "b", item.ID)
}
