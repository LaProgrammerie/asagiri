package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/onboardingcli"
	"github.com/spf13/cobra"
)

func newOnboardCmd() *cobra.Command {
	return onboardingcli.OnboardCommand(onboardingcli.Options{
		RunOnboardUI: runUIScreenCommand(nil, "onboarding"),
	})
}

func newReadyCmd() *cobra.Command {
	return onboardingcli.ReadyCommand()
}
