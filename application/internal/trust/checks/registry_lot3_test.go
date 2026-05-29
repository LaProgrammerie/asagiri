package checks

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAllRegisteredCheckTypes(t *testing.T) {
	types := AllRegisteredCheckTypes()
	require.Len(t, types, 14)
	require.Equal(t, typeStaticAnalysis, types[0])
	require.Equal(t, typeBlastRadius, types[12])
	require.Equal(t, typeTests, types[13])
}

func TestNewDefaultRegistryFourteenRunnersOrdered(t *testing.T) {
	reg := NewDefaultRegistry(DefaultDependencies())
	require.Len(t, reg.Runners(), 14)
	got := make([]string, len(reg.Runners()))
	for i, r := range reg.Runners() {
		got[i] = r.Type()
	}
	require.Equal(t, AllRegisteredCheckTypes(), got)
}
