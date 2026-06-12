package trustcli_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/trustcli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootCommandWiring(t *testing.T) {
	opts := trustcli.Options{
		DryRun:          boolPtr(false),
		LoadWorkContext: nil,
		RunRootUI:       func(cmd *cobra.Command, args []string) error { return nil },
	}
	cmd := trustcli.RootCommand(opts)
	require.Equal(t, "trust", cmd.Use)

	want := []string{
		"gates", "replay", "diff", "task", "feature", "run",
	}
	names := subcommandNames(cmd)
	for _, name := range want {
		require.Contains(t, names, name, "missing subcommand %q", name)
	}

	diff := findSub(cmd, "diff")
	require.NotNil(t, diff)
	for _, sub := range []string{"task", "feature", "run"} {
		require.Contains(t, subcommandNames(diff), sub)
	}
}

func boolPtr(v bool) *bool { return &v }

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
