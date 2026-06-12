package worktrust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestFormatTaskReportSnapshot(t *testing.T) {
	report := snapshotTaskReport()
	assertGolden(t, "formatter_task.txt", FormatTaskReport(report, FormatOptions{}))
}

func TestFormatTaskReportExplainSnapshot(t *testing.T) {
	report := snapshotTaskReport()
	assertGolden(t, "formatter_task_explain.txt", FormatTaskReport(report, FormatOptions{Explain: true}))
}

func TestFormatFeatureReportSnapshot(t *testing.T) {
	report := snapshotFeatureReport()
	assertGolden(t, "formatter_feature.txt", FormatFeatureReport(report, FormatOptions{}))
}

func TestFormatRunReportSnapshot(t *testing.T) {
	report := snapshotRunReport()
	assertGolden(t, "formatter_run.txt", FormatRunReport(report, FormatOptions{}))
}

func snapshotTaskReport() WorkTrustReport {
	return WorkTrustReport{
		Scope: TrustScope{Kind: "task", ID: "task-42", Feature: "onboarding", TaskID: "task-42", Status: asagiri.StatusVerified},
		Score: WorkTrustScore{Overall: 78, Verdict: VerdictAcceptable, Summary: verdictSummary(VerdictAcceptable, 78)},
		Dimensions: []WorkTrustDimension{
			{ID: DimSpecificationAlignment, Label: "Alignement spec", Score: 82, Status: DimStatusStrong, Summary: "enrich pass", SourceGates: []string{gates.EnrichGateName}},
			{ID: DimGateConfidence, Label: "Confiance gates", Score: 75, Status: DimStatusModerate, Summary: "enrich pass", SourceGates: []string{gates.EnrichGateName}},
			{ID: DimValidationStrength, Label: "Force validation", Score: 70, Status: DimStatusModerate, Summary: "tests OK", SourceGates: []string{gates.VerifyEvidenceGateName}},
		},
		Evidences: []WorkTrustEvidence{
			{Kind: "gate_history", Ref: gates.EnrichGateName, Summary: "pass"},
		},
		Findings: []WorkTrustFinding{
			{Severity: "medium", Source: "governance", Message: "spec drift section 3"},
		},
		Recommendation: WorkTrustRecommendation{
			Command:   "asa review onboarding --task task-42 --agent reviewer",
			Rationale: "task verified, ready for review",
		},
	}
}

func snapshotFeatureReport() FeatureTrustReport {
	return FeatureTrustReport{
		Scope:     TrustScope{Kind: "feature", ID: "onboarding", Feature: "onboarding"},
		TaskCount: 2,
		Score:     WorkTrustScore{Overall: 58, Verdict: VerdictRisky, Summary: verdictSummary(VerdictRisky, 58)},
		Tasks: []FeatureTaskSummary{
			{TaskID: "task-bad", Verdict: VerdictBlocked, Score: 32, Status: asagiri.StatusPlanned},
			{TaskID: "task-ok", Verdict: VerdictAcceptable, Score: 84, Status: asagiri.StatusVerified},
		},
		NextActions: []WorkTrustRecommendation{
			{Command: "asa next --feature onboarding", Rationale: "task-bad — Bloqué (planned)"},
		},
	}
}

func snapshotRunReport() RunTrustReport {
	return RunTrustReport{
		Scope:     TrustScope{Kind: "run", ID: "run-99", Feature: "billing", Status: "running"},
		TaskCount: 2,
		Score:     WorkTrustScore{Overall: 62, Verdict: VerdictRisky, Summary: verdictSummary(VerdictRisky, 62)},
		PlanGate:  &RunPlanGateSummary{Status: "pass", Confidence: 0.91, Notes: "plan reviewed"},
		Tasks: []FeatureTaskSummary{
			{TaskID: "task-r2", Verdict: VerdictRisky, Score: 48, Status: asagiri.StatusImplemented},
			{TaskID: "task-r1", Verdict: VerdictAcceptable, Score: 76, Status: asagiri.StatusVerified},
		},
		NextActions: []WorkTrustRecommendation{
			{Command: "asa next --feature billing", Rationale: "task-r2 — À surveiller (implemented)"},
		},
	}
}

func assertGolden(t *testing.T, name, got string) {
	t.Helper()
	golden := filepath.Join("testdata", name)
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.MkdirAll(filepath.Dir(golden), 0o755))
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
		return
	}
	want, err := os.ReadFile(golden)
	if os.IsNotExist(err) {
		t.Fatalf("golden %s missing; run with UPDATE_GOLDEN=1", golden)
	}
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}
