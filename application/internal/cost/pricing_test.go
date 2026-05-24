package cost

import (
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

func TestCostFromPricing(t *testing.T) {
	cfg := config.NewTestConfig("x")
	cfg.Pricing.Models["m1"] = config.ModelPricing{
		InputPer1MTokens: 1.0, OutputPer1MTokens: 2.0,
	}
	m, err := CostFromPricing(cfg, "m1", 1_000_000, 500_000)
	if err != nil {
		t.Fatal(err)
	}
	if m.Cents != 200 {
		t.Fatalf("expected 200 cents, got %d", m.Cents)
	}
	if _, err := CostFromPricing(cfg, "missing", 1, 1); err == nil {
		t.Fatal("expected error for unknown model")
	}
}
