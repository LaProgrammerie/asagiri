package workflow

import (
	"context"
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

func trustGateTestConfig(trustEnabled bool, hrEnabled bool, warnAdvisory *bool, minScore float64) *config.Config {
	if minScore <= 0 {
		minScore = 70
	}
	min := minScore
	trust := config.WorkTrustGateConfig{
		Enabled:        trustEnabled,
		Mode:           config.GovernanceModePerTask,
		MinScore:       &min,
		BlockVerdicts:  config.DefaultTrustGateBlockVerdicts(),
		WarnVerdicts:   config.DefaultTrustGateWarnVerdicts(),
		WarnIsAdvisory: warnAdvisory,
	}
	return &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs:   config.Specs{KiroPath: ".kiro/specs"},
		State:   config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:     ".asagiri/worktrees",
			BranchPrefix: "asa",
		},
		Agents: map[string]config.Agent{
			"reviewer": {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: hrEnabled,
					Mode:    config.GovernanceModePerTask,
				},
				Trust: trust,
			},
		},
	}
}

func newTrustGateService(t *testing.T, trustEnabled bool, hrEnabled bool, warnAdvisory *bool, minScore ...float64) (*Service, *sqlite.Store) {
	t.Helper()
	repo := t.TempDir()
	min := 70.0
	if len(minScore) > 0 {
		min = minScore[0]
	}
	cfg := trustGateTestConfig(trustEnabled, hrEnabled, warnAdvisory, min)
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return NewService(repo, cfg, store, false), store
}

func trustPassPayload(taskID string) string {
	payload := map[string]any{
		"task_id": taskID,
		"gates": map[string]any{
			"history": []map[string]any{
				{"gate": "enrich", "status": "pass", "at": "2026-06-08T12:00:00Z", "confidence": 0.9},
				{"gate": "human_review", "status": "pass", "at": "2026-06-08T12:00:00Z", "confidence": 0.9},
				{"gate": "verify_evidence", "status": "pass", "at": "2026-06-08T12:00:00Z", "confidence": 0.9},
			},
		},
		"validation_commands": []string{"echo ok"},
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func lastTrustHistoryEntry(t *testing.T, payloadJSON string) (asagiri.GateHistoryEntry, bool) {
	t.Helper()
	canonical := payloadToCanonicalMust(t, payloadJSON)
	if canonical.Gates == nil {
		return asagiri.GateHistoryEntry{}, false
	}
	for i := len(canonical.Gates.History) - 1; i >= 0; i-- {
		if canonical.Gates.History[i].Gate == trustGateName {
			return canonical.Gates.History[i], true
		}
	}
	return asagiri.GateHistoryEntry{}, false
}

func TestProcessTrustGateInactive(t *testing.T) {
	svc, store := newTrustGateService(t, false, false, nil)
	task := seedTask(t, store, "feat", "task-off", asagiri.StatusImplemented)
	require.NoError(t, svc.processTrustGate(context.Background(), "feat", task))

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.NotNil(t, fresh)
	require.NotContains(t, fresh.PayloadJSON, `"gate":"trust"`)
}

func writeTrustValidationResults(t *testing.T, repoRoot, taskID string) {
	t.Helper()
	dir := filepath.Join(repoRoot, ".asagiri", "logs", taskID, "validation")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	doc := map[string]any{
		"task_id": taskID,
		"at":      "2026-06-08T12:00:00Z",
		"commands": []map[string]any{
			{"name": "test", "command": "go test ./...", "exit_code": 0},
		},
	}
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "results.json"), b, 0o644))
}

func TestProcessTrustGatePersistsPass(t *testing.T) {
	svc, store := newTrustGateService(t, true, false, nil, 55)
	task := seedTask(t, store, "feat", "task-pass", asagiri.StatusVerified)
	task.PayloadJSON = trustPassPayload(task.ID)
	require.NoError(t, store.UpdateTask(&task))
	writeTrustValidationResults(t, svc.repoRoot, task.ID)

	require.NoError(t, svc.processTrustGate(context.Background(), "feat", task))

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	entry, ok := lastTrustHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "pass", entry.Status)

	logJSON := gateLogJSONPath(svc.repoRoot, task.ID, trustGateName)
	require.FileExists(t, logJSON)
}

func TestProcessTrustGatePersistsFailWhenHRBlocking(t *testing.T) {
	svc, store := newTrustGateService(t, true, true, nil)
	task := seedTask(t, store, "feat", "task-fail", asagiri.StatusImplemented)

	require.NoError(t, svc.processTrustGate(context.Background(), "feat", task))

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	entry, ok := lastTrustHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestReviewFeatureBlockedWhenTrustUnsatisfied(t *testing.T) {
	svc, store := newTrustGateService(t, true, false, nil)
	task := seedTask(t, store, "feat", "task-trust-block", asagiri.StatusVerified)

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "trust gate required before review")

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
}

func TestReviewFeatureAllowedWhenTrustSatisfied(t *testing.T) {
	svc, store := newTrustGateService(t, true, false, nil)
	task := seedTaskWithGateHistory(t, store, "feat", "task-trust-ok", asagiri.StatusVerified, gates.TrustGateName, "pass")

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusReviewed, fresh.Status)
}

func TestVerifyFeatureRunsTrustGate(t *testing.T) {
	svc, store := newTrustGateService(t, true, false, nil)
	task := seedTask(t, store, "feat", "task-verify-trust", asagiri.StatusImplemented)
	task.PayloadJSON = verifyCandidatePayload(task.ID)
	require.NoError(t, store.UpdateTask(&task))

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
	_, ok := lastTrustHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.FileExists(t, gateLogJSONPath(svc.repoRoot, task.ID, trustGateName))
}

func TestTrustGateFailBlocksReviewTaskStaysVerified(t *testing.T) {
	svc, store := newTrustGateService(t, true, false, nil)
	task := seedTask(t, store, "feat", "task-trust-fail-review", asagiri.StatusVerified)
	fail := gates.Result{
		GateID: "trust_gate", GateType: "trust_gate", Scope: task.ID,
		Status: gates.VerdictFail, Confidence: 0.2, Notes: []string{"blocked"},
	}
	require.NoError(t, svc.persistTrustGateVerdict("feat", task, fail))

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "trust gate required before review")

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
}

func TestReportIncludesTrustGateFromHistory(t *testing.T) {
	svc, store := newTrustGateService(t, true, false, nil)
	task := seedTask(t, store, "feat", "task-report", asagiri.StatusVerified)
	task.PayloadJSON = trustPassPayload(task.ID)
	require.NoError(t, store.UpdateTask(&task))
	writeTrustValidationResults(t, svc.repoRoot, task.ID)
	require.NoError(t, svc.processTrustGate(context.Background(), "feat", task))

	logJSON := gateLogJSONPath(svc.repoRoot, task.ID, trustGateName)
	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	require.Contains(t, string(body), `"gate_name": "trust"`)
}
