package cost

import (
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// BudgetOverrides are CLI/runtime overrides for budget gates.
type BudgetOverrides struct {
	MaxCostMajor     float64 // optional cap from --budget (major units, e.g. EUR)
	AllowOverBudget  bool
	InteractiveAllow bool // if false and status would be CONFIRM, treat as BLOCK
}

// BudgetCheckResult is the outcome of CheckBudget.
type BudgetCheckResult struct {
	Status BudgetStatus
	Reason string
}

// BudgetPendingConfirmError signals interactive confirmation is required before spend.
type BudgetPendingConfirmError struct {
	Reason string
}

func (e *BudgetPendingConfirmError) Error() string {
	if e == nil {
		return ""
	}
	return "budget: confirmation requise — " + e.Reason
}

// majorFromMoney converts minor units to major assuming 1 major = 100 minor.
func majorFromMoney(m Money) float64 {
	return float64(m.Cents) / 100
}

// CheckBudget enforces per-task / per-run ceilings and confirmation bands (specv3 §3.2).
func CheckBudget(est ExecutionEstimate, cfg *config.Config, o BudgetOverrides) (BudgetCheckResult, error) {
	if cfg == nil {
		return BudgetCheckResult{}, fmt.Errorf("budget: config nil")
	}
	spend := majorFromMoney(est.EstimatedCost)
	maxTokTask := cfg.Budgets.PerTask.MaxEstimatedTokens
	maxTokRun := cfg.Budgets.PerRun.MaxEstimatedTokens
	maxCostTask := cfg.Budgets.PerTask.MaxEstimatedCost
	maxCostRun := cfg.Budgets.PerRun.MaxEstimatedCost
	confirmAbove := cfg.Budgets.PerRun.RequireConfirmationAboveCost

	if o.MaxCostMajor > 0 && spend > o.MaxCostMajor {
		if o.AllowOverBudget {
			return BudgetCheckResult{Status: BudgetOK, Reason: "under CLI budget via allow-over-budget"}, nil
		}
		return BudgetCheckResult{
			Status: BudgetBlock,
			Reason: fmt.Sprintf("estimation %.2f dépasse le plafond CLI %.2f", spend, o.MaxCostMajor),
		}, nil
	}

	if maxTokTask > 0 && est.TotalTokens > maxTokTask {
		if cfg.Budgets.Policies.BlockWhenOverBudget && !o.AllowOverBudget {
			return BudgetCheckResult{Status: BudgetBlock, Reason: "tokens estimés au-delà du budget tâche"}, nil
		}
	}
	if maxTokRun > 0 && est.TotalTokens > maxTokRun {
		if cfg.Budgets.Policies.BlockWhenOverBudget && !o.AllowOverBudget {
			return BudgetCheckResult{Status: BudgetBlock, Reason: "tokens estimés au-delà du budget run"}, nil
		}
	}
	if maxCostTask > 0 && spend > maxCostTask {
		if cfg.Budgets.Policies.BlockWhenOverBudget && !o.AllowOverBudget {
			return BudgetCheckResult{Status: BudgetBlock, Reason: "coût estimé au-delà du budget tâche"}, nil
		}
	}
	if maxCostRun > 0 && spend > maxCostRun {
		if cfg.Budgets.Policies.BlockWhenOverBudget && !o.AllowOverBudget {
			return BudgetCheckResult{Status: BudgetBlock, Reason: "coût estimé au-delà du budget run"}, nil
		}
	}

	if confirmAbove > 0 && spend >= confirmAbove {
		if o.AllowOverBudget {
			return BudgetCheckResult{Status: BudgetOK, Reason: "confirmation contournée"}, nil
		}
		if !o.InteractiveAllow {
			return BudgetCheckResult{Status: BudgetBlock, Reason: "confirmation requise (hors interactif)"}, nil
		}
		return BudgetCheckResult{Status: BudgetConfirm, Reason: fmt.Sprintf("coût estimé %.2f au-dessus du seuil %.2f", spend, confirmAbove)}, nil
	}

	return BudgetCheckResult{Status: BudgetOK, Reason: ""}, nil
}
