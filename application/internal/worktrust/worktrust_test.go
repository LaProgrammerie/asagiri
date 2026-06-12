package worktrust

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestBuildTaskReport_NoGates(t *testing.T) {
	task := sqlite.Task{
		ID: "task-plain", Feature: "feat", Status: asagiri.StatusPlanned,
		PayloadJSON: `{"title":"plain"}`,
	}
	report, err := BuildTaskReport(t.TempDir(), &config.Config{}, task)
	require.NoError(t, err)
	require.Equal(t, VerdictAcceptable, report.Score.Verdict)
	require.Greater(t, report.Score.Overall, 40.0)
	require.Equal(t, "task", report.Scope.Kind)
}

func TestBuildTaskReport_EnrichPass(t *testing.T) {
	cfg := enrichActiveCfg(nil)
	payload := payloadWithGate("enrich", "pass", 0.9)
	task := sqlite.Task{
		ID: "task-enrich", Feature: "feat", Status: asagiri.StatusEnriched,
		PayloadJSON: payload,
	}
	report, err := BuildTaskReport(t.TempDir(), cfg, task)
	require.NoError(t, err)
	require.GreaterOrEqual(t, report.Score.Overall, 60.0)
	require.Contains(t, report.Recommendation.Command, "asa next")
	require.False(t, report.Score.Verdict == VerdictBlocked)
}

func TestBuildTaskReport_VerifyEvidencePassWithValidation(t *testing.T) {
	dir := t.TempDir()
	cfg := verifyEvidenceActiveCfg(nil)
	payload := payloadWithGate("verify_evidence", "pass", 0.85)
	task := sqlite.Task{
		ID: "task-ve", Feature: "feat", Status: asagiri.StatusVerified,
		PayloadJSON: payload,
	}
	writeValidationResults(t, dir, task.ID, []validationEvidenceCommand{
		{Name: "test", Command: "go test ./...", ExitCode: 0},
	})

	report, err := BuildTaskReport(dir, cfg, task)
	require.NoError(t, err)
	require.GreaterOrEqual(t, report.Score.Overall, 75.0)
	require.Equal(t, VerdictTrusted, report.Score.Verdict)
	dim := findDimension(report, DimValidationStrength)
	require.GreaterOrEqual(t, dim.Score, 85.0)
}

func TestBuildTaskReport_HumanReviewPendingBlocked(t *testing.T) {
	dir := t.TempDir()
	cfg := hrActiveCfg(nil)
	task := sqlite.Task{
		ID: "task-hr", Feature: "feat", Status: asagiri.StatusImplemented,
		PayloadJSON: `{}`,
	}
	report, err := BuildTaskReport(dir, cfg, task)
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.LessOrEqual(t, report.Score.Overall, 40.0)
	require.Contains(t, report.Recommendation.Command, "asa next")
	require.Equal(t, "required", report.Recommendation.Priority)
}

func TestBuildTaskReport_GateFailRiskyOrBlocked(t *testing.T) {
	cfg := enrichActiveCfg(nil)
	payload := payloadWithGate("enrich", "fail", 0.2)
	task := sqlite.Task{
		ID: "task-fail", Feature: "feat", Status: asagiri.StatusPlanned,
		PayloadJSON: payload,
	}
	report, err := BuildTaskReport(t.TempDir(), cfg, task)
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.NotEmpty(t, report.Findings)
}

func TestBuildTaskReport_VerifyFailedStatus(t *testing.T) {
	task := sqlite.Task{
		ID: "task-vf", Feature: "feat", Status: asagiri.StatusVerifyFailed,
		PayloadJSON: `{}`,
	}
	report, err := BuildTaskReport(t.TempDir(), &config.Config{}, task)
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.LessOrEqual(t, report.Score.Overall, 35.0)
}

