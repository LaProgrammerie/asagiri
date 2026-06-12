package memory_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/stretchr/testify/require"
)

func TestEmbedAndSimilarity(t *testing.T) {
	t.Parallel()
	a := memory.Embed("onboarding invite member failure")
	b := memory.Embed("onboarding invitation async")
	s := memory.CosineSimilarity(a, b)
	require.Greater(t, s, 0.0)
}
