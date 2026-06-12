package intent_test

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func initTestGitRepo(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "config", "user.email", "test@example.com")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "config", "user.name", "Test")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	require.NoError(t, os.WriteFile(filepath.Join(dir, "README.md"), []byte("test\n"), 0o644))
	cmd = exec.Command("git", "add", "README.md")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = dir
	require.NoError(t, cmd.Run())
}

func executorRefreshFixture(t *testing.T, cfg *config.Config) (string, *sqlite.Store, *workflow.Service) {
	t.Helper()
	repo := t.TempDir()
	initTestGitRepo(t, repo)
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))
	store, err := sqlite.Open(filepath.Join(repo, ".asagiri", "state.sqlite"))
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())
	return repo, store, workflow.NewService(repo, cfg, store, false)
}

func seedExecutorTask(t *testing.T, store *sqlite.Store, feature, id, status, payload string) {
	t.Helper()
	run := &sqlite.Run{ID: "run-exec", Feature: feature, Status: sqlite.StatusRunning}
	require.NoError(t, store.CreateRun(run))
	if payload == "" {
		task := asagiri.Task{ID: id, Title: "t", Feature: feature, Status: status}
		b, err := json.Marshal(task)
		require.NoError(t, err)
		payload = string(b)
	}
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: id, RunID: run.ID, Feature: feature, Status: status, PayloadJSON: payload,
	}))
}

func enrichRefreshConfig(enabled bool) *config.Config {
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
			"dev":      {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			DefaultAgent:    "dev",
			DefaultEnricher: "reviewer",
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{
					Enabled: enabled,
					Mode:    config.GovernanceModePerTask,
					Agent:   "reviewer",
					FailOn:  config.DefaultEnrichGateFailOn(),
				},
			},
		},
	}
}

func verifyRefreshConfig(enabled bool) *config.Config {
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
			DefaultReviewer: "reviewer",
			AutoReview:      true,
			Gates: config.WorkGatesConfig{
				VerifyEvidence: config.WorkVerifyEvidenceGateConfig{
					Enabled: enabled,
					Mode:    config.GovernanceModePerTask,
					Agent:   "reviewer",
					FailOn:  config.DefaultVerifyEvidenceGateFailOn(),
				},
			},
		},
	}
}

func humanReviewRefreshConfig() *config.Config {
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
			"dev":      {Command: "echo", Args: []string{"ok"}},
		},
		Work: config.WorkConfig{
			DefaultAgent: "dev",
			AutoVerify:   true,
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
}

func withTask(feature, taskID string) []string {
	args := []string{feature}
	if taskID != "" {
		args = append(args, "--task", taskID)
	}
	return args
}

func executedCommands(res intent.ExecuteResult) []string {
	out := make([]string, 0, len(res.Executed))
	for _, line := range res.Executed {
		if len(line) >= 7 && line[:7] == "[plan] " {
			line = line[7:]
		}
		fields := splitFields(line)
		if len(fields) >= 2 && fields[0] == "asa" {
			out = append(out, fields[1])
		}
	}
	return out
}

func splitFields(s string) []string {
	var out []string
	cur := ""
	for _, r := range s {
		if r == ' ' {
			if cur != "" {
				out = append(out, cur)
				cur = ""
			}
			continue
		}
		cur += string(r)
	}
	if cur != "" {
		out = append(out, cur)
	}
	return out
}

func passEnrichGateYAML() string {
	return `enrich_gate:
  status: pass
  confidence: 0.95
  notes:
    - ready for dev
`
}

func passVerifyEvidenceGateYAML() string {
	return `verify_evidence_gate:
  status: pass
  confidence: 0.95
  notes:
    - validation evidence sufficient
`
}

func writeHumanReviewVerdict(t *testing.T, repo, taskID string) {
	t.Helper()
	dir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := "human_review:\n  status: pass\n  confidence: 1.0\n  notes:\n    - reviewed\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "human_review.verdict.yaml"), []byte(body), 0o644))
}

