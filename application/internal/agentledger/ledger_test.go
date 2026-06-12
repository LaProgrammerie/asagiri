package agentledger_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agent"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func sampleResult() agent.RunResult {
	start := time.Date(2026, 6, 8, 12, 0, 0, 0, time.UTC)
	end := start.Add(1500 * time.Millisecond)
	return agent.RunResult{
		ExitCode:  0,
		Stdout:    `{"status":"ok","summary":"done"}`,
		StartedAt: start.Format(time.RFC3339Nano),
		EndedAt:   end.Format(time.RFC3339Nano),
	}
}

func TestRecordAndListRoundTrip(t *testing.T) {
	dir := t.TempDir()
	valid := true
	require.NoError(t, agentledger.Record(dir, agentledger.Params{
		TaskID:        "task-1",
		RunID:         "run-1",
		Feature:       "feat",
		AgentKey:      "dev",
		AgentID:       "dev",
		Role:          "dev",
		Provider:      "exec",
		Phase:         "dev",
		Prompt:        "prompt body",
		ContextHash:   "ctxhash",
		ContractValid: &valid,
		LogDir:        ".asagiri/logs/task-1/agents/dev",
		Result:        sampleResult(),
	}))

	report, err := agentledger.List(dir, agentledger.ListOptions{})
	require.NoError(t, err)
	require.Equal(t, agentledger.ReportVersion, report.ReportVersion)
	require.Len(t, report.Entries, 1)
	e := report.Entries[0]
	require.Equal(t, "task-1", e.TaskID)
	require.Equal(t, "run-1", e.RunID)
	require.Equal(t, int64(1500), e.DurationMS)
	require.NotEmpty(t, e.PromptHash)
	require.NotEmpty(t, e.OutputHash)
	require.NotNil(t, e.ContractValid)
	require.True(t, *e.ContractValid)

	filtered, err := agentledger.List(dir, agentledger.ListOptions{TaskID: "other"})
	require.NoError(t, err)
	require.Empty(t, filtered.Entries)

	filtered, err = agentledger.List(dir, agentledger.ListOptions{TaskID: "task-1"})
	require.NoError(t, err)
	require.Len(t, filtered.Entries, 1)
}

func TestListMissingLedgerEmpty(t *testing.T) {
	report, err := agentledger.List(t.TempDir(), agentledger.ListOptions{})
	require.NoError(t, err)
	require.Equal(t, 0, report.Count)
}

func TestHashTextStable(t *testing.T) {
	h1 := agentledger.HashText("same")
	h2 := agentledger.HashText("same")
	require.Equal(t, h1, h2)
	require.NotEqual(t, h1, agentledger.HashText("other"))
}

func TestLedgerPathFileExists(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, agentledger.Append(dir, agentledger.EntryFromParams(agentledger.Params{
		TaskID: "t", RunID: "r", Feature: "f", AgentKey: "dev", Phase: "dev", Prompt: "p", Result: sampleResult(),
	})))
	require.FileExists(t, filepath.Join(dir, agentledger.LedgerPath()))
	data, err := os.ReadFile(filepath.Join(dir, agentledger.LedgerPath()))
	require.NoError(t, err)
	require.Contains(t, string(data), `"task_id":"t"`)
}
