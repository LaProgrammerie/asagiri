package worktrust

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestFormatDailyNextBlock(t *testing.T) {
	report := WorkTrustReport{
		Scope: TrustScope{Feature: "onboarding", TaskID: "task-42", Status: asagiri.StatusVerified},
		Score: WorkTrustScore{Verdict: VerdictAcceptable},
		Findings: []WorkTrustFinding{
			{Severity: "medium", Source: "governance", Message: "spec drift"},
		},
		Recommendation: WorkTrustRecommendation{
			Command: "asa review onboarding --task task-42 --agent reviewer",
		},
	}
	out := FormatDailyNextBlock(report)
	require.Contains(t, out, "Trust")
	require.Contains(t, out, "Verdict: Acceptable")
	require.Contains(t, out, "Risque: governance — spec drift")
	require.Contains(t, out, "→ asa review onboarding --task task-42 --agent reviewer")
}

func TestFormatDailyStatusBlock(t *testing.T) {
	report := FeatureTrustReport{
		Scope: TrustScope{Feature: "billing"},
		Score: WorkTrustScore{Verdict: VerdictRisky},
		Tasks: []FeatureTaskSummary{{TaskID: "t1", Verdict: VerdictBlocked, Status: asagiri.StatusPlanned}},
		NextActions: []WorkTrustRecommendation{
			{Command: "asa next --feature billing"},
		},
	}
	out := FormatDailyStatusBlock(report)
	require.Contains(t, out, "Trust")
	require.Contains(t, out, "Feature: billing")
	require.Contains(t, out, "Verdict: À surveiller")
	require.Contains(t, out, "1 task à risque")
	require.Contains(t, out, "→ asa next --feature billing")
}

func TestFormatDailyPostWorkLine(t *testing.T) {
	report := WorkTrustReport{
		Scope: TrustScope{Kind: "task", TaskID: "task-7", ID: "task-7"},
		Score: WorkTrustScore{Verdict: VerdictBlocked},
		Findings: []WorkTrustFinding{
			{Source: "enrich", Message: "gate failed"},
		},
	}
	out := FormatDailyPostWorkLine("onboarding / task-7", report)
	require.Contains(t, out, "Trust: Bloqué")
	require.Contains(t, out, "enrich — gate failed")
	require.Contains(t, out, "asa trust task task-7")
}

func TestFormatDailyNextBlockSnapshot(t *testing.T) {
	report := snapshotTaskReport()
	assertGolden(t, "daily_next.txt", FormatDailyNextBlock(report))
}

func TestFormatDailyStatusBlockSnapshot(t *testing.T) {
	report := snapshotFeatureReport()
	assertGolden(t, "daily_status.txt", FormatDailyStatusBlock(report))
}