func TestBuildTaskReport_FailedStatus(t *testing.T) {
	task := sqlite.Task{
		ID: "task-failed", Feature: "feat", Status: asagiri.StatusFailed,
		PayloadJSON: `{}`,
	}
	report, err := BuildTaskReport(t.TempDir(), &config.Config{}, task)
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.LessOrEqual(t, report.Score.Overall, 35.0)
}

func TestFormatTaskReport_Stable(t *testing.T) {
	report := WorkTrustReport{
		Scope: TrustScope{Feature: "feat", TaskID: "t1", Status: asagiri.StatusVerified},
		Score: WorkTrustScore{Overall: 78, Verdict: VerdictAcceptable, Summary: "acceptable — score 78/100"},
		Dimensions: []WorkTrustDimension{
			{ID: DimGateConfidence, Label: "Confiance gates", Score: 80, Status: DimStatusStrong, Summary: "enrich pass", SourceGates: []string{"enrich"}},
		},
		Findings: []WorkTrustFinding{
			{Severity: "medium", Source: "governance", Message: "spec drift"},
		},
		Recommendation: WorkTrustRecommendation{
			Command:   "asa review feat --task t1 --agent reviewer",
			Rationale: "task verified",
		},
	}
	out := FormatTaskReport(report, FormatOptions{})
	require.Contains(t, out, "Summary")
	require.Contains(t, out, "Verdict: Acceptable")
	require.Contains(t, out, "Gates")
	require.Contains(t, out, "Risks")
	require.Contains(t, out, "Next actions")
	require.Contains(t, out, "spec drift")
	require.NotContains(t, out, "78/100")
	require.Contains(t, out, "asa review feat --task t1 --agent reviewer")
}

func enrichActiveCfg(warnAdvisory *bool) *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{
					Enabled:        true,
					Mode:           config.GovernanceModePerTask,
					WarnIsAdvisory: warnAdvisory,
				},
			},
		},
	}
}

func verifyEvidenceActiveCfg(warnAdvisory *bool) *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				VerifyEvidence: config.WorkVerifyEvidenceGateConfig{
					Enabled:        true,
					Mode:           config.GovernanceModePerTask,
					WarnIsAdvisory: warnAdvisory,
				},
			},
		},
	}
}

func hrActiveCfg(warnAdvisory *bool) *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled:        true,
					Mode:           config.GovernanceModePerTask,
					WarnIsAdvisory: warnAdvisory,
				},
			},
		},
	}
}

func payloadWithGate(gate, status string, confidence float64) string {
	p := map[string]any{
		"gates": map[string]any{
			"history": []map[string]any{
				{
					"gate":       gate,
					"status":     status,
					"at":         "2026-06-06T12:00:00Z",
					"confidence": confidence,
				},
			},
		},
	}
	b, _ := json.Marshal(p)
	return string(b)
}

func writeValidationResults(t *testing.T, repoRoot, taskID string, cmds []validationEvidenceCommand) {
	t.Helper()
	dir := filepath.Join(repoRoot, ".asagiri", "logs", taskID, "validation")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	doc := validationEvidenceDocument{
		TaskID:   taskID,
		At:       "2026-06-06T12:00:00Z",
		Commands: cmds,
	}
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.json"), b, 0o644))
}

func findDimension(report WorkTrustReport, id DimensionID) WorkTrustDimension {
	for _, d := range report.Dimensions {
		if d.ID == id {
			return d
		}
	}
	return WorkTrustDimension{}
}

func TestBuildTaskReport_EnrichPassRecommendsDev(t *testing.T) {
	cfg := enrichActiveCfg(nil)
	payload := payloadWithGate(gates.EnrichGateName, "pass", 0.9)
	task := sqlite.Task{
		ID: "task-e2", Feature: "myfeat", Status: asagiri.StatusEnriched,
		PayloadJSON: payload,
	}
	report, err := BuildTaskReport(t.TempDir(), cfg, task)
	require.NoError(t, err)
	require.Contains(t, report.Recommendation.Command, "asa next --feature myfeat")
}
