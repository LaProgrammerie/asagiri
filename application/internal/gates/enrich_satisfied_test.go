package gates

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func enrichGateActiveCfg(warnAdvisory *bool) *config.Config {
	return &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{
					Enabled:        true,
					Mode:           config.GovernanceModePerTask,
					WarnIsAdvisory: warnAdvisory,
				},
			},
		},
	}
}

func payloadWithEnrichGate(status string) string {
	return `{"gates":{"history":[{"gate":"enrich","status":"` + status + `","at":"2026-01-01T00:00:00Z"}]}}`
}

func TestEnrichGateSatisfiedInactiveAlwaysTrue(t *testing.T) {
	cfg := &config.Config{}
	require.True(t, EnrichGateSatisfied(cfg, `{}`))
	require.True(t, EnrichGateSatisfied(nil, `{}`))
}

func TestEnrichGateSatisfiedNoHistoryFalse(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.False(t, EnrichGateSatisfied(cfg, `{}`))
}

func TestEnrichGateSatisfiedPass(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.True(t, EnrichGateSatisfied(cfg, payloadWithEnrichGate("pass")))
}

func TestEnrichGateSatisfiedWarnAdvisory(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.True(t, EnrichGateSatisfied(cfg, payloadWithEnrichGate("warn")))
}

func TestEnrichGateSatisfiedWarnNonAdvisory(t *testing.T) {
	f := false
	cfg := enrichGateActiveCfg(&f)
	require.False(t, EnrichGateSatisfied(cfg, payloadWithEnrichGate("warn")))
}

func TestEnrichGateSatisfiedFail(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.False(t, EnrichGateSatisfied(cfg, payloadWithEnrichGate("fail")))
}

func TestEnrichGateBlocksDevPlannedWithoutHistory(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.True(t, EnrichGateBlocksDev(cfg, asagiri.StatusPlanned, `{}`))
}

func TestEnrichGateBlocksDevEnrichedWithoutHistoryBlocks(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.True(t, EnrichGateBlocksDev(cfg, asagiri.StatusEnriched, `{}`))
}

func TestEnrichGateBlocksDevEnrichedWithPassHistoryAllowed(t *testing.T) {
	cfg := enrichGateActiveCfg(nil)
	require.False(t, EnrichGateBlocksDev(cfg, asagiri.StatusEnriched, payloadWithEnrichGate("pass")))
}

func TestEnrichGateBlocksDevInactiveNeverBlocks(t *testing.T) {
	cfg := &config.Config{}
	require.False(t, EnrichGateBlocksDev(cfg, asagiri.StatusPlanned, `{}`))
}
