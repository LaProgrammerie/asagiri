package agentledger_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func TestInspectRunNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := agentledger.Inspect(dir, "missing-run")
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")
}

func TestInspectRunPresentWithArtifacts(t *testing.T) {
	dir := t.TempDir()
	valid := true
	logDir := ".asagiri/logs/task-1/agents/dev"
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-1", RunID: "run-abc", Feature: "feat", AgentID: "dev",
		Role: "dev", Provider: "exec", Phase: "dev",
		StartedAt: "2026-06-08T12:00:00Z", EndedAt: "2026-06-08T12:00:01Z",
		DurationMS: 1000, ExitCode: 0,
		PromptHash: "ph", ContextHash: "ch", OutputHash: "oh",
		ContractValid: &valid, LogDir: logDir,
	}))

	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	mod := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	for _, name := range []string{"prompt.md", "invocation.json", "context.json", "contract.json"} {
		p := filepath.Join(absLog, name)
		require.NoError(t, os.WriteFile(p, []byte("x"), 0o644))
		require.NoError(t, os.Chtimes(p, mod, mod))
	}

	report, err := agentledger.Inspect(dir, "run-abc")
	require.NoError(t, err)
	require.Equal(t, agentledger.InspectReportVersion, report.ReportVersion)
	require.Equal(t, "task-1", report.TaskID)
	require.Equal(t, "dev", report.AgentID)
	require.Len(t, report.Artifacts, 4)
	for _, a := range report.Artifacts {
		require.True(t, a.Exists, a.Name)
		require.NotNil(t, a.SizeBytes)
		require.Equal(t, int64(1), *a.SizeBytes)
		require.NotEmpty(t, a.ModifiedAt)
		require.Contains(t, a.Path, logDir)
	}
}

func TestInspectRunPresentMissingLogs(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, agentledger.Append(dir, agentledger.Entry{
		TaskID: "task-2", RunID: "run-empty", Feature: "f", AgentID: "dev",
		LogDir: ".asagiri/logs/task-2/agents/dev",
	}))

	report, err := agentledger.Inspect(dir, "run-empty")
	require.NoError(t, err)
	require.Len(t, report.Artifacts, 4)
	for _, a := range report.Artifacts {
		require.False(t, a.Exists)
		require.Nil(t, a.SizeBytes)
		require.Empty(t, a.ModifiedAt)
	}
}
