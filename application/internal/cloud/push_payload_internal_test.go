package cloud

import (
	"encoding/json"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
	"github.com/stretchr/testify/require"
)

func TestBuildRunCreateRequestSameTaskID(t *testing.T) {
	entry := agentledger.Entry{
		TaskID: "task-stable", RunID: "run-1", Feature: "f", AgentID: "dev",
		StartedAt: "2026-01-01T10:00:00Z", EndedAt: "2026-01-01T10:01:00Z",
		PromptHash: "a", OutputHash: "b",
	}
	g := summarizeRunGroup("run-1", []agentledger.Entry{entry})

	req1 := BuildRunCreateRequest("/api/v1/projects/p1", g)
	req2 := BuildRunCreateRequest("/api/v1/projects/p1", g)
	b1, err := json.Marshal(req1)
	require.NoError(t, err)
	b2, err := json.Marshal(req2)
	require.NoError(t, err)
	require.Equal(t, string(b1), string(b2))
	require.Equal(t, "task-stable", req1.TaskID)
}

func TestBuildLedgerCreateRequestStableLocalEntryID(t *testing.T) {
	entry := agentledger.Entry{
		RunID: "run-1", AgentID: "dev", Phase: "dev",
		StartedAt: "2026-01-01T10:00:00Z", PromptHash: "p", OutputHash: "o",
	}
	req1 := BuildLedgerCreateRequest("/api/v1/runs/r1", entry)
	req2 := BuildLedgerCreateRequest("/api/v1/runs/r1", entry)
	b1, err := json.Marshal(req1)
	require.NoError(t, err)
	b2, err := json.Marshal(req2)
	require.NoError(t, err)
	require.Equal(t, string(b1), string(b2))
	require.Equal(t, agentledger.LocalEntryID(entry), req1.LocalEntryID)
}

func TestBuildLedgerCreateRequestUsageFields(t *testing.T) {
	entry := agentledger.Entry{
		RunID: "run-1", AgentID: "dev", Phase: "dev",
		StartedAt: "2026-01-01T10:00:00Z", PromptHash: "p", OutputHash: "o",
		Provider: "anthropic", Model: "claude-sonnet-4",
		InputTokens: 12400, OutputTokens: 3200, CostUSD: 0.042,
	}
	req := BuildLedgerCreateRequest("/api/v1/runs/r1", entry)
	require.Equal(t, "anthropic", req.Provider)
	require.Equal(t, "claude-sonnet-4", req.Model)
	require.NotNil(t, req.InputTokens)
	require.Equal(t, 12400, *req.InputTokens)
	require.NotNil(t, req.OutputTokens)
	require.Equal(t, 3200, *req.OutputTokens)
	require.NotNil(t, req.CostUsd)
	require.InDelta(t, 0.042, *req.CostUsd, 0.0001)

	raw, err := json.Marshal(req)
	require.NoError(t, err)
	body := string(raw)
	require.Contains(t, body, `"provider":"anthropic"`)
	require.Contains(t, body, `"model":"claude-sonnet-4"`)
	require.Contains(t, body, `"inputTokens":12400`)
	require.Contains(t, body, `"outputTokens":3200`)
	require.Contains(t, body, `"costUsd":0.042`)
}
