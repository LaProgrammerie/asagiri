package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentscli"
	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/routing"
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

func newRunsCmd(dryRun *bool) *cobra.Command {
	return &cobra.Command{
		Use:   "runs",
		Short: "Ouvrir l'écran Runs (liste + détail)",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenRuns),
	}
}

func newAgentsCmd(dryRun *bool) *cobra.Command {
	cmd := agentscli.RootCommand()
	cmd.AddCommand(&cobra.Command{
		Use:   "watch",
		Short: "Ouvrir Agent Theatre (TUI)",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenAgents),
	})
	return cmd
}

func newExplainCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain",
		Short: "Ouvrir le panel Explain",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenExplain),
	}
	cmd.AddCommand(newExplainRoutingCmd())
	return cmd
}

// RoutingExplanation is the presentation DTO for `asa explain routing`. It is a
// pure presentation projection of a routing.Decision (no business logic lives in
// the CLI; the decision stays in the routing package — ADR-027). It is rendered
// in plain/json parity: every information field is present in both modes
// (Requirement 4.8).
type RoutingExplanation struct {
	StepClass string `json:"step_class"`
	Agent     string `json:"agent"`
	Model     string `json:"model"`
	Local     bool   `json:"local"`
	Reason    string `json:"reason"`
}

// newExplainRoutingCmd adds the non-interactive routing explanation mode:
//
//	asa explain routing --step-class <cls> [--prefer-local --no-cloud --allow-cloud --json]
//
// It calls routing.Route for the given step class and names the selected
// Agent_Backend and the reason, without requiring the user to know the
// underlying backends (Requirement 4.8). The TUI rendering remains unchanged
// (ADR-027).
func newExplainRoutingCmd() *cobra.Command {
	var (
		stepClass   string
		preferLocal bool
		noCloud     bool
		allowCloud  bool
		jsonOut     bool
	)
	cmd := &cobra.Command{
		Use:   "routing",
		Short: "Expliquer le routage d'agent pour une classe d'étape (mode non interactif)",
		Example: "  asa explain routing --step-class implement\n" +
			"  asa explain routing --step-class review --prefer-local\n" +
			"  asa explain routing --step-class implement --no-cloud --json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, _ []string) error {
			repoRoot, err := bootstrap.GitRoot(mustWd())
			if err != nil {
				return err
			}
			cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
			if err != nil {
				return err
			}

			decision, err := routing.Route(cfg, stepClass, preferLocal, noCloud, allowCloud)
			if err != nil {
				return err
			}

			explanation := RoutingExplanation{
				StepClass: decision.StepClass,
				Agent:     decision.Agent,
				Model:     decision.Model,
				Local:     decision.Local,
				Reason:    decision.Reason,
			}

			out := cmd.OutOrStdout()
			if jsonOut {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(explanation)
			}
			_, _ = fmt.Fprint(out, formatRoutingExplanation(explanation))
			return nil
		},
	}
	cmd.Flags().StringVar(&stepClass, "step-class", "", "Classe d'étape à expliquer (ex. implement, review)")
	cmd.Flags().BoolVar(&preferLocal, "prefer-local", false, "Préférer un backend local")
	cmd.Flags().BoolVar(&noCloud, "no-cloud", false, "Interdire le cloud (prévaut sur les autres flags)")
	cmd.Flags().BoolVar(&allowCloud, "allow-cloud", false, "Autoriser le cloud")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

// formatRoutingExplanation renders the plain-text view of a RoutingExplanation.
// It names the backend and the reason, in parity with the JSON output: every
// field present in RoutingExplanation is emitted here too (Requirement 4.8).
func formatRoutingExplanation(e RoutingExplanation) string {
	var b strings.Builder
	_, _ = fmt.Fprintf(&b, "step_class: %s\n", e.StepClass)
	_, _ = fmt.Fprintf(&b, "agent: %s\n", e.Agent)
	_, _ = fmt.Fprintf(&b, "model: %s\n", e.Model)
	_, _ = fmt.Fprintf(&b, "local: %t\n", e.Local)
	_, _ = fmt.Fprintf(&b, "reason: %s\n", e.Reason)
	return b.String()
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
		RunE:  runUIScreenCommandWithOptions(dryRun, uiapp.ScreenFlow, flowOpenOptionsSetup),
	})
	return cmd
}

func flowOpenOptionsSetup(args []string, opts *uiapp.Options) {
	if len(args) == 0 {
		return
	}
	opts.FlowID = strings.TrimSpace(args[0])
}
