package cost

import (
	"fmt"
	"math"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// CostFromPricing computes spend from configured per-1M token prices only (specv3 §6).
func CostFromPricing(cfg *config.Config, model string, inputTokens, outputTokens int) (Money, error) {
	if cfg == nil {
		return Money{}, fmt.Errorf("cost: config nil")
	}
	if model == "" {
		return Money{}, fmt.Errorf("cost: modèle requis")
	}
	p, ok := cfg.Pricing.Models[model]
	if !ok {
		return Money{}, fmt.Errorf("cost: modèle %q absent de pricing.models (prix non configurés)", model)
	}
	cur := cfg.Pricing.Currency
	if cur == "" {
		cur = cfg.Budgets.DefaultCurrency
	}
	major := float64(inputTokens)/1_000_000*p.InputPer1MTokens +
		float64(outputTokens)/1_000_000*p.OutputPer1MTokens
	cents := int64(math.Round(major * 100))
	return Money{Cents: cents, Currency: cur}, nil
}
