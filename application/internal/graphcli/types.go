package graphcli

import "github.com/LaProgrammerie/asagiri/application/internal/executiongraph"

// PlanGraphResult is the structured JSON output for `asa plan graph --json`.
type PlanGraphResult struct {
	Graph     executiongraph.ExecutionGraph    `json:"graph"`
	Schedule  executiongraph.ExecutionSchedule `json:"schedule"`
	Estimate  executiongraph.GraphEstimate     `json:"estimate"`
	Artifacts executiongraph.GraphArtifacts    `json:"artifacts"`
}

// GraphRunJSONResult is the structured JSON output for `asa graph run --json`.
type GraphRunJSONResult struct {
	Graph     executiongraph.ExecutionGraph    `json:"graph"`
	Schedule  executiongraph.ExecutionSchedule `json:"schedule"`
	Estimate  executiongraph.GraphEstimate     `json:"estimate"`
	Artifacts executiongraph.GraphArtifacts    `json:"artifacts"`
	Result    executiongraph.GraphRunResult    `json:"result"`
	DryRun    bool                             `json:"dry_run"`
}

// GraphStatusResult is the structured JSON output for `asa graph status --json`.
type GraphStatusResult struct {
	Graph    executiongraph.ExecutionGraph `json:"graph"`
	Estimate executiongraph.GraphEstimate  `json:"estimate"`
}

// GraphResumeResult is the structured JSON output for `asa graph resume --json`.
type GraphResumeResult struct {
	Result executiongraph.GraphRunResult `json:"result"`
}
