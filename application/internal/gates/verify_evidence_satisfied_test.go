package gates

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func verifyEvidenceGateActiveCfg(warnAdvisory *bool) *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				VerifyEvidence: config.WorkVerifyEvidenceGateConfig{
					Enabled:        true,
					Mode:           config.GovernanceModePerTask,
					WarnIsAdvisory: warnAdvisory,
				},
			},
		},
	}
}

func payloadWithVerifyEvidenceGate(status string) string {
	return `{"gates":{"history":[{"gate":"verify_evidence","status":"` + status + `","at":"2026-01-01T00:00:00Z"}]}}`
}

func TestVerifyEvidenceGateSatisfiedInactiveAlwaysTrue(t *testing.T) {
	cfg := &config.Config{}
	require.True(t, VerifyEvidenceGateSatisfied(cfg, `{}`))
	require.True(t, VerifyEvidenceGateSatisfied(nil, `{}`))
}

func TestVerifyEvidenceGateSatisfiedNoHistoryFalse(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.False(t, VerifyEvidenceGateSatisfied(cfg, `{}`))
}

func TestVerifyEvidenceGateSatisfiedPass(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.True(t, VerifyEvidenceGateSatisfied(cfg, payloadWithVerifyEvidenceGate("pass")))
}

func TestVerifyEvidenceGateSatisfiedWarnAdvisory(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.True(t, VerifyEvidenceGateSatisfied(cfg, payloadWithVerifyEvidenceGate("warn")))
}

func TestVerifyEvidenceGateSatisfiedWarnNonAdvisory(t *testing.T) {
	f := false
	cfg := verifyEvidenceGateActiveCfg(&f)
	require.False(t, VerifyEvidenceGateSatisfied(cfg, payloadWithVerifyEvidenceGate("warn")))
}

func TestVerifyEvidenceGateSatisfiedFail(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.False(t, VerifyEvidenceGateSatisfied(cfg, payloadWithVerifyEvidenceGate("fail")))
}

func TestVerifyEvidenceGateBlocksReviewVerifiedWithoutHistory(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.True(t, VerifyEvidenceGateBlocksReview(cfg, asagiri.StatusVerified, `{}`))
}

func TestVerifyEvidenceGateBlocksReviewVerifiedWithPassHistoryAllowed(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.False(t, VerifyEvidenceGateBlocksReview(cfg, asagiri.StatusVerified, payloadWithVerifyEvidenceGate("pass")))
}

func TestVerifyEvidenceGateBlocksReviewInactiveNeverBlocks(t *testing.T) {
	cfg := &config.Config{}
	require.False(t, VerifyEvidenceGateBlocksReview(cfg, asagiri.StatusVerified, `{}`))
}

func TestVerifyEvidenceGateBlocksReviewImplementedNotBlocked(t *testing.T) {
	cfg := verifyEvidenceGateActiveCfg(nil)
	require.False(t, VerifyEvidenceGateBlocksReview(cfg, asagiri.StatusImplemented, `{}`))
}
