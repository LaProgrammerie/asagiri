package trust

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/confidence"
)

func TestGateEvaluatorNotConfigured(t *testing.T) {
	var g GateEvaluator
	ev := g.Evaluate(context.Background(), confidence.Report{}, nil)
	require.Equal(t, GateStatusNotConfigured, ev.Status)
}

func TestGateEvaluatorBlocksLowConfidence(t *testing.T) {
	g := NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {
				MinConfidence: map[string]float64{"security": 0.85},
			},
		},
	})
	ev := g.Evaluate(context.Background(), confidence.Report{Security: 0.5}, nil)
	require.Equal(t, GateStatusBlocked, ev.Status)
	require.Contains(t, ev.Reason, "security confidence")
	require.Equal(t, "production", ev.Profile)
}

func TestGateEvaluatorBlocksMissingRequiredCheck(t *testing.T) {
	g := NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {RequiredChecks: []string{"contracts"}},
		},
	})
	ev := g.Evaluate(context.Background(), confidence.Report{}, nil)
	require.Equal(t, GateStatusBlocked, ev.Status)
	require.Contains(t, ev.Reason, `required check "contracts" was not executed`)
}

func TestGateEvaluatorPasses(t *testing.T) {
	g := NewGateEvaluator(&config.VerificationConfig{
		Gates: map[string]config.GateProfile{
			"production": {
				MinConfidence:  map[string]float64{"overall": 0.5},
				RequiredChecks: []string{"contracts"},
			},
		},
	})
	ev := g.Evaluate(context.Background(), confidence.Report{Overall: 0.9}, []VerificationCheck{
		{Type: CheckContracts, Status: CheckStatusPassed},
	})
	require.Equal(t, GateStatusPassed, ev.Status)
}

func TestCIShouldFail(t *testing.T) {
	require.True(t, CIShouldFail(TrustReport{Gate: GateEvaluation{Status: GateStatusBlocked}}, false))
	require.True(t, CIShouldFail(TrustReport{
		Checks: []VerificationCheck{{Status: CheckStatusFailed}},
	}, false))
	require.True(t, CIShouldFail(TrustReport{
		Checks: []VerificationCheck{{Status: CheckStatusWarn}},
	}, true))
	require.False(t, CIShouldFail(TrustReport{
		Checks: []VerificationCheck{{Status: CheckStatusWarn}},
	}, false))
}
