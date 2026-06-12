package trustcli

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/reportdiff"
	"github.com/LaProgrammerie/asagiri/application/internal/reportsink"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/spf13/cobra"
)

func newTrustDiffCmd(opts Options) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diff",
		Short: "Comparer deux snapshots trust enregistrés",
		Long: `Compare le snapshot latest avec l'entrée d'historique la plus récente.

Nécessite au moins deux sauvegardes (--save) pour le même scope.`,
	}
	cmd.AddCommand(
		newTrustDiffTaskCmd(opts),
		newTrustDiffFeatureCmd(opts),
		newTrustDiffRunCmd(opts),
	)
	return cmd
}

func newTrustDiffTaskCmd(opts Options) *cobra.Command {
	var jsonOut bool
	var fromPath, toPath string
	dryRun := opts.DryRun
	cmd := &cobra.Command{
		Use:   "task <task-id>",
		Short: "Diff des snapshots trust task",
		Args:  cobra.ExactArgs(1),
		Example: `  asa trust diff task task-42
  asa trust diff task task-42 --json
  asa trust diff task task-42 --from .asagiri/reports/trust/tasks/history/task-42_20260608T120000Z.json`,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := opts.LoadWorkContext(osGetwdMust(), *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			taskID := args[0]
			rel, err := reportsink.TrustTaskRel(taskID)
			if err != nil {
				return err
			}
			beforeRel, afterRel, err := resolveDiffPaths(ctx.RepoRoot, rel, fromPath, toPath)
			if err != nil {
				return err
			}
			var before, after worktrust.WorkTrustReport
			if err := reportsink.ReadJSONFile(ctx.RepoRoot, beforeRel, &before); err != nil {
				return err
			}
			if err := reportsink.ReadJSONFile(ctx.RepoRoot, afterRel, &after); err != nil {
				return err
			}
			diff := reportdiff.DiffTrustTask(before, after, reportdiff.ReportPaths{Before: beforeRel, After: afterRel})
			return writeTrustTaskDiff(cmd, diff, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du diff sur stdout")
	cmd.Flags().StringVar(&fromPath, "from", "", "Snapshot source (défaut: dernière entrée history/)")
	cmd.Flags().StringVar(&toPath, "to", "", "Snapshot cible (défaut: latest)")
	return cmd
}

func newTrustDiffFeatureCmd(opts Options) *cobra.Command {
	var jsonOut bool
	var fromPath, toPath string
	dryRun := opts.DryRun
	cmd := &cobra.Command{
		Use:   "feature <feature>",
		Short: "Diff des snapshots trust feature",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := opts.LoadWorkContext(osGetwdMust(), *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			feature := args[0]
			rel, err := reportsink.TrustFeatureRel(feature)
			if err != nil {
				return err
			}
			beforeRel, afterRel, err := resolveDiffPaths(ctx.RepoRoot, rel, fromPath, toPath)
			if err != nil {
				return err
			}
			var before, after worktrust.FeatureTrustReport
			if err := reportsink.ReadJSONFile(ctx.RepoRoot, beforeRel, &before); err != nil {
				return err
			}
			if err := reportsink.ReadJSONFile(ctx.RepoRoot, afterRel, &after); err != nil {
				return err
			}
			diff := reportdiff.DiffTrustFeature(before, after, reportdiff.ReportPaths{Before: beforeRel, After: afterRel})
			return writeTrustFeatureDiff(cmd, diff, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du diff sur stdout")
	cmd.Flags().StringVar(&fromPath, "from", "", "Snapshot source (défaut: dernière entrée history/)")
	cmd.Flags().StringVar(&toPath, "to", "", "Snapshot cible (défaut: latest)")
	return cmd
}

func newTrustDiffRunCmd(opts Options) *cobra.Command {
	var jsonOut bool
	var fromPath, toPath string
	dryRun := opts.DryRun
	cmd := &cobra.Command{
		Use:   "run <run-id>",
		Short: "Diff des snapshots trust run",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx, err := opts.LoadWorkContext(osGetwdMust(), *dryRun)
			if err != nil {
				return err
			}
			defer ctx.Close()
			runID := args[0]
			rel, err := reportsink.TrustRunRel(runID)
			if err != nil {
				return err
			}
			beforeRel, afterRel, err := resolveDiffPaths(ctx.RepoRoot, rel, fromPath, toPath)
			if err != nil {
				return err
			}
			var before, after worktrust.RunTrustReport
			if err := reportsink.ReadJSONFile(ctx.RepoRoot, beforeRel, &before); err != nil {
				return err
			}
			if err := reportsink.ReadJSONFile(ctx.RepoRoot, afterRel, &after); err != nil {
				return err
			}
			diff := reportdiff.DiffTrustRun(before, after, reportdiff.ReportPaths{Before: beforeRel, After: afterRel})
			return writeTrustRunDiff(cmd, diff, jsonOut)
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Sortie JSON du diff sur stdout")
	cmd.Flags().StringVar(&fromPath, "from", "", "Snapshot source (défaut: dernière entrée history/)")
	cmd.Flags().StringVar(&toPath, "to", "", "Snapshot cible (défaut: latest)")
	return cmd
}

// ResolveDiffPaths picks before/after snapshot paths for trust report diffs.
func ResolveDiffPaths(repoRoot, latestRel, fromPath, toPath string) (beforeRel, afterRel string, err error) {
	return resolveDiffPaths(repoRoot, latestRel, fromPath, toPath)
}

func resolveDiffPaths(repoRoot, latestRel, fromPath, toPath string) (beforeRel, afterRel string, err error) {
	fromPath = strings.TrimSpace(fromPath)
	toPath = strings.TrimSpace(toPath)
	if fromPath != "" && toPath != "" {
		return fromPath, toPath, nil
	}
	if fromPath == "" && toPath == "" {
		return reportsink.DiffPairPaths(repoRoot, latestRel)
	}
	if toPath == "" {
		toRel := relRepoOrPass(repoRoot, latestRel)
		return fromPath, toRel, nil
	}
	if fromPath == "" {
		before, _, err := reportsink.DiffPairPaths(repoRoot, latestRel)
		if err != nil {
			return "", "", err
		}
		return before, toPath, nil
	}
	return fromPath, toPath, nil
}

func relRepoOrPass(repoRoot, rel string) string {
	if strings.HasPrefix(rel, ".asagiri/") {
		return rel
	}
	return rel
}

func writeTrustTaskDiff(cmd *cobra.Command, diff reportdiff.TrustTaskDiff, jsonOut bool) error {
	out := cmd.OutOrStdout()
	if jsonOut {
		return reportdiff.FormatJSON(out, diff)
	}
	reportdiff.FormatTrustTaskText(out, diff)
	return nil
}

func writeTrustFeatureDiff(cmd *cobra.Command, diff reportdiff.TrustFeatureDiff, jsonOut bool) error {
	out := cmd.OutOrStdout()
	if jsonOut {
		return reportdiff.FormatJSON(out, diff)
	}
	reportdiff.FormatTrustFeatureText(out, diff)
	return nil
}

func writeTrustRunDiff(cmd *cobra.Command, diff reportdiff.TrustRunDiff, jsonOut bool) error {
	out := cmd.OutOrStdout()
	if jsonOut {
		return reportdiff.FormatJSON(out, diff)
	}
	reportdiff.FormatTrustRunText(out, diff)
	return nil
}
