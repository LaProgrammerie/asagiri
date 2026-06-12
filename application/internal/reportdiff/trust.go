package reportdiff

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
)

func DiffTrustTask(before, after worktrust.WorkTrustReport, paths ReportPaths) TrustTaskDiff {
	return TrustTaskDiff{
		ReportVersion: ReportVersion,
		Scope:         "task",
		ScopeID:       after.Scope.TaskID,
		Paths:         paths,
		Score:         scoreDelta(before.Score.Overall, after.Score.Overall),
		Verdict:       verdictDelta(string(before.Score.Verdict), string(after.Score.Verdict)),
		Dimensions:    diffDimensions(before.Dimensions, after.Dimensions),
		NextAction:    nextActionDelta(before.Recommendation.Command, after.Recommendation.Command),
	}
}

func DiffTrustFeature(before, after worktrust.FeatureTrustReport, paths ReportPaths) TrustFeatureDiff {
	return TrustFeatureDiff{
		ReportVersion:  ReportVersion,
		Scope:          "feature",
		ScopeID:        after.Scope.ID,
		Paths:          paths,
		Score:          scoreDelta(before.Score.Overall, after.Score.Overall),
		Verdict:        verdictDelta(string(before.Score.Verdict), string(after.Score.Verdict)),
		NextAction:     nextActionsDelta(before.NextActions, after.NextActions),
		TaskCount:      before.TaskCount,
		TaskCountAfter: after.TaskCount,
	}
}

func DiffTrustRun(before, after worktrust.RunTrustReport, paths ReportPaths) TrustRunDiff {
	return TrustRunDiff{
		ReportVersion: ReportVersion,
		Scope:         "run",
		ScopeID:       after.Scope.ID,
		Paths:         paths,
		Score:         scoreDelta(before.Score.Overall, after.Score.Overall),
		Verdict:       verdictDelta(string(before.Score.Verdict), string(after.Score.Verdict)),
		NextAction:    nextActionsDelta(before.NextActions, after.NextActions),
	}
}

func scoreDelta(before, after float64) ScoreDelta {
	return ScoreDelta{
		Before: before,
		After:  after,
		Delta:  round2(after - before),
	}
}

func verdictDelta(before, after string) VerdictDelta {
	before = strings.TrimSpace(before)
	after = strings.TrimSpace(after)
	return VerdictDelta{
		Before:  before,
		After:   after,
		Changed: before != after,
	}
}

func nextActionDelta(beforeCmd, afterCmd string) NextActionDelta {
	beforeCmd = strings.TrimSpace(beforeCmd)
	afterCmd = strings.TrimSpace(afterCmd)
	return NextActionDelta{
		BeforeCommand: beforeCmd,
		AfterCommand:  afterCmd,
		Changed:       beforeCmd != afterCmd,
	}
}

func nextActionsDelta(before, after []worktrust.WorkTrustRecommendation) NextActionDelta {
	return nextActionDelta(firstCommand(before), firstCommand(after))
}

func firstCommand(recs []worktrust.WorkTrustRecommendation) string {
	if len(recs) == 0 {
		return ""
	}
	return recs[0].Command
}

func diffDimensions(before, after []worktrust.WorkTrustDimension) []DimensionDelta {
	byID := make(map[string]worktrust.WorkTrustDimension, len(before))
	for _, d := range before {
		byID[string(d.ID)] = d
	}
	seen := make(map[string]struct{}, len(after))
	out := make([]DimensionDelta, 0, len(after))
	for _, a := range after {
		id := string(a.ID)
		seen[id] = struct{}{}
		b, ok := byID[id]
		beforeScore := float64(worktrust.UnevaluatedScore)
		label := a.Label
		if ok {
			beforeScore = b.Score
			if label == "" {
				label = b.Label
			}
		}
		delta := round2(a.Score - beforeScore)
		out = append(out, DimensionDelta{
			ID:      id,
			Label:   label,
			Before:  beforeScore,
			After:   a.Score,
			Delta:   delta,
			Changed: beforeScore != a.Score,
		})
	}
	for id, b := range byID {
		if _, ok := seen[id]; ok {
			continue
		}
		out = append(out, DimensionDelta{
			ID:      id,
			Label:   b.Label,
			Before:  b.Score,
			After:   float64(worktrust.UnevaluatedScore),
			Delta:   round2(float64(worktrust.UnevaluatedScore) - b.Score),
			Changed: true,
		})
	}
	return out
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
