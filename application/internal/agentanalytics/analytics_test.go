package agentanalytics_test

import (
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentanalytics"
	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func appendEntry(t *testing.T, dir string, e agentledger.Entry) {
	t.Helper()
	require.NoError(t, agentledger.Append(dir, e))
}

func TestBuildGlobalAndGroups(t *testing.T) {
	dir := t.TempDir()
	valid := true
	invalid := false
	start := time.Date(2026, 6, 8, 10, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)
	later := time.Date(2026, 6, 8, 11, 0, 0, 0, time.UTC).Format(time.RFC3339Nano)

	appendEntry(t, dir, agentledger.Entry{
		TaskID: "t1", RunID: "r1", Feature: "f", AgentID: "dev", Provider: "exec",
		StartedAt: start, DurationMS: 100, ExitCode: 0, ContractValid: &valid,
	})
	appendEntry(t, dir, agentledger.Entry{
		TaskID: "t2", RunID: "r2", Feature: "f", AgentID: "dev", Provider: "exec",
		StartedAt: later, DurationMS: 200, ExitCode: 1, ContractValid: &invalid,
	})
	appendEntry(t, dir, agentledger.Entry{
		TaskID: "t3", RunID: "r3", Feature: "f", AgentID: "reviewer", Provider: "codex",
		StartedAt: later, DurationMS: 300, ExitCode: 0,
	})

	report, err := agentanalytics.Build(dir, agentanalytics.Options{})
	require.NoError(t, err)
	require.Equal(t, agentanalytics.ReportVersion, report.ReportVersion)
	require.Equal(t, 3, report.Global.TotalRuns)
	require.Equal(t, 2, report.Global.SuccessCount)
	require.Equal(t, 1, report.Global.FailureCount)
	require.InDelta(t, 200.0, report.Global.AvgDurationMS, 0.01)
	require.Equal(t, int64(300), report.Global.P95DurationMS)
	require.Equal(t, later, report.Global.LastRunAt)
	require.NotNil(t, report.Global.ContractValidRatio)
	require.InDelta(t, 0.5, *report.Global.ContractValidRatio, 0.01)

	require.Len(t, report.ByAgent, 2)
	require.Len(t, report.ByProvider, 2)
}

func TestBuildFilterByAgent(t *testing.T) {
	dir := t.TempDir()
	appendEntry(t, dir, agentledger.Entry{TaskID: "t1", AgentID: "dev", Provider: "exec", ExitCode: 0, DurationMS: 50})
	appendEntry(t, dir, agentledger.Entry{TaskID: "t2", AgentID: "reviewer", Provider: "codex", ExitCode: 0, DurationMS: 80})

	report, err := agentanalytics.Build(dir, agentanalytics.Options{AgentID: "dev"})
	require.NoError(t, err)
	require.Equal(t, 1, report.Global.TotalRuns)
	require.Equal(t, "dev", report.Filter.AgentID)
	require.Len(t, report.ByAgent, 1)
	require.Equal(t, "dev", report.ByAgent[0].ID)
}

func TestBuildFilterByProvider(t *testing.T) {
	dir := t.TempDir()
	appendEntry(t, dir, agentledger.Entry{TaskID: "t1", AgentID: "dev", Provider: "exec", ExitCode: 0, DurationMS: 50})
	appendEntry(t, dir, agentledger.Entry{TaskID: "t2", AgentID: "reviewer", Provider: "codex", ExitCode: 0, DurationMS: 80})

	report, err := agentanalytics.Build(dir, agentanalytics.Options{Provider: "codex"})
	require.NoError(t, err)
	require.Equal(t, 1, report.Global.TotalRuns)
	require.Equal(t, "codex", report.Filter.Provider)
	require.Len(t, report.ByProvider, 1)
	require.Equal(t, "codex", report.ByProvider[0].ID)
}

func TestBuildEmptyLedger(t *testing.T) {
	report, err := agentanalytics.Build(t.TempDir(), agentanalytics.Options{})
	require.NoError(t, err)
	require.Equal(t, 0, report.Global.TotalRuns)
	require.Nil(t, report.Global.ContractValidRatio)
}
