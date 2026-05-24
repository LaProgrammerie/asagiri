package cost

import "time"

// Money is a minor-unit amount (e.g. cents) with ISO currency tag.
type Money struct {
	Cents    int64
	Currency string
}

// ExecutionEstimate captures pre-run token/cost/time projection (specv3 §3.3).
type ExecutionEstimate struct {
	RunID             string
	Feature           string
	TaskID            string
	PlannedSteps      []EstimatedStep
	TotalInputTokens  int
	TotalOutputTokens int
	TotalTokens       int
	EstimatedCost     Money
	EstimatedDuration time.Duration
	Confidence        float64
	BudgetStatus      BudgetStatus
	Risk              string
	Warnings          []string
}

// EstimatedStep is one projected model or local step.
type EstimatedStep struct {
	Name              string
	Agent             string
	Model             string
	Local             bool
	InputTokens       int
	OutputTokens      int
	EstimatedCost     Money
	EstimatedDuration time.Duration
	Reason            string
}

// BudgetStatus is a coarse gated outcome before spend.
type BudgetStatus string

const (
	BudgetOK      BudgetStatus = "OK"
	BudgetConfirm BudgetStatus = "CONFIRM"
	BudgetBlock   BudgetStatus = "BLOCK"
)

// PlannedStep is input to duration modeling (specv3 §7).
type PlannedStep struct {
	Name        string
	Agent       string
	Model       string
	Local       bool
	InputTokens int
	TaskType    string
}
