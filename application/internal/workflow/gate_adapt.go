package workflow

import (
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

func gateHistoryEntryFromResult(gateName string, r gates.Result, at string, retry int) asagiri.GateHistoryEntry {
	return asagiri.GateHistoryEntry{
		Gate:       gateName,
		At:         at,
		Status:     string(r.Status),
		Confidence: r.Confidence,
		Notes:      r.Notes,
		Findings:   gateFindingsFromGates(r.Findings),
		Retry:      retry,
		DryRun:     r.DryRun,
		ParseError: r.ParseError,
	}
}

func governanceRecordFromEntry(e asagiri.GateHistoryEntry) asagiri.GovernanceRecord {
	return asagiri.GovernanceRecord{
		At:         e.At,
		Status:     e.Status,
		Confidence: e.Confidence,
		Notes:      e.Notes,
		Findings:   e.Findings,
		Retry:      e.Retry,
		DryRun:     e.DryRun,
		ParseError: e.ParseError,
	}
}

func governanceRecordFromResult(r gates.Result, at string, retry int) asagiri.GovernanceRecord {
	return governanceRecordFromEntry(gateHistoryEntryFromResult(governanceGateName, r, at, retry))
}

func gateFindingsFromGates(findings []gates.Finding) []asagiri.GateFinding {
	if len(findings) == 0 {
		return nil
	}
	out := make([]asagiri.GateFinding, len(findings))
	for i, f := range findings {
		out[i] = asagiri.GateFinding{
			Code:     f.Code,
			Severity: f.Severity,
			Message:  f.Message,
			Actions:  f.Actions,
		}
	}
	return out
}
