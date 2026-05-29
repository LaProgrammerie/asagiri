package trust

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

func TestComputeResidualRiskMediumWarnings(t *testing.T) {
	checks := []VerificationCheck{{
		Status: CheckStatusPassed,
		Findings: []Finding{
			{Severity: SeverityWarning},
			{Severity: SeverityWarning},
		},
	}}
	conf := confidence.Report{Overall: 0.9}
	require.Equal(t, ResidualRiskMedium, ComputeResidualRisk(checks, conf))
}

func TestComputeResidualRiskMediumContractsWarn(t *testing.T) {
	checks := []VerificationCheck{{
		Type:   CheckContracts,
		Status: CheckStatusWarn,
	}}
	conf := confidence.Report{Overall: 0.9}
	require.Equal(t, ResidualRiskMedium, ComputeResidualRisk(checks, conf))
}

func TestComputeResidualRiskMediumLowOverall(t *testing.T) {
	checks := []VerificationCheck{{Status: CheckStatusPassed}}
	conf := confidence.Report{Overall: 0.65}
	require.Equal(t, ResidualRiskMedium, ComputeResidualRisk(checks, conf))
}
