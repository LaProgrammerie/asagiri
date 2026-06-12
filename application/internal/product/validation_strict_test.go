package product

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/product/derivation"
	"github.com/stretchr/testify/require"
)

func TestValidateProjectionCouplingStrictSuccess(t *testing.T) {
	p := derivation.Projection{
		Analytics: []string{
			"metric:onboarding_completion_rate",
			"dashboard:onboarding_completion_rate",
		},
		Observability: []string{
			"metric:onboarding_completion_rate",
		},
		MetricsCoverage: []string{
			"coupled:onboarding:onboarding_completion_rate",
		},
	}
	require.NoError(t, ValidateProjectionCouplingStrict(p))
}

func TestValidateProjectionCouplingStrictFailsOnMissingMetric(t *testing.T) {
	p := derivation.Projection{
		Analytics: []string{
			"metric:onboarding_completion_rate",
		},
		Observability: []string{
			"metric:onboarding_completion_rate",
		},
		MetricsCoverage: []string{
			"missing_metrics:onboarding",
		},
	}
	require.Error(t, ValidateProjectionCouplingStrict(p))
}
