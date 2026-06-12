package replaycli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestRootCommand_Subcommands(t *testing.T) {
	cmd := RootCommand(Options{
		DryRun:           nil,
		LoadContext:      nil,
		OpenReplayScreen: func(*cobra.Command, []string) error { return nil },
	})
	names := make([]string, 0, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		names = append(names, sub.Name())
	}
	require.ElementsMatch(t, []string{"open", "create", "run", "compare", "explain", "snapshot"}, names)
}
