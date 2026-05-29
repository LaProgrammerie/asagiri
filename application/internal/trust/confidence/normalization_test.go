package confidence

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestClamp01(t *testing.T) {
	t.Parallel()
	require.Equal(t, 0.0, Clamp01(-0.5))
	require.Equal(t, 0.5, Clamp01(0.5))
	require.Equal(t, 1.0, Clamp01(2.0))
}
