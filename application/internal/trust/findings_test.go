package trust

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerificationCheckCheckConfidence(t *testing.T) {
	c := VerificationCheck{Confidence: 0.42}
	require.Equal(t, 0.42, c.CheckConfidence())
}
