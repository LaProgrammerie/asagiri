package checks

import (
	"context"
	"time"
)

// Runner executes one verification check type (spec §8).
type Runner interface {
	Type() string
	Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error)
}

// Finding is a check-local finding before mapping to trust.Finding.
type Finding struct {
	Severity     string
	Category     string
	Message      string
	SuggestedFix string
}

// Evidence links a check to verifiable artefacts.
type Evidence struct {
	Kind    string
	Source  string
	Summary string
}

// CheckResult is the full runner output before registry mapping.
type CheckResult struct {
	ID          string
	Name        string
	Type        string
	Status      string
	Confidence  float64
	Findings    []Finding
	Evidence    []Evidence
	Duration    time.Duration
	BlastRadius *BlastRadiusSummary
}
