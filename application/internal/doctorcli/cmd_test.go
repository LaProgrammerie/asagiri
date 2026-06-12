package doctorcli_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/doctorcli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootCommandWiring(t *testing.T) {
	cmd := doctorcli.RootCommand()
	require.Equal(t, "doctor", cmd.Use)
	require.True(t, cmd.SilenceUsage)
	require.True(t, cmd.SilenceErrors)

	names := subcommandNames(cmd)
	require.Contains(t, names, "diff")
	require.Contains(t, names, "architecture")

	full, err := cmd.Flags().GetBool("full")
	require.NoError(t, err)
	require.False(t, full)
	jsonOut, err := cmd.Flags().GetBool("json")
	require.NoError(t, err)
	require.False(t, jsonOut)
	strict, err := cmd.Flags().GetBool("strict")
	require.NoError(t, err)
	require.False(t, strict)
	save, err := cmd.Flags().GetBool("save")
	require.NoError(t, err)
	require.False(t, save)
}

func subcommandNames(cmd *cobra.Command) []string {
	subs := cmd.Commands()
	out := make([]string, 0, len(subs))
	for _, c := range subs {
		out = append(out, c.Name())
	}
	return out
}