func TestExecutorRefreshEnrichThenDev(t *testing.T) {
	cfg := enrichRefreshConfig(true)
	repo, store, svc := executorRefreshFixture(t, cfg)
	svc.SetEnrichGateAgentHookForTest(func(_ context.Context, _ string) (string, error) {
		return passEnrichGateYAML(), nil
	})
	svc.SetDryRunForTest(true)

	const taskID = "task-enrich-dev"
	seedExecutorTask(t, store, "feat", taskID, asagiri.StatusPlanned, "")

	snap, err := intent.BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	fs := intent.FeatureStateFor(snap, "feat")
	require.True(t, fs.EnrichGateBlocksDev)

	plan := intent.ExecutionPlan{
		Intent: intent.ResolvedIntent{Feature: "feat", TaskID: taskID, Action: intent.IntentDevelop},
		Steps: []intent.PlanStep{
			{Command: "enrich", Args: withTask("feat", taskID), Condition: "task_not_enriched"},
			{Command: "dev", Args: withTask("feat", taskID), Condition: "task_ready_for_dev"},
		},
	}
	ex := &intent.Executor{Workflow: svc, Config: cfg, RepoRoot: repo, Store: store}
	res, err := ex.Execute(context.Background(), plan, snap, intent.WorkOptions{MaxTasks: 1})
	require.NoError(t, err)
	require.Equal(t, []string{"enrich", "dev"}, executedCommands(res))
}

func TestExecutorRefreshVerifyThenReview(t *testing.T) {
	cfg := verifyRefreshConfig(true)
	repo, store, svc := executorRefreshFixture(t, cfg)
	svc.SetVerifyEvidenceGateAgentHookForTest(func(_ context.Context, _ string) (string, error) {
		return passVerifyEvidenceGateYAML(), nil
	})

	const taskID = "task-verify-review"
	payload, err := json.Marshal(map[string]any{
		"validation_commands": []string{"echo ok"},
		"gates": map[string]any{
			"history": []map[string]any{
				{"gate": "verify_evidence", "status": "fail", "confidence": 0.2},
			},
		},
	})
	require.NoError(t, err)
	seedExecutorTask(t, store, "feat", taskID, asagiri.StatusVerified, string(payload))

	snap, err := intent.BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	fs := intent.FeatureStateFor(snap, "feat")
	require.True(t, fs.VerifyEvidenceGateBlocksReview)

	plan := intent.ExecutionPlan{
		Intent: intent.ResolvedIntent{Feature: "feat", TaskID: taskID, Action: intent.IntentDevelop},
		Steps: []intent.PlanStep{
			{Command: "verify", Args: append(withTask("feat", taskID), "--force"), Condition: "always"},
			{Command: "review", Args: append(withTask("feat", taskID), "--agent", "reviewer"), Condition: "review_enabled"},
		},
	}
	ex := &intent.Executor{Workflow: svc, Config: cfg, RepoRoot: repo, Store: store}
	res, err := ex.Execute(context.Background(), plan, snap, intent.WorkOptions{})
	require.NoError(t, err)
	require.Equal(t, []string{"verify", "review"}, executedCommands(res))
}

func TestExecutorRefreshHumanReviewResumeThenVerify(t *testing.T) {
	cfg := humanReviewRefreshConfig()
	repo, store, svc := executorRefreshFixture(t, cfg)
	svc.SetDryRunForTest(true)

	const taskID = "task-hr-resume"
	seedExecutorTask(t, store, "feat", taskID, asagiri.StatusImplemented, "")
	writeHumanReviewVerdict(t, repo, taskID)

	snap, err := intent.BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	fs := intent.FeatureStateFor(snap, "feat")
	require.NotNil(t, fs.PendingGate)
	require.Equal(t, gates.PendingPhaseResume, fs.PendingGate.Phase)

	plan := intent.ExecutionPlan{
		Intent: intent.ResolvedIntent{Feature: "feat", TaskID: taskID, Action: intent.IntentResume},
		Steps: []intent.PlanStep{
			{Command: "dev", Args: withTask("feat", taskID), Condition: "gate_needs_resume"},
			{Command: "verify", Args: withTask("feat", taskID), Condition: "implementation_done"},
		},
	}
	ex := &intent.Executor{Workflow: svc, Config: cfg, RepoRoot: repo, Store: store}
	res, err := ex.Execute(context.Background(), plan, snap, intent.WorkOptions{MaxTasks: 1})
	require.NoError(t, err)
	require.Equal(t, []string{"dev", "verify"}, executedCommands(res))
}

