package agentledger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func samplePreviewEntry(t *testing.T, dir, runID string) {
	t.Helper()
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-1", RunID: runID, Feature: "feat", AgentID: "dev",
		Provider: "exec", Phase: "dev",
		PromptHash:  agentledger.HashText("hello prompt"),
		ContextHash: "ctx-ledger", OutputHash: "out-ledger",
		LogDir: ".asagiri/logs/task-1/agents/dev",
	}))
}

func writePreviewArtifacts(t *testing.T, dir string) {
	logDir := filepath.Join(dir, ".asagiri", "logs", "task-1", "agents", "dev")
	require.NoError(t, os.MkdirAll(logDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "prompt.md"), []byte("hello prompt"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "context.json"), []byte(`{"task_id":"task-1"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "invocation.json"), []byte(`{"provider":"exec"}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "contract.json"), []byte(`{"valid":true}`), 0o644))
}

func TestReplayPreviewRunNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := agentledger.ReplayPreview(dir, "missing", agentledger.ReplayPreviewOptions{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")
}

func TestReplayPreviewFullArtifacts(t *testing.T) {
	dir := t.TempDir()
	samplePreviewEntry(t, dir, "run-prev")
	writePreviewArtifacts(t, dir)

	report, err := agentledger.ReplayPreview(dir, "run-prev", agentledger.ReplayPreviewOptions{})
	require.NoError(t, err)
	require.Equal(t, agentledger.ReplayPreviewReportVersion, report.ReportVersion)
	require.Equal(t, agentledger.HashText("hello prompt"), report.PromptHash)
	require.Len(t, report.Artifacts, 4)

	byName := map[string]agentledger.PreviewArtifact{}
	for _, a := range report.Artifacts {
		byName[a.Name] = a
	}
	require.True(t, byName["prompt.md"].Exists)
	require.Empty(t, byName["prompt.md"].Content)
	require.Contains(t, byName["context.json"].Content, "task-1")
	require.Contains(t, byName["invocation.json"].Content, "exec")
	require.Contains(t, byName["contract.json"].Content, "valid")
}

func TestReplayPreviewMissingArtifacts(t *testing.T) {
	dir := t.TempDir()
	samplePreviewEntry(t, dir, "run-empty")

	report, err := agentledger.ReplayPreview(dir, "run-empty", agentledger.ReplayPreviewOptions{})
	require.NoError(t, err)
	for _, a := range report.Artifacts {
		require.False(t, a.Exists)
		require.Empty(t, a.Content)
	}
}

func TestReplayPreviewIncludePrompt(t *testing.T) {
	dir := t.TempDir()
	samplePreviewEntry(t, dir, "run-prompt")
	writePreviewArtifacts(t, dir)

	report, err := agentledger.ReplayPreview(dir, "run-prompt", agentledger.ReplayPreviewOptions{IncludePrompt: true})
	require.NoError(t, err)
	var prompt agentledger.PreviewArtifact
	for _, a := range report.Artifacts {
		if a.Name == "prompt.md" {
			prompt = a
			break
		}
	}
	require.Equal(t, "hello prompt", prompt.Content)
}
