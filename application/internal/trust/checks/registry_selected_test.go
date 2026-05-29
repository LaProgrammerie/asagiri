package checks

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistryRunSelected(t *testing.T) {
	reg := NewRegistry()
	reg.runners = []Runner{StaticAnalysisRunner{}, ContractsRunner{}}
	out, err := reg.RunSelected(context.Background(), Scope{TrustID: "t1", Flow: "f"}, []string{"contracts"})
	require.NoError(t, err)
	require.Len(t, out, 1)
	require.Equal(t, "contracts", out[0].Type)
}
