package telemetry

import "time"

// StepMetric is one persisted step measurement (specv3 §14.1).
type StepMetric struct {
	ID                     string
	RunID                  string
	StepName               string
	Agent                  string
	Model                  string
	Local                  bool
	EstimatedInputTokens   int
	EstimatedOutputTokens  int
	ActualInputTokens      int
	ActualOutputTokens     int
	EstimatedCostCents     int64
	ActualCostCents        int64
	EstimatedDuration      time.Duration
	ActualDuration         time.Duration
	Status                 string
}

// RunMetric aggregates one workflow run (specv3 §14.1).
type RunMetric struct {
	RunID                  string
	Feature                string
	TaskID                 string
	StartedAt              time.Time
	FinishedAt             time.Time
	EstimatedInputTokens   int
	EstimatedOutputTokens  int
	ActualInputTokens      int
	ActualOutputTokens     int
	EstimatedCostCents     int64
	ActualCostCents        int64
	EstimatedDuration      time.Duration
	ActualDuration         time.Duration
	Status                 string
}

// DurationSample is one historical duration for duration modeling.
type DurationSample struct {
	StepName string
	Model    string
	Agent    string
	Local    bool
	Duration time.Duration
}

// StepTokenTotals aggregates token usage split by local/cloud steps for savings computation.
type StepTokenTotals struct {
	LocalInputTokens  int64
	LocalOutputTokens int64
	CloudInputTokens  int64
	CloudOutputTokens int64
}
