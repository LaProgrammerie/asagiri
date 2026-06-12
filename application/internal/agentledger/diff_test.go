package agentledger_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func appendDiffRun(t *testing.T, dir string, e agentledger.Entry) {
	t.Helper()
	require.NoError(t, agentledger.Append(dir, e))
}

func TestDiffIdenticalRuns(t *testing.T) {
	dir := t.TempDir()
	valid := true
	logDir := ".asagiri/logs/task-1/agents/dev"
	entry := agentledger.Entry{
		TaskID: "task-1", RunID: "run-a", Feature: "feat", AgentID: "dev",
		Role: "dev", Provider: "exec", Phase: "dev",
		PromptHash: "ph", ContextHash: "ch", OutputHash: "oh",
		ContractValid: &valid, LogDir: logDir, ExitCode: 0, DurationMS: 100,
	}
	appendDiffRun(t, dir, entry)
	entry.RunID = "run-b"
	appendDiffRun(t, dir, entry)

	absLog := filepath.Join(dir, filepath.FromSlash(logDir))
	require.NoError(t, os.MkdirAll(absLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(absLog, "context.json"), []byte("{}"), 0o644))

	report, err := agentledger.Diff(dir, "run-a", "run-b")
	require.NoError(t, err)
	require.Equal(t, agentledger.DiffReportVersion, report.ReportVersion)
	require.True(t, report.Identical)
}

func TestDiffDifferentHashes(t *testing.T) {
	dir := t.TempDir()
	appendDiffRun(t, dir, agentledger.Entry{
		RunID: "run-left", TaskID: "t", AgentID: "dev",
		PromptHash: "hash-a", ContextHash: "c1", OutputHash: "o1",
	})
	appendDiffRun(t, dir, agentledger.Entry{
		RunID: "run-right", TaskID: "t", AgentID: "dev",
		PromptHash: "hash-b", ContextHash: "c2", OutputHash: "o2",
	})

	report, err := agentledger.Diff(dir, "run-left", "run-right")
	require.NoError(t, err)
	require.False(t, report.Identical)
	fieldEq := map[string]bool{}
	for _, f := range report.Fields {
		fieldEq[f.Field] = f.Equal
	}
	require.False(t, fieldEq["prompt_hash"])
	require.False(t, fieldEq["context_hash"])
	require.False(t, fieldEq["output_hash"])
}

func TestDiffMissingArtifacts(t *testing.T) {
	dir := t.TempDir()
	logLeft := ".asagiri/logs/t1/agents/dev"
	logRight := ".asagiri/logs/t2/agents/dev"
	appendDiffRun(t, dir, agentledger.Entry{RunID: "run-l", TaskID: "t1", AgentID: "dev", LogDir: logLeft})
	appendDiffRun(t, dir, agentledger.Entry{RunID: "run-r", TaskID: "t2", AgentID: "dev", LogDir: logRight})

	leftLog := filepath.Join(dir, filepath.FromSlash(logLeft))
	require.NoError(t, os.MkdirAll(leftLog, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(leftLog, "prompt.md"), []byte("p"), 0o644))

	report, err := agentledger.Diff(dir, "run-l", "run-r")
	require.NoError(t, err)
	require.False(t, report.Identical)
	var promptDiff agentledger.ArtifactDiff
	for _, a := range report.Artifacts {
		if a.Name == "prompt.md" {
			promptDiff = a
			break
		}
	}
	require.True(t, promptDiff.Left.Exists)
	require.False(t, promptDiff.Right.Exists)
	require.False(t, promptDiff.ExistsEqual)
}

func TestDiffRunNotFound(t *testing.T) {
	dir := t.TempDir()
	appendDiffRun(t, dir, agentledger.Entry{RunID: "only", TaskID: "t", AgentID: "dev"})

	_, err := agentledger.Diff(dir, "only", "missing")
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")

	_, err = agentledger.Diff(dir, "missing", "only")
	require.Error(t, err)
}
