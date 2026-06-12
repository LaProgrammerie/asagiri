package trustcli

import (
	"encoding/json"
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/reportsink"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrustrecommend"
	"github.com/spf13/cobra"
)

type trustWorkOutputOpts struct {
	jsonOut bool
	explain bool
	save    bool
}

func newTrustWorkTaskCmd(opts Options) *cobra.Command {
	var outOpts trustWorkOutputOpts
	dryRun := opts.DryRun
	cmd := &cobra.Command{
		Use:   "task <task-id>",
		Short: "Synthèse trust work gates pour une tâche (read-only)",
		Long: `Agrège gates.history, validation et statut task en un score UX.

Distinct de « asa verify trust » (checks produit spec-my-B).`,
		Args: cobra.ExactArgs(1),
		Example: `  asa trust task task-42
  asa trust task task-42 --json
  asa trust task task-42 --explain
  asa trust task task-42 --save`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := opts.LoadWorkContext(osGetwdMust(), *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()

			taskID := args[0]
			task, err := ctx.Store.GetTask(taskID)
			if err != nil {
				return fmt.Errorf("task %q not found in local state (check task id): %w", taskID, err)
			}

			report, err := worktrust.BuildTaskReport(ctx.RepoRoot, ctx.Config, *task)
			if err != nil {
				return err
			}
			report.Recommendation = worktrustrecommend.RecommendationFromIntent(ctx.RepoRoot, ctx.Config, *task, report)
			return writeWorkTrustTaskOutput(cmd, ctx.RepoRoot, report, outOpts)
		},
	}
	cmd.Flags().BoolVar(&outOpts.jsonOut, "json", false, "Sortie JSON du work trust report sur stdout")
	cmd.Flags().BoolVar(&outOpts.explain, "explain", false, "Afficher les dimensions détaillées (scores)")
	cmd.Flags().BoolVar(&outOpts.save, "save", false, "Enregistrer le rapport JSON sous .asagiri/reports/trust/tasks/ (confirmation sur stderr)")
	return cmd
}

func newTrustWorkFeatureCmd(opts Options) *cobra.Command {
	var outOpts trustWorkOutputOpts
	dryRun := opts.DryRun
	cmd := &cobra.Command{
		Use:   "feature <feature>",
		Short: "Synthèse trust work gates pour une feature (read-only)",
		Long: `Agrège les rapports task de la feature : score moyen, tasks risky/blocked.

Distinct de « asa verify trust » (checks produit spec-my-B).`,
		Args: cobra.ExactArgs(1),
		Example: `  asa trust feature onboarding
  asa trust feature onboarding --json
  asa trust feature onboarding --explain
  asa trust feature onboarding --save`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := opts.LoadWorkContext(osGetwdMust(), *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()

			feature := args[0]
			tasks, err := ctx.Store.ListTasksByFeature(feature)
			if err != nil {
				return err
			}
			if len(tasks) == 0 {
				return fmt.Errorf("no tasks for feature %q in local state", feature)
			}

			report, err := worktrust.BuildFeatureReport(ctx.RepoRoot, ctx.Config, feature, tasks)
			if err != nil {
				return err
			}
			report.NextActions = worktrustrecommend.FeatureNextFromIntent(ctx.RepoRoot, ctx.Config, feature, tasks, report.Tasks)
			return writeWorkTrustFeatureOutput(cmd, ctx.RepoRoot, feature, report, outOpts)
		},
	}
	cmd.Flags().BoolVar(&outOpts.jsonOut, "json", false, "Sortie JSON du feature trust report sur stdout")
	cmd.Flags().BoolVar(&outOpts.explain, "explain", false, "Afficher les scores numériques par task")
	cmd.Flags().BoolVar(&outOpts.save, "save", false, "Enregistrer le rapport JSON sous .asagiri/reports/trust/features/ (confirmation sur stderr)")
	return cmd
}

