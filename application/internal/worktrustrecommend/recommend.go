package worktrustrecommend

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
)

// RecommendationFromIntent aligns trust next-action with asa next / intent planner.
func RecommendationFromIntent(repoRoot string, cfg *config.Config, task sqlite.Task, report worktrust.WorkTrustReport) worktrust.WorkTrustRecommendation {
	rec, err := intent.RecommendForTask(repoRoot, cfg, task)
	if err != nil || strings.TrimSpace(rec.Primitive) == "" {
		return fallbackNextRecommendation(task.Feature, report)
	}
	return workTrustFromIntent(rec, report.Score.Verdict)
}

// FeatureNextFromIntent aligns feature-level next actions with intent planner.
func FeatureNextFromIntent(repoRoot string, cfg *config.Config, feature string, tasks []sqlite.Task, summaries []worktrust.FeatureTaskSummary) []worktrust.WorkTrustRecommendation {
	rec, err := intent.RecommendNextFromTasks(repoRoot, cfg, feature, tasks)
	if err != nil || strings.TrimSpace(rec.Primitive) == "" {
		return fallbackFeatureNextActions(feature, summaries)
	}
	worst := worktrust.VerdictTrusted
	for _, s := range summaries {
		worst = worstVerdict(worst, s.Verdict)
	}
	return []worktrust.WorkTrustRecommendation{workTrustFromIntent(rec, worst)}
}

func workTrustFromIntent(rec intent.NextRecommendation, verdict worktrust.Verdict) worktrust.WorkTrustRecommendation {
	priority := "suggested"
	switch rec.Action {
	case "enrich", "verify", "human_review", "plan":
		priority = "required"
	}
	if verdictRank(verdict) >= verdictRank(worktrust.VerdictRisky) && priority != "required" {
		priority = "required"
	}
	rationale := strings.TrimSpace(rec.Reason)
	if rationale == "" {
		rationale = "prochaine action workflow"
	}
	return worktrust.WorkTrustRecommendation{
		Action:    rec.Action,
		Command:   rec.Primitive,
		Rationale: rationale,
		Priority:  priority,
	}
}

func fallbackNextRecommendation(feature string, report worktrust.WorkTrustReport) worktrust.WorkTrustRecommendation {
	priority := "suggested"
	if report.Score.Verdict == worktrust.VerdictBlocked || report.Score.Verdict == worktrust.VerdictRisky {
		priority = "required"
	}
	return worktrust.WorkTrustRecommendation{
		Action:    "next",
		Command:   worktrust.NextCommandForFeature(feature),
		Rationale: "continuer le workflow",
		Priority:  priority,
	}
}

func fallbackFeatureNextActions(feature string, summaries []worktrust.FeatureTaskSummary) []worktrust.WorkTrustRecommendation {
	priority := "suggested"
	rationale := "continuer le workflow"
	for _, s := range summaries {
		if verdictRank(s.Verdict) >= verdictRank(worktrust.VerdictRisky) {
			priority = "required"
			rationale = strings.TrimSpace(s.TaskID)
			if rationale != "" {
				rationale += " — "
			}
			rationale += worktrust.VerdictLabel(s.Verdict) + " (" + worktrust.HumanStatus(s.Status) + ")"
			break
		}
	}
	return []worktrust.WorkTrustRecommendation{{
		Action:    "next",
		Command:   worktrust.NextCommandForFeature(feature),
		Rationale: rationale,
		Priority:  priority,
	}}
}

func worstVerdict(a, b worktrust.Verdict) worktrust.Verdict {
	if verdictRank(a) >= verdictRank(b) {
		return a
	}
	return b
}

func verdictRank(v worktrust.Verdict) int {
	switch v {
	case worktrust.VerdictBlocked:
		return 4
	case worktrust.VerdictRisky:
		return 3
	case worktrust.VerdictAcceptable:
		return 2
	case worktrust.VerdictTrusted:
		return 1
	default:
		return 0
	}
}
