package reportdiff

import (
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/stretchr/testify/require"
)

func TestDiffTrustTask_ScoreVerdictDimensionsNext(t *testing.T) {
	before := worktrust.WorkTrustReport{
		Scope: worktrust.TrustScope{TaskID: "t1"},
		Score: worktrust.WorkTrustScore{Overall: 70, Verdict: worktrust.VerdictAcceptable},
		Dimensions: []worktrust.WorkTrustDimension{
			{ID: worktrust.DimGateConfidence, Label: "Gates", Score: 70},
		},
		Recommendation: worktrust.WorkTrustRecommendation{Command: "asa dev feat --task t1 --agent dev"},
	}
	after := worktrust.WorkTrustReport{
		Scope: worktrust.TrustScope{TaskID: "t1"},
		Score: worktrust.WorkTrustScore{Overall: 85, Verdict: worktrust.VerdictTrusted},
		Dimensions: []worktrust.WorkTrustDimension{
			{ID: worktrust.DimGateConfidence, Label: "Gates", Score: 90},
		},
		Recommendation: worktrust.WorkTrustRecommendation{Command: "asa verify feat --task t1"},
	}
	diff := DiffTrustTask(before, after, ReportPaths{Before: "before.json", After: "after.json"})
	require.Equal(t, ReportVersion, diff.ReportVersion)
	require.Equal(t, 15.0, diff.Score.Delta)
	require.True(t, diff.Verdict.Changed)
	require.True(t, diff.NextAction.Changed)
	require.Len(t, diff.Dimensions, 1)
	require.True(t, diff.Dimensions[0].Changed)
}

func TestFormatTrustTaskText(t *testing.T) {
	diff := TrustTaskDiff{
		ScopeID: "t1",
		Paths:   ReportPaths{Before: "a.json", After: "b.json"},
		Score:   ScoreDelta{Before: 70, After: 85, Delta: 15},
		Verdict: VerdictDelta{Before: "acceptable", After: "trusted", Changed: true},
		NextAction: NextActionDelta{
			BeforeCommand: "asa dev",
			AfterCommand:  "asa verify",
			Changed:       true,
		},
	}
	var b strings.Builder
	FormatTrustTaskText(&b, diff)
	out := b.String()
	require.Contains(t, out, "Score:")
	require.Contains(t, out, "acceptable → trusted")
	require.Contains(t, out, "asa dev → asa verify")
}