func newTrustWorkRunCmd(opts Options) *cobra.Command {
	var outOpts trustWorkOutputOpts
	dryRun := opts.DryRun
	cmd := &cobra.Command{
		Use:   "run <run-id>",
		Short: "Synthèse trust work gates pour un run (read-only)",
		Long: `Agrège plan gate (run scope) et rapports task : score moyen, tasks blocked/risky.

Distinct de « asa verify trust » (checks produit spec-my-B).`,
		Args: cobra.ExactArgs(1),
		Example: `  asa trust run run-42
  asa trust run run-42 --json
  asa trust run run-42 --explain
  asa trust run run-42 --save`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := opts.LoadWorkContext(osGetwdMust(), *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()

			runID := args[0]
			report, err := worktrust.BuildRunReport(ctx.RepoRoot, ctx.Config, ctx.Store, runID)
			if err != nil {
				return err
			}
			return writeWorkTrustRunOutput(cmd, ctx.RepoRoot, runID, report, outOpts)
		},
	}
	cmd.Flags().BoolVar(&outOpts.jsonOut, "json", false, "Sortie JSON du run trust report sur stdout")
	cmd.Flags().BoolVar(&outOpts.explain, "explain", false, "Afficher les scores numériques par task")
	cmd.Flags().BoolVar(&outOpts.save, "save", false, "Enregistrer le rapport JSON sous .asagiri/reports/trust/runs/ (confirmation sur stderr)")
	return cmd
}

func writeWorkTrustTaskOutput(cmd *cobra.Command, repoRoot string, report worktrust.WorkTrustReport, outOpts trustWorkOutputOpts) error {
	out := cmd.OutOrStdout()
	if outOpts.jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("encode work trust report: %w", err)
		}
	} else {
		_, _ = fmt.Fprint(out, worktrust.FormatTaskReport(report, worktrust.FormatOptions{Explain: outOpts.explain}))
	}
	if outOpts.save {
		rel, err := reportsink.SaveTrustTask(repoRoot, report.Scope.TaskID, report)
		if err != nil {
			return err
		}
		printReportSaved(cmd, rel)
	}
	return nil
}

func writeWorkTrustFeatureOutput(cmd *cobra.Command, repoRoot, feature string, report worktrust.FeatureTrustReport, outOpts trustWorkOutputOpts) error {
	out := cmd.OutOrStdout()
	if outOpts.jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("encode feature trust report: %w", err)
		}
	} else {
		_, _ = fmt.Fprint(out, worktrust.FormatFeatureReport(report, worktrust.FormatOptions{Explain: outOpts.explain}))
	}
	if outOpts.save {
		rel, err := reportsink.SaveTrustFeature(repoRoot, feature, report)
		if err != nil {
			return err
		}
		printReportSaved(cmd, rel)
	}
	return nil
}

func writeWorkTrustRunOutput(cmd *cobra.Command, repoRoot, runID string, report worktrust.RunTrustReport, outOpts trustWorkOutputOpts) error {
	out := cmd.OutOrStdout()
	if outOpts.jsonOut {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		if err := enc.Encode(report); err != nil {
			return fmt.Errorf("encode run trust report: %w", err)
		}
	} else {
		_, _ = fmt.Fprint(out, worktrust.FormatRunReport(report, worktrust.FormatOptions{Explain: outOpts.explain}))
	}
	if outOpts.save {
		rel, err := reportsink.SaveTrustRun(repoRoot, runID, report)
		if err != nil {
			return err
		}
		printReportSaved(cmd, rel)
	}
	return nil
}

func printReportSaved(cmd *cobra.Command, relPath string) {
	PrintReportSaved(cmd, relPath)
}

// PrintReportSaved writes the save confirmation line to stderr.
func PrintReportSaved(cmd *cobra.Command, relPath string) {
	_, _ = fmt.Fprintf(cmd.ErrOrStderr(), "Rapport enregistré : %s\n", relPath)
}
