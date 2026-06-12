package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func patchConfigHumanReviewEnabled(t *testing.T, repo string) {
	t.Helper()
	cfgPath := filepath.Join(repo, ".asagiri", "config.yaml")
	data, err := os.ReadFile(cfgPath)
	require.NoError(t, err)
	content := string(data)
	if strings.Contains(content, "human_review:") {
		if strings.Contains(content, "    human_review:\n      enabled: false") {
			content = strings.Replace(content,
				"    human_review:\n      enabled: false",
				"    human_review:\n      enabled: true", 1)
			require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))
		}
		return
	}
	block := `
    human_review:
      enabled: true
      mode: per-task
`
	const anchor = "      max_retries: 2"
	if strings.Contains(content, anchor) {
		content = strings.Replace(content, anchor, anchor+block, 1)
	} else {
		content += "\nwork:\n  gates:" + block
	}
	require.NoError(t, os.WriteFile(cfgPath, []byte(content), 0o644))
}

func seedImplementedTask(t *testing.T, repo, taskID string) {
	t.Helper()
	dbPath := filepath.Join(repo, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	run := &sqlite.Run{ID: "run-gate", Feature: "feat", Status: sqlite.StatusRunning}
	require.NoError(t, store.CreateRun(run))
	payload, err := json.Marshal(asagiri.Task{ID: taskID, Title: "t", Feature: "feat", Status: asagiri.StatusImplemented})
	require.NoError(t, err)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: taskID, RunID: run.ID, Feature: "feat", Status: asagiri.StatusImplemented, PayloadJSON: string(payload),
	}))
}

func TestContinueShowsHumanReviewWhenPending(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	patchConfigHumanReviewEnabled(t, dir)
	seedImplementedTask(t, dir, "task-hr-cont")

	var out bytes.Buffer
	root = newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"continue", "--feature", "feat", "--dry-run"})
	require.NoError(t, root.Execute())

	text := out.String()
	require.Contains(t, text, "Gate human_review requires action")
	require.Contains(t, text, "gates submit human_review --task task-hr-cont")
	require.NotContains(t, text, "verify feat")
}

func TestContinueYesRejectedWhenHumanReviewSubmitPending(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	patchConfigHumanReviewEnabled(t, dir)
	seedImplementedTask(t, dir, "task-hr-yes")

	root = newRootCmd()
	root.SetArgs([]string{"continue", "--feature", "feat", "--yes", "--dry-run"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "complete the gate action before --yes")
}

func TestContinueAfterHumanReviewPassRecommendsVerify(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	patchConfigHumanReviewEnabled(t, dir)

	dbPath := filepath.Join(dir, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	run := &sqlite.Run{ID: "run-ok", Feature: "feat", Status: sqlite.StatusRunning}
	require.NoError(t, store.CreateRun(run))
	task := asagiri.Task{
		ID: "task-hr-done", Title: "t", Feature: "feat", Status: asagiri.StatusImplemented,
		Gates: &asagiri.TaskGates{History: []asagiri.GateHistoryEntry{
			{Gate: "human_review", Status: "pass", Confidence: 1},
		}},
	}
	payload, err := json.Marshal(task)
	require.NoError(t, err)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-hr-done", RunID: run.ID, Feature: "feat", Status: asagiri.StatusImplemented, PayloadJSON: string(payload),
	}))

	var out bytes.Buffer
	root = newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"continue", "--feature", "feat", "--dry-run"})
	require.NoError(t, root.Execute())

	text := out.String()
	require.Contains(t, text, "verify")
	require.NotContains(t, text, "Gate human_review requires action")
}

func writeHumanReviewVerdictFile(t *testing.T, repo, taskID string) {
	t.Helper()
	dir := filepath.Join(repo, ".asagiri", "logs", taskID, "gates")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := "human_review:\n  status: pass\n  confidence: 1.0\n  notes:\n    - reviewed\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, "human_review.verdict.yaml"), []byte(body), 0o644))
}

func TestContinueYesAllowedWhenHumanReviewResumePending(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)
	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	patchConfigHumanReviewEnabled(t, dir)
	seedImplementedTask(t, dir, "task-hr-resume")
	writeHumanReviewVerdictFile(t, dir, "task-hr-resume")

	var preview bytes.Buffer
	root = newRootCmd()
	root.SetOut(&preview)
	root.SetErr(&preview)
	root.SetArgs([]string{"continue", "--feature", "feat", "--dry-run"})
	require.NoError(t, root.Execute())
	previewText := preview.String()
	require.Contains(t, previewText, "asa continue --yes")
	require.NotContains(t, previewText, "gates submit human_review")

	var out bytes.Buffer
	root = newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"--dry-run", "continue", "--feature", "feat", "--yes"})
	err := root.Execute()
	require.NoError(t, err)
	require.NotContains(t, out.String(), "complete the gate action before --yes")
	require.NotContains(t, out.String(), "gates submit human_review")
}
