package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/graphcli"
	"github.com/spf13/cobra"
)

func newGraphCmd(dryRun *bool) *cobra.Command {
	return graphcli.RootCommand(graphcli.Options{
		DryRun:    dryRun,
		RunRootUI: runUIScreenCommand(dryRun, "graph"),
	})
}

func newPlanGraphCmd() *cobra.Command {
	return graphcli.PlanGraphCommand()
}

func newPlanExplainCmd() *cobra.Command {
	return graphcli.PlanExplainCommand()
}
