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

type runTestStore struct {
	run   *sqlite.Run
	tasks []sqlite.Task
}

func (s *runTestStore) GetRun(id string) (*sqlite.Run, error) {
	if s.run == nil || s.run.ID != id {
		return nil, os.ErrNotExist
	}
	return s.run, nil
}

func (s *runTestStore) ListTasksByRun(runID string) ([]sqlite.Task, error) {
	if s.run == nil || s.run.ID != runID {
		return nil, nil
	}
	return s.tasks, nil
}

func TestBuildRunReport_AggregatesBlockedRisky(t *testing.T) {
	repo := t.TempDir()
	cfg := enrichActiveCfg(nil)
	store := &runTestStore{
		run: &sqlite.Run{ID: "run-1", Feature: "feat", Status: sqlite.StatusRunning},
		tasks: []sqlite.Task{
			{ID: "t-ok", RunID: "run-1", Feature: "feat", Status: asagiri.StatusEnriched, PayloadJSON: payloadWithGate("enrich", "pass", 0.9)},
			{ID: "t-bad", RunID: "run-1", Feature: "feat", Status: asagiri.StatusPlanned, PayloadJSON: payloadWithGate("enrich", "fail", 0.1)},
		},
	}
	report, err := BuildRunReport(repo, cfg, store, "run-1")
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.Equal(t, 2, report.TaskCount)
	require.NotEmpty(t, report.NextActions)
}

func TestBuildRunReport_PlanGateFail(t *testing.T) {
	repo := t.TempDir()
	runID := "run-plan-fail"
	logDir := filepath.Join(repo, ".asagiri", "logs", runID, "gates")
	require.NoError(t, os.MkdirAll(logDir, 0o755))
	doc := gates.NewLogDocument(runID, "run", "plan", "feat", "reviewer", gates.Result{
		Status:     gates.VerdictFail,
		Confidence: 0.2,
		Notes:      []string{"plan rejected"},
	}, "2026-06-06T12:00:00Z")
	body, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "plan.json"), body, 0o644))

	cfg := enrichActiveCfg(nil)
	store := &runTestStore{
		run: &sqlite.Run{ID: runID, Feature: "feat", Status: sqlite.StatusRunning},
		tasks: []sqlite.Task{
			{ID: "t-ok", RunID: runID, Feature: "feat", Status: asagiri.StatusEnriched, PayloadJSON: payloadWithGate("enrich", "pass", 0.9)},
		},
	}
	report, err := BuildRunReport(repo, cfg, store, runID)
	require.NoError(t, err)
	require.Equal(t, VerdictBlocked, report.Score.Verdict)
	require.NotNil(t, report.PlanGate)
	require.Equal(t, "fail", report.PlanGate.Status)
}

func TestBuildRunReport_NoTasks(t *testing.T) {
	store := &runTestStore{
		run: &sqlite.Run{ID: "run-empty", Feature: "feat", Status: sqlite.StatusRunning},
	}
	_, err := BuildRunReport(t.TempDir(), &config.Config{}, store, "run-empty")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no tasks")
}

func TestBuildRunReport_UnknownRun(t *testing.T) {
	store := &runTestStore{}
	_, err := BuildRunReport(t.TempDir(), &config.Config{}, store, "missing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "missing")
}

func TestFormatRunReport_Stable(t *testing.T) {
	report := RunTrustReport{
		Scope:     TrustScope{ID: "run-1", Feature: "feat", Status: "running"},
		TaskCount: 2,
		Score:     WorkTrustScore{Overall: 55, Verdict: VerdictBlocked, Summary: "blocked — score 55/100"},
		PlanGate:  &RunPlanGateSummary{Status: "pass", Confidence: 0.95},
		Tasks: []FeatureTaskSummary{
			{TaskID: "t1", Verdict: VerdictBlocked, Score: 30, Status: asagiri.StatusImplemented},
		},
		NextActions: []WorkTrustRecommendation{
			{Command: "asa next --feature feat", Rationale: "t1 — Bloqué (implemented)"},
		},
	}
	out := FormatRunReport(report, FormatOptions{})
	require.Contains(t, out, "Summary")
	require.Contains(t, out, "Verdict: Bloqué")
	require.Contains(t, out, "Run: run-1")
	require.Contains(t, out, "plan (run)  OK")
	require.Contains(t, out, "Next actions")
}
