package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/trustcli"
	"github.com/spf13/cobra"
)

func newTrustCmd(dryRun *bool) *cobra.Command {
	return trustcli.RootCommand(trustcli.Options{
		DryRun:          dryRun,
		LoadWorkContext: adaptTrustWorkContext,
		RunRootUI:       runUIScreenCommand(dryRun, "trust"),
	})
}

func adaptTrustWorkContext(startDir string, dryRun bool) (*trustcli.WorkContext, error) {
	ctx, err := loadContext(startDir, dryRun)
	if err != nil {
		return nil, err
	}
	return trustcli.NewWorkContext(ctx.RepoRoot, ctx.Config, ctx.Store, ctx.Close), nil
}
