package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/workcli"
	"github.com/spf13/cobra"
)

func newWorkCmd(dryRun *bool) *cobra.Command {
	return workcli.RootCommand(workcli.Options{
		DryRun:          dryRun,
		LoadWorkContext: adaptWorkContext,
	})
}

func adaptWorkContext(startDir string, dryRun bool) (*workcli.WorkContext, error) {
	ctx, err := loadContext(startDir, dryRun)
	if err != nil {
		return nil, err
	}
	return workcli.NewWorkContext(workcli.ContextDeps{
		RepoRoot: ctx.RepoRoot,
		Config:   ctx.Config,
		Store:    ctx.Store,
		DryRun:   ctx.DryRun,
		Snapshot: ctx.snapshot,
		Workflow: ctx.Workflow,
		SyncFn:   ctx.syncPrimitive,
		Close:    ctx.Close,
	}), nil
}
