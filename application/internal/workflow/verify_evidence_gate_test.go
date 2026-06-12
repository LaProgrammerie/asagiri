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
	appvalidate "github.com/LaProgrammerie/asagiri/application/internal/validation"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func verifyEvidenceGateTestConfig(enabled bool, warnAdvisory *bool) *config.Config {
	verify := config.WorkVerifyEvidenceGateConfig{
		Enabled: enabled,
		Mode:    config.GovernanceModePerTask,
		Agent:   "reviewer",
		FailOn:  config.DefaultVerifyEvidenceGateFailOn(),
	}
	if warnAdvisory != nil {
		verify.WarnIsAdvisory = warnAdvisory
	}
	return &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs: config.Specs{
			KiroPath:       ".kiro/specs",
			ActiveSpecPath: "docs/ai/active/current-spec.md",
			HandoffPath:    "docs/ai/active/handoff.md",
		},
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:     ".asagiri/worktrees",
			BranchPrefix: "asa",
		},
		Agents: map[string]config.Agent{
			"reviewer": {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				VerifyEvidence: verify,
			},
		},
	}
}

func newVerifyEvidenceGateService(t *testing.T, enabled bool, warnAdvisory *bool) (*Service, *sqlite.Store) {
	t.Helper()
	repo := t.TempDir()
	cfg := verifyEvidenceGateTestConfig(enabled, warnAdvisory)
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return NewService(repo, cfg, store, false), store
}

func sampleValidationResults() []appvalidate.Result {
	return []appvalidate.Result{
		{Name: "task-0", Command: "echo ok", ExitCode: 0, Output: "ok"},
	}
}

func verifyCandidatePayload(taskID string) string {
	return `{"task_id":"` + taskID + `","validation_commands":["echo ok"],"title":"Verify evidence test"}`
}

func passVerifyEvidenceGateYAML() string {
	return `verify_evidence_gate:
  status: pass
  confidence: 0.95
  notes:
    - validation evidence sufficient
`
}

func warnVerifyEvidenceGateYAML() string {
	return `verify_evidence_gate:
  status: warn
  confidence: 0.7
  findings:
    - code: weak_assertion_signal
      severity: warn
      message: output lacks strong signals
`
}

func failVerifyEvidenceGateYAML() string {
	return `verify_evidence_gate:
  status: fail
  confidence: 0.2
  findings:
    - code: verification_not_actionable
      severity: fail
      message: validation evidence does not support verify
      actions:
        - Re-run asa verify feat --task task-1 --force
`
}

func lastVerifyEvidenceHistoryEntry(t *testing.T, payloadJSON string) (asagiri.GateHistoryEntry, bool) {
	t.Helper()
	return gates.LastGateEntry(payloadJSON, verifyEvidenceGateName)
}

func TestProcessVerifyEvidenceGateDisabledNoOp(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, false, nil)
	task := seedTask(t, store, "feat", "task-off", asagiri.StatusImplemented)

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-off", verifyEvidenceGateName)
	require.NoFileExists(t, logJSON)

	fresh, err := store.GetTask("task-off")
	require.NoError(t, err)
	_, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.False(t, ok)
}

func TestProcessVerifyEvidenceGateDryRunPassHistoryAndLogs(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, true, nil)
	svc.dryRun = true
	task := seedTask(t, store, "feat", "task-dry", asagiri.StatusImplemented)

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-dry", verifyEvidenceGateName)
	require.FileExists(t, logJSON)
	logMD := gateLogMarkdownPath(svc.repoRoot, "task-dry", verifyEvidenceGateName)
	require.FileExists(t, logMD)

	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	var doc gates.LogDocument
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, "pass", doc.Status)
	require.True(t, doc.DryRun)

	fresh, err := store.GetTask("task-dry")
	require.NoError(t, err)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, verifyEvidenceGateName, entry.Gate)
	require.Equal(t, "pass", entry.Status)
	require.True(t, entry.DryRun)
}

func TestProcessVerifyEvidenceGatePassHistoryAndLogs(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, true, nil)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return passVerifyEvidenceGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-pass", asagiri.StatusImplemented)
	task.PayloadJSON = verifyCandidatePayload("task-pass")

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-pass", verifyEvidenceGateName)
	require.FileExists(t, logJSON)

	fresh, err := store.GetTask("task-pass")
	require.NoError(t, err)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "pass", entry.Status)
}

func TestProcessVerifyEvidenceGateWarnAdvisoryOK(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, true, nil)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnVerifyEvidenceGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-warn", asagiri.StatusImplemented)

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.NoError(t, err)

	fresh, err := store.GetTask("task-warn")
	require.NoError(t, err)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "warn", entry.Status)
}

func TestProcessVerifyEvidenceGateWarnNonAdvisoryError(t *testing.T) {
	f := false
	svc, store := newVerifyEvidenceGateService(t, true, &f)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnVerifyEvidenceGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-warn-block", asagiri.StatusImplemented)

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify evidence gate warn (non-advisory)")
}

func TestProcessVerifyEvidenceGateFailBlocks(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, true, nil)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return failVerifyEvidenceGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-fail", asagiri.StatusImplemented)

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify evidence gate failed")
	require.Contains(t, err.Error(), "verification_not_actionable")

	logJSON := gateLogJSONPath(svc.repoRoot, "task-fail", verifyEvidenceGateName)
	require.FileExists(t, logJSON)

	fresh, err := store.GetTask("task-fail")
	require.NoError(t, err)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestProcessVerifyEvidenceGateParseErrorBlocks(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, true, nil)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return "not a gate block", nil
	}
	task := seedTask(t, store, "feat", "task-parse", asagiri.StatusImplemented)

	err := svc.processVerifyEvidenceGate(context.Background(), "feat", task, sampleValidationResults())
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify evidence gate failed")

	fresh, err := store.GetTask("task-parse")
	require.NoError(t, err)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestPersistVerifyEvidenceGateVerdictDoesNotWriteGovernanceHistory(t *testing.T) {
	svc, store := newVerifyEvidenceGateService(t, true, nil)
	task := seedTask(t, store, "feat", "task-gov", asagiri.StatusImplemented)

	v := gates.Result{
		GateID:   "verify_evidence_gate",
		GateType: "verify_evidence_gate",
		Scope:    "task-gov",
		Status:   gates.VerdictPass,
	}
	require.NoError(t, svc.persistVerifyEvidenceGateVerdict("feat", task, v, passVerifyEvidenceGateYAML()))

	fresh, err := store.GetTask("task-gov")
	require.NoError(t, err)
	canonical := payloadToCanonicalMust(t, fresh.PayloadJSON)
	require.Nil(t, canonical.Governance)
	require.Len(t, canonical.Gates.History, 1)
	require.Equal(t, verifyEvidenceGateName, canonical.Gates.History[0].Gate)
}

func TestVerifyEvidenceGateUsesGatesParseAndClassify(t *testing.T) {
	parsed := gates.ParseResult(failVerifyEvidenceGateYAML(), verifyEvidenceGateParseConfig)
	require.Equal(t, gates.VerdictFail, parsed.Status)
	classified := gates.ClassifyResult(parsed, config.DefaultVerifyEvidenceGateFailOn())
	require.Equal(t, gates.VerdictFail, classified.Status)

	warnParsed := gates.ParseResult(warnVerifyEvidenceGateYAML(), verifyEvidenceGateParseConfig)
	warnClassified := gates.ClassifyResult(warnParsed, config.DefaultVerifyEvidenceGateFailOn())
	require.Equal(t, gates.VerdictWarn, warnClassified.Status)
}
