package trust

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

func TestWriteReport(t *testing.T) {
	repo := t.TempDir()
	report := skeletonReport(repo)

	mdPath, jsonPath, err := WriteReport(repo, report.TrustID, report)
	require.NoError(t, err)
	require.FileExists(t, mdPath)
	require.FileExists(t, jsonPath)

	body, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	var decoded TrustReport
	require.NoError(t, json.Unmarshal(body, &decoded))
	require.Equal(t, report.TrustID, decoded.TrustID)
	require.Equal(t, report.Flow, decoded.Flow)
	require.Contains(t, string(body), `"residual_risk": "unknown"`)

	mdBody, err := os.ReadFile(mdPath)
	require.NoError(t, err)
	require.Contains(t, string(mdBody), "# Trust Report")
	require.Contains(t, string(mdBody), "onboarding-flow")
	require.Contains(t, string(mdBody), "## Uncovered zones")
	require.Contains(t, string(mdBody), ConfidenceUnavailableLabel)
}

func TestTrustReportGoldenJSON(t *testing.T) {
	report := skeletonReport("/tmp/repo")
	got, err := json.MarshalIndent(report, "", "  ")
	require.NoError(t, err)

	golden := filepath.Join("testdata", "report_skeleton.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(golden, got, 0o644))
	}
	want, err := os.ReadFile(golden)
	require.NoError(t, err)
	require.JSONEq(t, string(want), string(got))
}

func TestFormatTerminalSummary(t *testing.T) {
	summary := FormatTerminalSummary(skeletonReport(t.TempDir()))
	require.Contains(t, summary, "Asagiri Trust Engine")
	require.Contains(t, summary, "onboarding-flow")
	require.Contains(t, summary, "(no checks executed)")
	require.Contains(t, summary, ConfidenceUnavailableLabel)
	require.NotContains(t, summary, "Architecture:    0.00")
	require.Contains(t, summary, "Uncovered zones")
	require.Contains(t, summary, "architecture: no evidence (checks not run)")
	require.Contains(t, summary, "Gate status: NOT_CONFIGURED")
}

func TestFormatTerminalSummaryWithChecks(t *testing.T) {
	r := skeletonReport(t.TempDir())
	r.Checks = []VerificationCheck{
		{ID: "1", Name: "passed-check", Status: CheckStatusPassed},
		{ID: "2", Name: "warn-check", Status: CheckStatusWarn},
		{ID: "3", Name: "failed-check", Status: CheckStatusFailed},
		{ID: "4", Name: "skipped-check", Status: CheckStatusSkipped},
	}
	summary := FormatTerminalSummary(r)
	require.Contains(t, summary, "✓ passed-check")
	require.Contains(t, summary, "⚠ warn-check")
	require.Contains(t, summary, "✗ failed-check")
	require.Contains(t, summary, "○ skipped-check")
}

func TestToTrustMarkdownFullSections(t *testing.T) {
	r := skeletonReport("/repo")
	r.Task = "task-42"
	r.Checks = []VerificationCheck{
		{ID: "c1", Name: "flows", Status: CheckStatusPassed},
	}
	md := toTrustMarkdown(r)
	require.Contains(t, md, "- Task: `task-42`")
	require.Contains(t, md, "## Checks")
	require.Contains(t, md, "`c1` [passed] flows")
	require.Contains(t, md, "- Reason: verification gates not configured")
	require.Contains(t, md, "Architecture: 0%")
	require.NotContains(t, md, ConfidenceUnavailableLabel)
	require.NotContains(t, md, ConfidenceInferredCapLabel)
}

func TestToTrustMarkdownInferredDimensions(t *testing.T) {
	r := skeletonReport("/repo")
	r.Checks = []VerificationCheck{{ID: "c1", Name: "flows", Status: CheckStatusPassed}}
	r.Confidence = confidence.Report{
		Architecture:       0.8,
		Implementation:     0.7,
		FlowIntegrity:      0.75,
		Observability:      0.5,
		Security:           0.5,
		Regression:         0.6,
		Overall:            0.65,
		InferredDimensions: []string{"observability", "security"},
	}
	md := toTrustMarkdown(r)
	require.Contains(t, md, ConfidenceInferredCapLabel)
	require.NotContains(t, md, "Observability: 50%")
}

func TestToTrustMarkdownNoChecksShowsNA(t *testing.T) {
	md := toTrustMarkdown(skeletonReport("/repo"))
	require.Contains(t, md, ConfidenceUnavailableLabel)
	require.NotContains(t, md, "Architecture: 0%")
	require.Contains(t, md, "## Uncovered zones")
}

func TestTrustReportGoldenMarkdown(t *testing.T) {
	got := toTrustMarkdown(skeletonReport("/tmp/repo"))
	golden := filepath.Join("testdata", "report_skeleton.md")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		require.NoError(t, os.WriteFile(golden, []byte(got), 0o644))
	}
	want, err := os.ReadFile(golden)
	require.NoError(t, err)
	require.Equal(t, string(want), got)
}

func skeletonReport(repo string) TrustReport {
	conf := confidence.Report{
		Limits: []string{
			"lot 1 skeleton: confidence derived from zero executed checks",
			"no scorer, weighting, or normalization pipeline active yet",
		},
		UncoveredZones: []string{
			"architecture: no evidence (checks not run)",
			"implementation: no evidence (checks not run)",
			"flow_integrity: no evidence (checks not run)",
			"observability: no evidence (checks not run)",
			"security: no evidence (checks not run)",
			"regression: no evidence (checks not run)",
		},
	}
	return TrustReport{
		TrustID:      "trust-2026-05-29-golden",
		GeneratedAt:  "2026-05-29T12:00:00Z",
		Flow:         "onboarding-flow",
		Branch:       "main",
		Repository:   repo,
		Checks:       []VerificationCheck{},
		Confidence:   conf,
		Warnings:     append([]string(nil), conf.Limits...),
		ResidualRisk: ResidualRiskUnknown,
		Gate: GateEvaluation{
			Status: GateStatusNotConfigured,
			Reason: "verification gates not configured",
		},
	}
}
