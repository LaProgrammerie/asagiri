package cloudcli_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cloudcli"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootCommandWiring(t *testing.T) {
	cmd := cloudcli.RootCommand()
	require.Equal(t, "cloud", cmd.Use)
	want := []string{"status", "login", "logout", "link", "push"}
	names := subcommandNames(cmd)
	for _, name := range want {
		require.Contains(t, names, name, "missing subcommand %q", name)
	}
}

func subcommandNames(cmd *cobra.Command) []string {
	out := make([]string, 0, len(cmd.Commands()))
	for _, c := range cmd.Commands() {
		out = append(out, c.Name())
	}
	return out
}
