package trust

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

func TestReportBlastRadiusSection(t *testing.T) {
	br := &BlastRadiusReport{
		FlowsImpacted:      4,
		CriticalAPIs:       2,
		SharedModules:      6,
		MigrationRisk:      "medium",
		PublicContractRisk: "high",
	}
	report := NewTrustReport(VerificationScope{TrustID: "trust-1"}, []VerificationCheck{
		{ID: "1", Name: "blast", Type: CheckBlastRadius, Status: CheckStatusPassed},
	}, confidence.Report{Overall: 0.8}, GateEvaluation{}, br, nil)
	md := toTrustMarkdown(report)
	require.Contains(t, md, "## Blast Radius")
	require.Contains(t, md, "Flows impacted: 4")
	require.Contains(t, md, "Public contract risk: high")

	term := FormatTerminalSummary(report)
	require.Contains(t, term, "Blast Radius")
	require.Contains(t, term, "Critical APIs: 2")
}

func TestBlastRadiusFromChecks(t *testing.T) {
	raw := []checks.Check{{
		Type: "blast-radius",
		BlastRadius: &checks.BlastRadiusSummary{
			FlowsImpacted: 2,
			CriticalAPIs:  1,
		},
	}}
	br := blastRadiusFromChecks(raw)
	require.NotNil(t, br)
	require.Equal(t, 2, br.FlowsImpacted)
}
