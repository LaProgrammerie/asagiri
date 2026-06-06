package workflow

import (
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

func governanceRecordFromResult(r gates.Result, at string, retry int) asagiri.GovernanceRecord {
	return asagiri.GovernanceRecord{
		At:         at,
		Status:     string(r.Status),
		Confidence: r.Confidence,
		Notes:      r.Notes,
		Findings:   governanceFindingsToAsagiri(r.Findings),
		Retry:      retry,
		DryRun:     r.DryRun,
		ParseError: r.ParseError,
	}
}

func governanceFindingsToAsagiri(findings []gates.Finding) []asagiri.GovernanceFinding {
	if len(findings) == 0 {
		return nil
	}
	out := make([]asagiri.GovernanceFinding, len(findings))
	for i, f := range findings {
		out[i] = asagiri.GovernanceFinding{
			Code:     f.Code,
			Severity: f.Severity,
			Message:  f.Message,
			Actions:  f.Actions,
		}
	}
	return out
}
