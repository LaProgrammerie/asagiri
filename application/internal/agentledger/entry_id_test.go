package agentledger_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func TestLedgerEntryIdStable(t *testing.T) {
	entry := agentledger.Entry{
		RunID:      "run-1",
		AgentID:    "dev",
		Phase:      "dev",
		StartedAt:  "2026-01-01T10:00:00Z",
		PromptHash: "abc",
		OutputHash: "def",
	}
	id1 := agentledger.LocalEntryID(entry)
	id2 := agentledger.LocalEntryID(entry)
	require.Equal(t, id1, id2)
	require.Len(t, id1, 64)
}

func TestLedgerEntryIdChangesWhenIdentityChanges(t *testing.T) {
	base := agentledger.Entry{
		RunID: "run-1", AgentID: "dev", Phase: "dev",
		StartedAt: "2026-01-01T10:00:00Z", PromptHash: "a", OutputHash: "b",
	}
	other := base
	other.OutputHash = "c"
	require.NotEqual(t, agentledger.LocalEntryID(base), agentledger.LocalEntryID(other))
}

func TestLedgerEntryIdIgnoresUsageFields(t *testing.T) {
	base := agentledger.Entry{
		RunID: "run-1", AgentID: "dev", Phase: "dev",
		StartedAt: "2026-01-01T10:00:00Z", PromptHash: "a", OutputHash: "b",
	}
	withUsage := base
	withUsage.Provider = "anthropic"
	withUsage.Model = "claude-sonnet-4"
	withUsage.InputTokens = 12400
	withUsage.OutputTokens = 3200
	withUsage.CostUSD = 0.042
	require.Equal(t, agentledger.LocalEntryID(base), agentledger.LocalEntryID(withUsage))
}

func TestLedgerEntryUsageRoundTrip(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, agentledger.Record(dir, agentledger.Params{
		TaskID: "task-1", RunID: "run-1", Feature: "feat", AgentKey: "dev",
		Provider: "anthropic", Model: "claude-sonnet-4",
		InputTokens: 12400, OutputTokens: 3200, CostUSD: 0.042,
		Phase: "dev", Prompt: "p", Result: sampleResult(),
	}))
	report, err := agentledger.List(dir, agentledger.ListOptions{})
	require.NoError(t, err)
	require.Len(t, report.Entries, 1)
	e := report.Entries[0]
	require.Equal(t, "anthropic", e.Provider)
	require.Equal(t, "claude-sonnet-4", e.Model)
	require.Equal(t, 12400, e.InputTokens)
	require.Equal(t, 3200, e.OutputTokens)
	require.InDelta(t, 0.042, e.CostUSD, 0.0001)
}

func TestLedgerEntryIdIgnoresCosmeticFields(t *testing.T) {
	a := agentledger.Entry{
		RunID: "run-1", AgentID: "dev", Phase: "dev",
		StartedAt: "2026-01-01T10:00:00Z", PromptHash: "a", OutputHash: "b",
		TaskID: "task-1", Feature: "feat", Role: "dev", LogDir: "logs",
	}
	b := a
	b.TaskID = "task-2"
	b.Feature = "other"
	b.LogDir = "other/logs"
	require.Equal(t, agentledger.LocalEntryID(a), agentledger.LocalEntryID(b))
}