package replaycli

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/replay"
)

// Options wires replay CLI commands from the root `asa` package.
type Options struct {
	DryRun           *bool
	LoadContext      func(startDir string, dryRun bool) (*Context, error)
	OpenReplayScreen func(cmd *cobra.Command, args []string) error
}

// RootCommand returns the `asa replay` command tree.
func RootCommand(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "replay",
		Short: "Capturer, rejouer et comparer des workflows d'ingénierie",
	}
	cmd.AddCommand(
		newOpenCmd(opts),
		newCreateCmd(opts),
		newRunCmd(opts),
		newCompareCmd(opts),
		newExplainCmd(opts),
		newSnapshotCmd(opts),
	)
	return cmd
}

func newOpenCmd(opts Options) *cobra.Command {
	return &cobra.Command{
		Use:   "open <replay-id>",
		Short: "Ouvrir Replay Explorer",
		Args:  cobra.ExactArgs(1),
		RunE:  opts.OpenReplayScreen,
	}
}

func newCreateCmd(opts Options) *cobra.Command {
	var (
		fromRun           string
		fromGraph         string
		fromInvestigation string
		includeRuntime    bool
		includePrompts    bool
		includeEvents     bool
		jsonOut           bool
	)
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Créer un replay package",
		RunE: func(cmd *cobra.Command, _ []string) error {
			c, err := opts.LoadContext(osGetwdMust(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			mgr := replay.DefaultManager(c.RepoRoot, replay.DefaultCapturePolicies(c.Config))
			pkg, err := mgr.Create(cmd.Context(), replay.ReplayCreateRequest{
				RepoRoot:          c.RepoRoot,
				Config:            replay.DefaultCapturePolicies(c.Config),
				FromRun:           fromRun,
				FromGraph:         fromGraph,
				FromInvestigation: fromInvestigation,
				IncludeRuntime:    includeRuntime,
				IncludePrompts:    includePrompts,
				IncludeEvents:     includeEvents,
			})
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(pkg)
			}
			replay.WriteReplayCreate(cmd.OutOrStdout(), pkg)
			return nil
		},
	}
	cmd.Flags().StringVar(&fromRun, "from-run", "", "Run source id")
	cmd.Flags().StringVar(&fromGraph, "from-graph", "", "Execution graph source id")
	cmd.Flags().StringVar(&fromInvestigation, "from-investigation", "", "Investigation source id")
	cmd.Flags().BoolVar(&includeRuntime, "include-runtime", false, "Capturer les événements runtime")
	cmd.Flags().BoolVar(&includePrompts, "include-prompts", false, "Capturer les prompts agents")
	cmd.Flags().BoolVar(&includeEvents, "include-events", false, "Capturer events.jsonl")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newRunCmd(opts Options) *cobra.Command {
	var (
		dryRun     bool
		compare    bool
		strict     bool
		offline    bool
		simulation bool
		jsonOut    bool
	)
	cmd := &cobra.Command{
		Use:   "run <replay-id>",
		Short: "Rejouer un workflow capturé",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opts.LoadContext(osGetwdMust(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			policies := replay.DefaultCapturePolicies(c.Config)
			if offline || policies.OfflineModeDefault {
				offline = true
			}

			mgr := replay.DefaultManager(c.RepoRoot, policies)
			result, err := mgr.Run(cmd.Context(), replay.ReplayRunRequest{
				RepoRoot:   c.RepoRoot,
				ReplayID:   args[0],
				DryRun:     dryRun,
				Compare:    compare,
				Strict:     strict,
				Offline:    offline,
				Simulation: simulation,
			})
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			replay.WriteReplayRun(cmd.OutOrStdout(), result)
			return nil
		},
	}
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Valider le package sans réexécution")
	cmd.Flags().BoolVar(&compare, "compare", false, "Préparer une comparaison avec la session précédente")
	cmd.Flags().BoolVar(&strict, "strict", false, "Échouer si divergence détectée")
	cmd.Flags().BoolVar(&offline, "offline", false, "Mode offline (pas d'appels cloud)")
	cmd.Flags().BoolVar(&simulation, "simulation", false, "Rejouer graph/events sans agents")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newCompareCmd(opts Options) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "compare <replay-a> <replay-b>",
		Short: "Comparer deux replay packages",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opts.LoadContext(osGetwdMust(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			cmp, err := replay.NewComparator(c.RepoRoot).Compare(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(cmp)
			}
			replay.WriteReplayComparison(cmd.OutOrStdout(), cmp)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newExplainCmd(opts Options) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "explain <replay-a> <replay-b>",
		Short: "Expliquer les divergences entre deux replays",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := opts.LoadContext(osGetwdMust(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			cmp, err := replay.NewComparator(c.RepoRoot).Compare(cmd.Context(), args[0], args[1])
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(cmp.Divergences)
			}
			replay.WriteReplayExplain(cmd.OutOrStdout(), cmp)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}

func newSnapshotCmd(opts Options) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "snapshot <replay-id>",
		Short: "Créer un snapshot d'un replay package",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name, err := cmd.Flags().GetString("name")
			if err != nil {
				return err
			}
			if name == "" {
				return fmt.Errorf("--name required")
			}
			c, err := opts.LoadContext(osGetwdMust(), false)
			if err != nil {
				return err
			}
			defer c.Close()

			result, err := replay.DefaultSnapshotter().Snapshot(cmd.Context(), replay.SnapshotRequest{
				RepoRoot: c.RepoRoot,
				ReplayID: args[0],
				Name:     name,
			})
			if err != nil {
				return err
			}
			if jsonOut {
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(result)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "snapshot: %s\npath: %s\n", result.Name, result.Path)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Nom du snapshot")
	_ = cmd.MarkFlagRequired("name")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON")
	return cmd
}
