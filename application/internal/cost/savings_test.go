package cost_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
	"github.com/stretchr/testify/require"
)

func tokens(li, lo, ci, co int64) telemetry.StepTokenTotals {
	return telemetry.StepTokenTotals{
		LocalInputTokens:  li,
		LocalOutputTokens: lo,
		CloudInputTokens:  ci,
		CloudOutputTokens: co,
	}
}

func TestComputeSavings_NoPremiumReference(t *testing.T) {
	cfg := &config.Config{}
	cfg.Pricing.Currency = "EUR"
	// PremiumReferenceModel intentionally empty

	tok := tokens(900_000, 100_000, 100_000, 10_000)
	r := cost.ComputeSavings(tok, 50, cfg)

	require.False(t, r.HasPremiumBaseline(), "no baseline → HasPremiumBaseline must be false")
	require.Equal(t, int64(0), r.PremiumEquivCents, "no savings invented without reference model")
	require.Equal(t, int64(0), r.SavingsCents)
	require.Equal(t, 0.0, r.SavingsRate)
	require.Equal(t, "EUR", r.Currency)
	// Local/cloud split still available
	require.InDelta(t, 90.1, r.LocalPct(), 0.2)
}

func TestComputeSavings_WithPremiumReference(t *testing.T) {
	cfg := &config.Config{}
	cfg.Pricing.Currency = "EUR"
	cfg.Pricing.PremiumReferenceModel = "gpt-4o"
	cfg.Pricing.Models = map[string]config.ModelPricing{
		"gpt-4o": {InputPer1MTokens: 5.0, OutputPer1MTokens: 15.0},
	}

	// 1M local input + 100k local output + 100k cloud input + 10k cloud output
	// actual cost = 50 cents
	tok := tokens(1_000_000, 100_000, 100_000, 10_000)
	actualCents := int64(50)
	r := cost.ComputeSavings(tok, actualCents, cfg)

	require.True(t, r.HasPremiumBaseline())
	require.Equal(t, "gpt-4o", r.PremiumReferenceModel)

	// Premium equiv: (1.1M input × €5/1M) + (110k output × €15/1M)
	// = 5.50 + 1.65 = €7.15 = 715 cents
	require.Equal(t, int64(715), r.PremiumEquivCents)
	require.Equal(t, int64(715-50), r.SavingsCents)
	require.InDelta(t, (715.0-50.0)/715.0, r.SavingsRate, 0.001)
}

func TestComputeSavings_ReferenceModelNotInPricing(t *testing.T) {
	cfg := &config.Config{}
	cfg.Pricing.Currency = "EUR"
	cfg.Pricing.PremiumReferenceModel = "unknown-model"
	cfg.Pricing.Models = map[string]config.ModelPricing{}

	r := cost.ComputeSavings(tokens(500_000, 50_000, 50_000, 5_000), 10, cfg)

	require.False(t, r.HasPremiumBaseline(), "missing pricing entry → no savings")
	require.Equal(t, int64(0), r.PremiumEquivCents)
}

func TestComputeSavings_NilConfig(t *testing.T) {
	r := cost.ComputeSavings(tokens(100_000, 10_000, 0, 0), 0, nil)
	require.False(t, r.HasPremiumBaseline())
	require.Equal(t, "EUR", r.Currency) // default currency
}

func TestFormatCents(t *testing.T) {
	require.Equal(t, "€4.90", cost.FormatCents(490, "EUR"))
	require.Equal(t, "€0.00", cost.FormatCents(0, "EUR"))
	require.Equal(t, "-€1.50", cost.FormatCents(-150, "EUR"))
	require.Equal(t, "$12.34", cost.FormatCents(1234, "USD"))
}

func TestLocalPct_ZeroTokens(t *testing.T) {
	r := cost.ComputeSavings(tokens(0, 0, 0, 0), 0, nil)
	require.Equal(t, 0.0, r.LocalPct())
}
