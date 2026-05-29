package cli

import (
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/tui"
	uiapp "github.com/LaProgrammerie/asagiri/application/internal/ui/app"
	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
	"github.com/spf13/cobra"
)

func runRootUICommand(dryRun *bool) func(cmd *cobra.Command, args []string) error {
	return runUIScreenCommand(dryRun, "")
}

func runUIScreenCommand(dryRun *bool, screen string) func(cmd *cobra.Command, args []string) error {
	return runUIScreenCommandWithOptions(dryRun, screen, nil)
}

type uiOptionsSetup func(args []string, opts *uiapp.Options)

func runUIScreenCommandWithOptions(dryRun *bool, screen string, setup uiOptionsSetup) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		if setup == nil && len(args) > 0 {
			return cmd.Help()
		}
		if dryRun != nil && *dryRun {
			return cmd.Help()
		}
		if !tui.DetectTTY(cmd.OutOrStdout()) {
			return cmd.Help()
		}

		startDir, err := os.Getwd()
		if err != nil {
			return err
		}
		repoRoot, err := bootstrap.GitRoot(startDir)
		if err != nil {
			return err
		}
		cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
		if err != nil {
			return err
		}
		mode := strings.ToLower(strings.TrimSpace(cfg.UI.Mode))
		if mode == "plain" || mode == "json" {
			return cmd.Help()
		}

		deps := bus.Deps{
			RepoRoot:    repoRoot,
			StateDBPath: cfg.StateDBPath(repoRoot),
			Config:      cfg,
			DryRun:      env.DryRunEnabled(),
		}
		opts := uiapp.Options{
			In:            cmd.InOrStdin(),
			Out:           cmd.OutOrStdout(),
			Err:           cmd.ErrOrStderr(),
			Config:        cfg.UI,
			InitialScreen: cfg.UI.DefaultScreen,
			CommandBus:    bus.NewCommandBus(deps),
			QueryBus:      bus.NewQueryBus(deps),
		}
		if screen != "" {
			opts.InitialScreen = screen
		}
		if setup != nil {
			setup(args, &opts)
		}
		return uiapp.Run(cmd.Context(), opts)
	}
}

func newMissionCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:     "mission",
		Aliases: []string{"mc"},
		Short:   "Ouvrir Mission Control",
		RunE:    runUIScreenCommand(dryRun, uiapp.ScreenMission),
	}
}

func newDashboardCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "dashboard",
		Short: "Ouvrir le dashboard live",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenDashboard),
	}
}

func newAgentsCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Visualiser les agents en temps réel",
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "watch",
		Short: "Ouvrir Agent Theatre",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenAgents),
	})
	return cmd
}

func newExplainCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "explain",
		Short: "Ouvrir le panel Explain",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenExplain),
	}
}

func newFlowCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flow",
		Short: "Ouvrir Flow Explorer",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenFlow),
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "open <name>",
		Short: "Ouvrir un flow dans la TUI",
		Args:  cobra.ExactArgs(1),
		RunE: runUIScreenCommandWithOptions(dryRun, uiapp.ScreenFlow, flowOpenOptionsSetup),
	})
	return cmd
}

func flowOpenOptionsSetup(args []string, opts *uiapp.Options) {
	if len(args) == 0 {
		return
	}
	opts.FlowID = strings.TrimSpace(args[0])
}
