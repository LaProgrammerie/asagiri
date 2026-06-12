package gates

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func trustGateActiveCfg(warnAdvisory *bool) *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				Trust: config.WorkTrustGateConfig{
					Enabled:        true,
					Mode:           config.GovernanceModePerTask,
					BlockVerdicts:  config.DefaultTrustGateBlockVerdicts(),
					WarnVerdicts:   config.DefaultTrustGateWarnVerdicts(),
					WarnIsAdvisory: warnAdvisory,
				},
			},
		},
	}
}

func TestTrustGateSatisfiedInactive(t *testing.T) {
	cfg := &config.Config{}
	require.True(t, TrustGateSatisfied(cfg, `{}`))
}

func TestTrustGateBlocksReviewWhenMissingEntry(t *testing.T) {
	cfg := trustGateActiveCfg(nil)
	require.True(t, TrustGateBlocksReview(cfg, "verified", `{}`))
}

func TestTrustGateAllowsReviewWhenPass(t *testing.T) {
	cfg := trustGateActiveCfg(nil)
	payload := `{"gates":{"history":[{"gate":"trust","status":"pass","at":"2026-06-08T12:00:00Z"}]}}`
	require.False(t, TrustGateBlocksReview(cfg, "verified", payload))
}

func TestTrustGateBlocksReviewOnWarnNonAdvisory(t *testing.T) {
	f := false
	cfg := trustGateActiveCfg(&f)
	payload := `{"gates":{"history":[{"gate":"trust","status":"warn","at":"2026-06-08T12:00:00Z"}]}}`
	require.True(t, TrustGateBlocksReview(cfg, "verified", payload))
}

func TestTrustGateAllowsReviewOnWarnAdvisory(t *testing.T) {
	cfg := trustGateActiveCfg(nil)
	payload := `{"gates":{"history":[{"gate":"trust","status":"warn","at":"2026-06-08T12:00:00Z"}]}}`
	require.False(t, TrustGateBlocksReview(cfg, "verified", payload))
}
