package onboardingcli

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestOnboardCommand_Flags(t *testing.T) {
	cmd := OnboardCommand(Options{
		RunOnboardUI: func(*cobra.Command, []string) error { return nil },
	})
	require.NotEmpty(t, cmd.Flags().Lookup("yes"))
	require.NotEmpty(t, cmd.Flags().Lookup("json"))
	require.NotEmpty(t, cmd.Flags().Lookup("ui"))
}

func TestReadyCommand_Flags(t *testing.T) {
	cmd := ReadyCommand()
	require.NotEmpty(t, cmd.Flags().Lookup("json"))
	require.NotEmpty(t, cmd.Flags().Lookup("autofix"))
}
