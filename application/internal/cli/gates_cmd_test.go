package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestGatesSubmitHumanReviewWritesExpectedPath(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	dbPath := filepath.Join(dir, ".asagiri", "state.sqlite")
	store, err := sqlite.Open(dbPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })
	require.NoError(t, store.Migrate())

	task := asagiri.Task{ID: "task-cli", Title: "t", Feature: "feat", Status: asagiri.StatusImplemented}
	payload, err := json.Marshal(task)
	require.NoError(t, err)
	require.NoError(t, store.CreateTask(&sqlite.Task{
		ID: "task-cli", RunID: "run-1", Feature: "feat", Status: asagiri.StatusImplemented, PayloadJSON: string(payload),
	}))

	var out bytes.Buffer
	root = newRootCmd()
	root.SetOut(&out)
	root.SetErr(&out)
	root.SetArgs([]string{"gates", "submit", "human_review", "--task", "task-cli", "--verdict", "pass", "--note", "LGTM"})
	require.NoError(t, root.Execute())

	wantPath := filepath.Join(dir, ".asagiri", "logs", "task-cli", "gates", "human_review.verdict.yaml")
	body, err := os.ReadFile(wantPath)
	require.NoError(t, err)
	require.Contains(t, string(body), "status: pass")
	require.Contains(t, string(body), "LGTM")
	require.Contains(t, out.String(), wantPath)
}

func TestGatesSubmitRejectsUnknownTask(t *testing.T) {
	dir := t.TempDir()
	initGitRepo(t, dir)
	writeExampleConfig(t, dir)

	old, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(old) })

	root := newRootCmd()
	root.SetArgs([]string{"init"})
	require.NoError(t, root.Execute())

	root = newRootCmd()
	root.SetArgs([]string{"gates", "submit", "human_review", "--task", "no-such-task", "--verdict", "pass"})
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, err.Error(), "no-such-task")
}
