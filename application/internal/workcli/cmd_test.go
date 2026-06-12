package workcli_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/workcli"
	"github.com/stretchr/testify/require"
)

func TestRootCommandWiring(t *testing.T) {
	dryRun := false
	cmd := workcli.RootCommand(workcli.Options{
		DryRun: &dryRun,
	})
	require.Equal(t, `work "<instruction>"`, cmd.Use)
	require.Contains(t, cmd.Short, "intention")

	flagNames := []string{
		"agent", "reviewer", "plan-only", "yes", "estimate-only",
		"strict-trust", "trust-flow", "investigate-first",
	}
	for _, name := range flagNames {
		require.NotNil(t, cmd.Flags().Lookup(name), "missing flag %q", name)
	}
}
