package workflow

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/stretchr/testify/require"
)

func writePlanFeatureFixture(t *testing.T, repo string) {
	t.Helper()
	featureDir := filepath.Join(repo, ".kiro", "specs", "agentflow-test")
	require.NoError(t, os.MkdirAll(featureDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "tasks.md"), []byte("- [ ] task A\n- [ ] task B\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "requirements.md"), []byte("R1: coverage\nR2: quality\n"), 0o644))
}

func gateLogJSONPath(repo, scopeID, gateName string) string {
	return filepath.Join(repo, ".asagiri", "logs", scopeID, "gates", gateName+".json")
}

func gateLogMarkdownPath(repo, scopeID, gateName string) string {
	return filepath.Join(repo, ".asagiri", "logs", scopeID, "gates", gateName+".log")
}

func planGateTestConfig(enabled bool) *config.Config {
	return &config.Config{
		Project: config.Project{DefaultBranch: "main"},
		Specs: config.Specs{
			KiroPath:       ".kiro/specs",
			ActiveSpecPath: "docs/ai/active/current-spec.md",
			HandoffPath:    "docs/ai/active/handoff.md",
		},
		State: config.State{Backend: "sqlite", Path: ".asagiri/state.sqlite"},
		Worktrees: config.Worktrees{
			BasePath:      ".asagiri/worktrees",
			BranchPrefix:  "asa",
			CleanupPolicy: "keep_failed",
		},
		Agents: map[string]config.Agent{
			"reviewer": {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				Plan: config.WorkPlanGateConfig{
					Enabled: enabled,
					Agent:   "reviewer",
					FailOn: []string{
						"missing_requirement_coverage",
						"oversized_task",
						"invalid_dependency",
					},
				},
			},
		},
	}
}

func newPlanGateService(t *testing.T, repo string, enabled bool) (*Service, *sqlite.Store) {
	t.Helper()
	cfg := planGateTestConfig(enabled)
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return NewService(repo, cfg, store, false), store
}

func passPlanGateYAML() string {
	return `plan_gate:
  status: pass
  confidence: 0.95
  notes:
    - plan ok
`
}

func warnPlanGateYAML() string {
	return `plan_gate:
  status: warn
  confidence: 0.7
  findings:
    - code: other
      severity: warn
      message: minor gap
`
}

func failPlanGateYAML() string {
	return `plan_gate:
  status: fail
  confidence: 0.2
  findings:
    - code: missing_requirement_coverage
      severity: fail
      message: Requirement R3 absent des tasks
      actions:
        - Ajouter une tâche couvrant R3
`
}

func TestPlanGateDisabledPlanFeatureUnchanged(t *testing.T) {
	repo := t.TempDir()
	writePlanFeatureFixture(t, repo)
	svc, _ := newPlanGateService(t, repo, false)

	runID, tasks, err := svc.PlanFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	logJSON := gateLogJSONPath(repo, runID, "plan")
	require.NoFileExists(t, logJSON)
}

func TestPlanGatePassPersistsTasks(t *testing.T) {
	repo := t.TempDir()
	writePlanFeatureFixture(t, repo)
	svc, store := newPlanGateService(t, repo, true)
	svc.planGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return passPlanGateYAML(), nil
	}

	runID, tasks, err := svc.PlanFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	stored, err := store.ListTasksByFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, stored, 2)

	logJSON := gateLogJSONPath(repo, runID, "plan")
	require.FileExists(t, logJSON)
	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	require.Contains(t, string(body), `"status": "pass"`)
}

func TestPlanGateWarnAdvisoryPersistsTasksAndLog(t *testing.T) {
	repo := t.TempDir()
	writePlanFeatureFixture(t, repo)
	svc, store := newPlanGateService(t, repo, true)
	svc.planGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnPlanGateYAML(), nil
	}

	runID, tasks, err := svc.PlanFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	stored, err := store.ListTasksByFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, stored, 2)

	logJSON := gateLogJSONPath(repo, runID, "plan")
	require.FileExists(t, logJSON)
	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	require.Contains(t, string(body), `"status": "warn"`)
}

func TestPlanGateFailBlocksPlan(t *testing.T) {
	repo := t.TempDir()
	writePlanFeatureFixture(t, repo)
	svc, store := newPlanGateService(t, repo, true)
	svc.planGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return failPlanGateYAML(), nil
	}

	runID, _, err := svc.PlanFeature("agentflow-test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "plan gate failed")
	require.Contains(t, err.Error(), "missing_requirement_coverage")

	stored, err := store.ListTasksByFeature("agentflow-test")
	require.NoError(t, err)
	require.Empty(t, stored)

	run, err := store.GetRun(runID)
	require.NoError(t, err)
	require.Equal(t, sqlite.StatusFailed, run.Status)

	logJSON := gateLogJSONPath(repo, runID, "plan")
	require.FileExists(t, logJSON)
}

func TestPlanGateInvalidParseFails(t *testing.T) {
	repo := t.TempDir()
	writePlanFeatureFixture(t, repo)
	svc, store := newPlanGateService(t, repo, true)
	svc.planGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return "not valid gate output", nil
	}

	_, _, err := svc.PlanFeature("agentflow-test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "plan gate failed")

	stored, err := store.ListTasksByFeature("agentflow-test")
	require.NoError(t, err)
	require.Empty(t, stored)
}

func TestPlanGateUsesGatesParseAndClassify(t *testing.T) {
	stdout := failPlanGateYAML()
	parsed := gates.ParseResult(stdout, planGateParseConfig)
	require.Equal(t, gates.VerdictFail, parsed.Status)
	classified := gates.ClassifyResult(parsed, []string{"missing_requirement_coverage"})
	require.Equal(t, gates.VerdictFail, classified.Status)
	require.NotEmpty(t, classified.Findings)
}

func TestPlanGateDryRunSimulatedPass(t *testing.T) {
	repo := t.TempDir()
	writePlanFeatureFixture(t, repo)
	cfg := planGateTestConfig(true)
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	svc := NewService(repo, cfg, store, true)
	runID, tasks, err := svc.PlanFeature("agentflow-test")
	require.NoError(t, err)
	require.Len(t, tasks, 2)

	logJSON := gateLogJSONPath(repo, runID, "plan")
	require.FileExists(t, logJSON)
	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	require.Contains(t, string(body), `"dry_run": true`)
}
