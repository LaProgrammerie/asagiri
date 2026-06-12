package theme_test

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/theme"
	"github.com/stretchr/testify/require"
)

func TestRenderRingGauge(t *testing.T) {
	st := theme.Default().Styles()
	got := st.RenderRingGauge(42)
	require.Contains(t, got, "42%")
	require.Equal(t, 7, len(strings.Split(got, "\n")))
}
