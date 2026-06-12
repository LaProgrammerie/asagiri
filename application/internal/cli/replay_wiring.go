package cli

import (
	"github.com/LaProgrammerie/asagiri/application/internal/replaycli"
	uiapp "github.com/LaProgrammerie/asagiri/application/internal/ui/app"
	"github.com/spf13/cobra"
)

func newReplayCmd(dryRun *bool) *cobra.Command {
	return replaycli.RootCommand(replaycli.Options{
		DryRun:      dryRun,
		LoadContext: adaptReplayContext,
		OpenReplayScreen: runUIScreenCommandWithOptions(dryRun, uiapp.ScreenReplay, func(args []string, opts *uiapp.Options) {
			opts.ReplayID = args[0]
		}),
	})
}

func adaptReplayContext(startDir string, dryRun bool) (*replaycli.Context, error) {
	ctx, err := loadContext(startDir, dryRun)
	if err != nil {
		return nil, err
	}
	return replaycli.NewContext(ctx.RepoRoot, ctx.Config, ctx.Close), nil
}