func TestExecutorRefreshGatesDisabledNoRegression(t *testing.T) {
	cfg := enrichRefreshConfig(false)
	repo, store, svc := executorRefreshFixture(t, cfg)
	svc.SetDryRunForTest(true)

	const taskID = "task-no-gate"
	seedExecutorTask(t, store, "feat", taskID, asagiri.StatusPlanned, "")

	snap, err := intent.BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)

	plan := intent.ExecutionPlan{
		Intent: intent.ResolvedIntent{Feature: "feat", TaskID: taskID, Action: intent.IntentDevelop},
		Steps: []intent.PlanStep{
			{Command: "enrich", Args: withTask("feat", taskID), Condition: "task_not_enriched"},
			{Command: "dev", Args: withTask("feat", taskID), Condition: "task_ready_for_dev"},
		},
	}
	ex := &intent.Executor{Workflow: svc, Config: cfg, RepoRoot: repo, Store: store}
	res, err := ex.Execute(context.Background(), plan, snap, intent.WorkOptions{MaxTasks: 1})
	require.NoError(t, err)
	require.Equal(t, []string{"enrich", "dev"}, executedCommands(res))
}

func TestExecutorRefreshPreservesSubmitPending(t *testing.T) {
	cfg := humanReviewRefreshConfig()
	repo, store, _ := executorRefreshFixture(t, cfg)

	const taskID = "task-submit"
	seedExecutorTask(t, store, "feat", taskID, asagiri.StatusImplemented, "")

	snap, err := intent.BuildSnapshot(repo, cfg, store)
	require.NoError(t, err)
	fs := intent.FeatureStateFor(snap, "feat")
	require.NotNil(t, fs.PendingGate)
	require.Equal(t, gates.PendingPhaseSubmit, fs.PendingGate.Phase)

	refreshed := intent.RefreshFeatureTaskState(repo, cfg, store, "feat", fs)
	require.NotNil(t, refreshed.PendingGate)
	require.Equal(t, gates.PendingPhaseSubmit, refreshed.PendingGate.Phase)
	require.False(t, intent.EvaluateCondition("implementation_done", intent.ResolvedIntent{}, refreshed, intent.WorkOptions{}))
}

func TestExecutorWithoutStoreSkipsRefresh(t *testing.T) {
	cfg := enrichRefreshConfig(true)
	repo, store, svc := executorRefreshFixture(t, cfg)
	svc.SetEnrichGateAgentHookForTest(func(_ context.Context, _ string) (string, error) {
		return passEnrichGateYAML(), nil
	})
	svc.SetDryRunForTest(true)

	const taskID = "task-no-store"
	seedExecutorTask(t, store, "feat", taskID, asagiri.StatusPlanned, "")

	snap, _ := intent.BuildSnapshot(repo, cfg, store)
	plan := intent.ExecutionPlan{
		Intent: intent.ResolvedIntent{Feature: "feat", TaskID: taskID},
		Steps: []intent.PlanStep{
			{Command: "enrich", Args: withTask("feat", taskID), Condition: "always"},
			{Command: "dev", Args: withTask("feat", taskID), Condition: "task_ready_for_dev"},
		},
	}
	ex := &intent.Executor{Workflow: svc, Config: cfg}
	res, err := ex.Execute(context.Background(), plan, snap, intent.WorkOptions{MaxTasks: 1})
	require.NoError(t, err)
	require.Equal(t, []string{"enrich"}, executedCommands(res))
	require.Len(t, res.Skipped, 1)
}
