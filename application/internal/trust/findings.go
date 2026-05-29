package trust

import "time"

// CheckType identifies a verification check category (spec §8).
type CheckType string

const (
	CheckStaticAnalysis        CheckType = "static-analysis"
	CheckContracts             CheckType = "contracts"
	CheckFlows                 CheckType = "flows"
	CheckPermissions           CheckType = "permissions"
	CheckObservability         CheckType = "observability"
	CheckAnalytics             CheckType = "analytics"
	CheckArchitecture          CheckType = "architecture"
	CheckSecurity              CheckType = "security"
	CheckPerformance           CheckType = "performance"
	CheckCost                  CheckType = "cost"
	CheckBackwardCompatibility CheckType = "backward-compatibility"
	CheckMigrationSafety       CheckType = "migration-safety"
	CheckBlastRadius           CheckType = "blast-radius"
)

// CheckStatus is the outcome of a single verification check.
type CheckStatus string

const (
	CheckStatusPassed  CheckStatus = "passed"
	CheckStatusFailed  CheckStatus = "failed"
	CheckStatusSkipped CheckStatus = "skipped"
	CheckStatusWarn    CheckStatus = "warn"
)

// Severity classifies a finding.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityError    Severity = "error"
	SeverityCritical Severity = "critical"
)

// Evidence links a finding or check to verifiable artefacts.
type Evidence struct {
	Kind    string `json:"kind"`
	Source  string `json:"source,omitempty"`
	Summary string `json:"summary,omitempty"`
}

// VerificationCheck is one executed check in the pipeline (spec §9).
type VerificationCheck struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Type       CheckType     `json:"type"`
	Status     CheckStatus   `json:"status"`
	Confidence float64       `json:"confidence"`
	Findings   []Finding     `json:"findings,omitempty"`
	Evidence   []Evidence    `json:"evidence,omitempty"`
	Duration   time.Duration `json:"duration"`
}

// CheckConfidence implements confidence.ContributoryCheck.
func (c VerificationCheck) CheckConfidence() float64 {
	return c.Confidence
}

// Finding is a structured issue surfaced by a check (spec §10).
type Finding struct {
	Severity     Severity   `json:"severity"`
	Category     string     `json:"category"`
	Message      string     `json:"message"`
	Evidence     []Evidence `json:"evidence,omitempty"`
	SuggestedFix string     `json:"suggested_fix,omitempty"`
}
