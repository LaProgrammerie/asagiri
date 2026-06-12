package checks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestConfidenceFromStatus(t *testing.T) {
	require.Equal(t, 1.0, confidenceFromStatus(statusPassed))
	require.Equal(t, 0.75, confidenceFromStatus(statusWarn))
	require.Zero(t, confidenceFromStatus(statusFailed))
	require.Zero(t, confidenceFromStatus(statusSkipped))
}
