package embedder_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/embedutil"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/stretchr/testify/require"
)

func TestHashNonRegression(t *testing.T) {
	t.Parallel()
	h := embedder.NewHash()
	text := "onboarding invite member failure"
	got, err := h.Embed(context.Background(), text)
	require.NoError(t, err)
	want := embedutil.Vector(text)
	require.Equal(t, want, got)
	require.Equal(t, embedutil.Dims, h.Dimensions())
	require.Equal(t, "hash", h.Name())
}
