package worktrust

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

// BuildFeatureReport synthesizes trust reports for all tasks in a feature.
func BuildFeatureReport(repoRoot string, cfg *config.Config, feature string, tasks []sqlite.Task) (FeatureTrustReport, error) {
	feature = strings.TrimSpace(feature)
	if feature == "" {
		return FeatureTrustReport{}, fmt.Errorf("feature name required")
	}
	if len(tasks) == 0 {
		return FeatureTrustReport{}, fmt.Errorf("no tasks for feature %q", feature)
	}

	summaries := make([]FeatureTaskSummary, 0, len(tasks))
	var scoreSum float64
	worst := VerdictTrusted

	for _, task := range tasks {
		report, err := BuildTaskReport(repoRoot, cfg, task)
		if err != nil {
			return FeatureTrustReport{}, fmt.Errorf("task %s: %w", task.ID, err)
		}
		summaries = append(summaries, FeatureTaskSummary{
			TaskID:    task.ID,
			Status:    task.Status,
			Score:     report.Score.Overall,
			Verdict:   report.Score.Verdict,
			Command:   report.Recommendation.Command,
			Rationale: report.Recommendation.Rationale,
		})
		scoreSum += report.Score.Overall
		worst = worstVerdict(worst, report.Score.Verdict)
	}

	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].Verdict != summaries[j].Verdict {
			return verdictRank(summaries[i].Verdict) > verdictRank(summaries[j].Verdict)
		}
		return summaries[i].Score < summaries[j].Score
	})

	avg := scoreSum / float64(len(summaries))

	return FeatureTrustReport{
		ReportVersion: ReportVersion,
		Scope: TrustScope{
			Kind:    "feature",
			ID:      feature,
			Feature: feature,
		},
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Score: WorkTrustScore{
			Overall: avg,
			Verdict: worst,
			Summary: verdictSummary(worst, avg),
		},
		TaskCount:   len(summaries),
		Tasks:       summaries,
		NextActions: collectFeatureNextActions(feature, summaries),
	}, nil
}

func collectFeatureNextActions(feature string, summaries []FeatureTaskSummary) []WorkTrustRecommendation {
	return fallbackFeatureNextActions(feature, summaries)
}

func fallbackFeatureNextActions(feature string, summaries []FeatureTaskSummary) []WorkTrustRecommendation {
	priority := "suggested"
	rationale := "continuer le workflow"
	for _, s := range summaries {
		if verdictRank(s.Verdict) >= verdictRank(VerdictRisky) {
			priority = "required"
			rationale = strings.TrimSpace(s.TaskID)
			if rationale != "" {
				rationale += " — "
			}
			rationale += VerdictLabel(s.Verdict) + " (" + HumanStatus(s.Status) + ")"
			break
		}
	}
	return []WorkTrustRecommendation{{
		Action:    "next",
		Command:   NextCommandForFeature(feature),
		Rationale: rationale,
		Priority:  priority,
	}}
}

func worstVerdict(a, b Verdict) Verdict {
	if verdictRank(a) >= verdictRank(b) {
		return a
	}
	return b
}

func verdictRank(v Verdict) int {
	switch v {
	case VerdictBlocked:
		return 4
	case VerdictRisky:
		return 3
	case VerdictAcceptable:
		return 2
	case VerdictTrusted:
		return 1
	default:
		return 0
	}
}
