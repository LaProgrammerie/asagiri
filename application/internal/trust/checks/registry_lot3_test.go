package checks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllRegisteredCheckTypes(t *testing.T) {
	types := AllRegisteredCheckTypes()
	require.Len(t, types, 15)
	require.Equal(t, typeStaticAnalysis, types[0])
	require.Equal(t, typeKnowledgeGraph, types[3])
	require.Equal(t, typeBlastRadius, types[13])
	require.Equal(t, typeTests, types[14])
}

func TestNewDefaultRegistryFourteenRunnersOrdered(t *testing.T) {
	reg := NewDefaultRegistry(DefaultDependencies())
	require.Len(t, reg.Runners(), 15)
	got := make([]string, len(reg.Runners()))
	for i, r := range reg.Runners() {
		got[i] = r.Type()
	}
	require.Equal(t, AllRegisteredCheckTypes(), got)
}
