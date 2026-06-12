package agentscli_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentscli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootCommandWiring(t *testing.T) {
	cmd := agentscli.RootCommand()
	require.Equal(t, "agents", cmd.Use)

	want := []string{
		"list", "show", "run", "runs", "diff", "export", "stats", "sync", "external",
	}
	names := subcommandNames(cmd)
	for _, name := range want {
		require.Contains(t, names, name, "missing subcommand %q", name)
	}

	external := findSub(cmd, "external")
	require.NotNil(t, external)
	require.Contains(t, subcommandNames(external), "sync")
}

func subcommandNames(cmd *cobra.Command) []string {
	out := make([]string, 0, len(cmd.Commands()))
	for _, c := range cmd.Commands() {
		out = append(out, c.Name())
	}
	return out
}

func findSub(cmd *cobra.Command, name string) *cobra.Command {
	for _, c := range cmd.Commands() {
		if c.Name() == name {
			return c
		}
	}
	return nil
}
