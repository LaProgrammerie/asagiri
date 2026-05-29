package executiongraph

// CIShouldFailPlan reports whether CI mode should exit non-zero after planning.
func CIShouldFailPlan(est GraphEstimate, _ RiskLevel) bool {
	return est.BudgetStatus == budgetStatusExceeded
}

// CIShouldFailRun reports whether CI mode should exit non-zero before or after dry-run execution.
func CIShouldFailRun(graph ExecutionGraph, est GraphEstimate) bool {
	stopOn := graph.Strategy.StopOnRisk
	if stopOn == "" {
		return false
	}
	if riskRank(est.HighestRisk) >= riskRank(stopOn) {
		return true
	}
	if est.BudgetStatus == budgetStatusExceeded {
		return true
	}
	return false
}
