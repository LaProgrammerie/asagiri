package coordination

import "context"

// EscalationTarget names a role to invoke after failures (spec-my-D §14).
type EscalationTarget string

const (
	EscalationInvestigator        EscalationTarget = "investigator"
	EscalationArchitectureReview  EscalationTarget = "architecture_review"
)

// EscalationHandler applies retry and escalation rules (Lot 4 stub).
type EscalationHandler interface {
	ShouldRetry(ctx context.Context, nodeID string, attempt int, maxAttempts int) bool
	Escalate(ctx context.Context, graphID, nodeID string, failureCount int) (EscalationTarget, error)
}

// StubEscalationHandler is a Lot-1 placeholder.
type StubEscalationHandler struct{}

func (StubEscalationHandler) ShouldRetry(_ context.Context, _ string, attempt, maxAttempts int) bool {
	if maxAttempts <= 0 {
		return false
	}
	return attempt < maxAttempts
}

func (StubEscalationHandler) Escalate(_ context.Context, _, _ string, _ int) (EscalationTarget, error) {
	return "", ErrNotImplemented
}
