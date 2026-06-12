package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryRunAllEmpty(t *testing.T) {
	var r Registry
	out, err := r.RunAll(context.Background(), Scope{TrustID: "trust-1", Flow: "f"})
	require.NoError(t, err)
	require.Empty(t, out)
	require.NotNil(t, out)
}
