package worktrust

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestBuildFeatureReport_AggregatesWorstVerdict(t *testing.T) {
	cfg := enrichActiveCfg(nil)
	tasks := []sqlite.Task{
		{ID: "t-ok", Feature: "feat", Status: asagiri.StatusEnriched, PayloadJSON: payloadWithGate("enrich", "pass", 0.9)},
		{ID: "t-bad", Feature: "feat", Status: asagiri.StatusPlanned, PayloadJSON: payloadWithGate("enrich", "fail", 0.1)},
	}
	report, err := BuildFeatureReport(t.TempDir(), cfg, "feat", tasks)
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.Equal(t, 2, report.TaskCount)
	require.NotEmpty(t, report.NextActions)
}

func TestBuildFeatureReport_EmptyTasks(t *testing.T) {
	_, err := BuildFeatureReport(t.TempDir(), &config.Config{}, "feat", nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "no tasks")
}

func TestFormatFeatureReport_Stable(t *testing.T) {
	report := FeatureTrustReport{
		Scope:     TrustScope{Feature: "feat"},
		TaskCount: 2,
		Score:     WorkTrustScore{Overall: 65, Verdict: VerdictRisky, Summary: "risky — score 65/100"},
		Tasks: []FeatureTaskSummary{
			{TaskID: "t1", Verdict: VerdictBlocked, Score: 30, Status: asagiri.StatusImplemented},
		},
		NextActions: []WorkTrustRecommendation{
			{Command: "asa next --feature feat", Rationale: "t1 — Bloqué (implemented)"},
		},
	}
	out := FormatFeatureReport(report, FormatOptions{})
	require.Contains(t, out, "Summary")
	require.Contains(t, out, "Verdict: À surveiller")
	require.Contains(t, out, "Gates")
	require.Contains(t, out, "t1  Bloqué")
	require.Contains(t, out, "Next actions")
	require.Contains(t, out, "asa next --feature feat")
}
