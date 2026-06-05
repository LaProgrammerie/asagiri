package cli

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/contextopt"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/spf13/cobra"
)

func newContextCmd(dryRun *bool) *cobra.Command {
	var (
		taskID    string
		show      bool
		optimize  bool
		fromGraph bool
		flow      string
		action    string
	)
	renderGraphPack := func(cmd *cobra.Command, c *appContext, feature string) error {
		graphPack, err := investigation.ResolveScopeFromGraph(cmd.Context(), c.RepoRoot, investigation.GraphScopeOptions{
			UseKnowledgeGraph: true,
			Flow:              flow,
			Action:            action,
		})
		if err != nil {
			return err
		}
		rep := investigation.Report{
			Request: investigation.Request{Symptom: feature, Flow: flow},
			Scope:   investigation.ResolvedScope{Flow: flow, Action: action},
		}
		_, _ = fmt.Fprint(cmd.OutOrStdout(), investigation.FormatContextPackMarkdown(rep, graphPack))
		return nil
	}
	run := func(cmd *cobra.Command, feature string, graphOnly bool) error {
		if !graphOnly && !show && !optimize {
			return fmt.Errorf("spécifier --show ou --optimize")
		}
		if fromGraph && flow == "" {
			return fmt.Errorf("--from-graph requiert --flow")
		}
		c, err := loadContext(mustWd(), *dryRun)
		if err != nil {
			return err
		}
		defer c.Close()

		if graphOnly {
			return renderGraphPack(cmd, c, feature)
		}

		inv, err := investigation.Run(cmd.Context(), c.RepoRoot, feature, taskID, c.Config)
		if err != nil {
			return err
		}
		if fromGraph {
			graphScope, err := investigation.ResolveScopeFromGraph(cmd.Context(), c.RepoRoot, investigation.GraphScopeOptions{
				UseKnowledgeGraph: true,
				Flow:              flow,
				Action:            action,
			})
			if err != nil {
				return err
			}
			merged := investigation.MergeGraphScope(investigation.ContextPack{
				Files: inv.CandidateFiles,
				Tests: inv.RelatedTests,
			}, graphScope)
			inv.CandidateFiles = merged.Files
			inv.RelatedTests = merged.Tests
		}

		entries, err := contextopt.Collect(c.RepoRoot, feature, c.Config, contextopt.CollectOpts{})
		if err != nil {
			return err
		}
		contextopt.ScoreByKeywords(entries, taskID+" "+feature, feature)
		reduced, _ := contextopt.Reduce(entries, c.Config, contextopt.ReduceOpts{})
		pack := contextopt.BuildPack(c.Config, contextopt.PackInput{
			Feature:      feature,
			TaskID:       taskID,
			Inv:          inv,
			ReducedFiles: reduced,
			OutputFormat: "markdown",
		})
		opt := contextopt.ComputeOptimize(entries, reduced, pack, c.Config.TokenEst)
		if show {
			_, _ = fmt.Fprint(cmd.OutOrStdout(), contextopt.RenderPackMarkdown(pack))
			if fromGraph {
				if err := renderGraphPack(cmd, c, feature); err != nil {
					return err
				}
			}
		}
		if optimize {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Original context: %d tokens\nOptimized: %d tokens\nSavings: %.1f%%\n",
				opt.OriginalTokens, opt.OptimizedTokens, opt.SavingsRatio*100)
		}
		return nil
	}

	cmd := &cobra.Command{
		Use:   "context <feature>",
		Short: "Afficher ou optimiser le contexte prévu",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd, args[0], false)
		},
	}
	buildCmd := &cobra.Command{
		Use:   "build",
		Short: "Construire un context pack depuis le graphe de connaissance (spec-my-E §15)",
		RunE: func(cmd *cobra.Command, args []string) error {
			fromGraph = true
			feature := flow
			if feature == "" {
				feature = "knowledge-graph"
			}
			return run(cmd, feature, true)
		},
	}
	buildCmd.Flags().StringVar(&flow, "flow", "", "ID de flow produit")
	buildCmd.Flags().StringVar(&action, "action", "", "Action dans le flow")
	buildCmd.Flags().BoolVar(&fromGraph, "from-graph", true, "Utiliser le graphe de connaissance")

	cmd.Flags().StringVar(&taskID, "task", "", "ID de tâche")
	cmd.Flags().BoolVar(&show, "show", false, "Afficher le contexte pack")
	cmd.Flags().BoolVar(&optimize, "optimize", false, "Afficher le résumé tokens")
	cmd.Flags().BoolVar(&fromGraph, "from-graph", false, "Enrichir le contexte depuis le graphe de connaissance")
	cmd.Flags().StringVar(&flow, "flow", "", "ID de flow produit")
	cmd.Flags().StringVar(&action, "action", "", "Action dans le flow")
	cmd.AddCommand(buildCmd)
	return cmd
}
