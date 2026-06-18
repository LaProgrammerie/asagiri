package cloud

import (
	"encoding/json"

	"github.com/LaProgrammerie/asagiri/application/internal/agentledger"
)

// BuildRunCreateRequest builds POST /api/v1/runs from grouped ledger entries.
func BuildRunCreateRequest(projectIRI string, g runGroup) RunCreateRequest {
	req := RunCreateRequest{
		Project:    projectIRI,
		LocalRunID: g.runID,
		TaskID:     g.taskID,
		Feature:    g.feature,
		Status:     g.status,
		StartedAt:  g.startedAt,
		EndedAt:    g.endedAt,
		DurationMs: g.durationMs,
	}
	if g.configReleaseVersion > 0 {
		v := g.configReleaseVersion
		req.ConfigReleaseVersion = &v
	}
	req.ConfigContentHash = g.configContentHash
	return req
}

// BuildLedgerCreateRequest builds POST /api/v1/ledger-entries from one ledger line.
func BuildLedgerCreateRequest(runIRI string, e agentledger.Entry) LedgerCreateRequest {
	raw, _ := json.Marshal(e)
	req := LedgerCreateRequest{
		Run:           runIRI,
		LocalEntryID:  agentledger.LocalEntryID(e),
		AgentID:       e.AgentID,
		Provider:      e.Provider,
		Model:         e.Model,
		InputTokens:   optionalInt(e.InputTokens),
		OutputTokens:  optionalInt(e.OutputTokens),
		CostUsd:       optionalCostUSD(e.CostUSD),
		Phase:         e.Phase,
		StartedAt:     e.StartedAt,
		EndedAt:       e.EndedAt,
		DurationMs:    e.DurationMS,
		ExitCode:      e.ExitCode,
		PromptHash:    e.PromptHash,
		ContextHash:   e.ContextHash,
		OutputHash:    e.OutputHash,
		ContractValid: e.ContractValid,
		LogDir:        e.LogDir,
		RawPayload:    raw,
	}
	if e.RegistryAgentVersion > 0 {
		v := e.RegistryAgentVersion
		req.RegistryAgentVersion = &v
	}
	if e.ConfigReleaseVersion > 0 {
		v := e.ConfigReleaseVersion
		req.ConfigReleaseVersion = &v
	}
	req.ResolutionSource = e.ResolutionSource
	req.ConfigContentHash = e.ConfigContentHash
	return req
}

func optionalInt(v int) *int {
	if v <= 0 {
		return nil
	}
	return &v
}

func optionalCostUSD(v float64) *float64 {
	if v <= 0 {
		return nil
	}
	return &v
}
