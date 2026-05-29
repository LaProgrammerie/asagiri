package layout

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPanelSessionSetTab(t *testing.T) {
	t.Parallel()
	s := NewSession()
	s.TabLabels = []string{"Overview", "Agents", "Events"}
	s.SetTab(1)
	require.Equal(t, 1, s.ActiveTab)
	s.SetTab(99)
	require.Equal(t, 2, s.ActiveTab)
	s.NextTab()
	require.Equal(t, 0, s.ActiveTab)
}
