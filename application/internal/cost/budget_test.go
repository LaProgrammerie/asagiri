package cost

import (
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

func TestCheckBudgetConfirm(t *testing.T) {
	cfg := config.NewTestConfig("x")
	cfg.Budgets.PerRun.RequireConfirmationAboveCost = 0.05
	cfg.Budgets.Policies.BlockWhenOverBudget = true
	est := ExecutionEstimate{
		EstimatedCost: Money{Cents: 500, Currency: "EUR"},
		TotalTokens:   1000,
	}
	r, err := CheckBudget(est, cfg, BudgetOverrides{InteractiveAllow: true})
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != BudgetConfirm {
		t.Fatalf("got %s", r.Status)
	}
}

func TestCheckBudgetBlockOverRunCap(t *testing.T) {
	cfg := config.NewTestConfig("x")
	cfg.Budgets.PerRun.MaxEstimatedCost = 1.0
	cfg.Budgets.Policies.BlockWhenOverBudget = true
	est := ExecutionEstimate{
		EstimatedCost: Money{Cents: 50000, Currency: "EUR"},
		TotalTokens:   100,
	}
	r, err := CheckBudget(est, cfg, BudgetOverrides{})
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != BudgetBlock {
		t.Fatalf("got %s", r.Status)
	}
}
