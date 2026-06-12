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

func enrichGateTestConfig(enabled bool, warnAdvisory *bool) *config.Config {
	enrich := config.WorkEnrichGateConfig{
		Enabled: enabled,
		Mode:    config.GovernanceModePerTask,
		Agent:   "reviewer",
		FailOn:  config.DefaultEnrichGateFailOn(),
	}
	if warnAdvisory != nil {
		enrich.WarnIsAdvisory = warnAdvisory
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
				Enrich: enrich,
			},
		},
	}
}

func newEnrichGateService(t *testing.T, enabled bool, warnAdvisory *bool) (*Service, *sqlite.Store) {
	t.Helper()
	repo := t.TempDir()
	cfg := enrichGateTestConfig(enabled, warnAdvisory)
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return NewService(repo, cfg, store, false), store
}

func enrichCandidatePayload(taskID string) string {
	return `{"task_id":"` + taskID + `","files_scope":["application/internal/workflow/"],"validation_commands":["make test"],"title":"Enrich test"}`
}

func passEnrichGateYAML() string {
	return `enrich_gate:
  status: pass
  confidence: 0.95
  notes:
    - ready for dev
`
}

func warnEnrichGateYAML() string {
	return `enrich_gate:
  status: warn
  confidence: 0.7
  findings:
    - code: other
      severity: warn
      message: minor gap
`
}

func emptyContextWarnEnrichGateYAML() string {
	return `enrich_gate:
  status: warn
  confidence: 0.6
  findings:
    - code: empty_context_when_required
      severity: warn
      message: no RAG context files attached
`
}

func failEnrichGateYAML() string {
	return `enrich_gate:
  status: fail
  confidence: 0.2
  findings:
    - code: enrichment_not_actionable
      severity: fail
      message: payload lacks actionable scope for dev
      actions:
        - Re-run asa enrich feat --task task-1 --force
`
}

func lastEnrichHistoryEntry(t *testing.T, payloadJSON string) (asagiri.GateHistoryEntry, bool) {
	t.Helper()
	entry, ok := gates.LastGateEntry(payloadJSON, enrichGateName)
	return entry, ok
}

func TestProcessEnrichGateDisabledNoOp(t *testing.T) {
	svc, store := newEnrichGateService(t, false, nil)
	task := seedTask(t, store, "feat", "task-off", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-off"), nil)
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-off", enrichGateName)
	require.NoFileExists(t, logJSON)

	fresh, err := store.GetTask("task-off")
	require.NoError(t, err)
	_, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.False(t, ok)
}

func TestProcessEnrichGateDryRunPassHistoryAndLogs(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	svc.dryRun = true
	task := seedTask(t, store, "feat", "task-dry", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-dry"), []string{"docs/ai/02-architecture.md"})
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-dry", enrichGateName)
	require.FileExists(t, logJSON)
	logMD := gateLogMarkdownPath(svc.repoRoot, "task-dry", enrichGateName)
	require.FileExists(t, logMD)

	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	var doc gates.LogDocument
	require.NoError(t, json.Unmarshal(body, &doc))
	require.Equal(t, "pass", doc.Status)
	require.True(t, doc.DryRun)

	fresh, err := store.GetTask("task-dry")
	require.NoError(t, err)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, enrichGateName, entry.Gate)
	require.Equal(t, "pass", entry.Status)
	require.True(t, entry.DryRun)
	require.Nil(t, payloadToCanonicalMust(t, fresh.PayloadJSON).Governance)
}

func payloadToCanonicalMust(t *testing.T, payloadJSON string) asagiri.Task {
	t.Helper()
	task, err := payloadToCanonical(payloadJSON)
	require.NoError(t, err)
	return task
}

func TestProcessEnrichGatePassHistoryAndLogs(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return passEnrichGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-pass", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-pass"), nil)
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-pass", enrichGateName)
	require.FileExists(t, logJSON)

	fresh, err := store.GetTask("task-pass")
	require.NoError(t, err)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "pass", entry.Status)
}

func TestProcessEnrichGateWarnAdvisoryOK(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnEnrichGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-warn", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-warn"), nil)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-warn")
	require.NoError(t, err)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "warn", entry.Status)
}

func TestProcessEnrichGateWarnNonAdvisoryError(t *testing.T) {
	f := false
	svc, store := newEnrichGateService(t, true, &f)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnEnrichGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-warn-block", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-warn-block"), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate warn (non-advisory)")
}

func TestProcessEnrichGateFailActionableError(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return failEnrichGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-fail", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-fail"), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate failed")
	require.Contains(t, err.Error(), "enrichment_not_actionable")

	logJSON := gateLogJSONPath(svc.repoRoot, "task-fail", enrichGateName)
	require.FileExists(t, logJSON)

	fresh, err := store.GetTask("task-fail")
	require.NoError(t, err)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestProcessEnrichGateParseErrorFails(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return "not a gate block", nil
	}
	task := seedTask(t, store, "feat", "task-parse", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-parse"), nil)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate failed")

	fresh, err := store.GetTask("task-parse")
	require.NoError(t, err)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestProcessEnrichGateEmptyContextWarnDoesNotBlock(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return emptyContextWarnEnrichGateYAML(), nil
	}
	task := seedTask(t, store, "feat", "task-ctx", asagiri.StatusPlanned)

	err := svc.processEnrichGate(context.Background(), "feat", task, enrichCandidatePayload("task-ctx"), nil)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-ctx")
	require.NoError(t, err)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "warn", entry.Status)
}

func TestEnrichGateUsesGatesParseAndClassify(t *testing.T) {
	stdout := failEnrichGateYAML()
	parsed := gates.ParseResult(stdout, enrichGateParseConfig)
	require.Equal(t, gates.VerdictFail, parsed.Status)
	classified := gates.ClassifyResult(parsed, config.DefaultEnrichGateFailOn())
	require.Equal(t, gates.VerdictFail, classified.Status)

	warnStdout := emptyContextWarnEnrichGateYAML()
	warnParsed := gates.ParseResult(warnStdout, enrichGateParseConfig)
	warnClassified := gates.ClassifyResult(warnParsed, config.DefaultEnrichGateFailOn())
	require.Equal(t, gates.VerdictWarn, warnClassified.Status)
}

func TestPersistEnrichGateVerdictDoesNotWriteGovernanceHistory(t *testing.T) {
	svc, store := newEnrichGateService(t, true, nil)
	task := seedTask(t, store, "feat", "task-gov", asagiri.StatusPlanned)

	v := gates.Result{
		GateID:   "enrich_gate",
		GateType: "enrich_gate",
		Scope:    "task-gov",
		Status:   gates.VerdictPass,
	}
	require.NoError(t, svc.persistEnrichGateVerdict("feat", task, v, passEnrichGateYAML()))

	fresh, err := store.GetTask("task-gov")
	require.NoError(t, err)
	canonical := payloadToCanonicalMust(t, fresh.PayloadJSON)
	require.Nil(t, canonical.Governance)
	require.Len(t, canonical.Gates.History, 1)
	require.Equal(t, enrichGateName, canonical.Gates.History[0].Gate)
}
