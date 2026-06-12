package coordination

import (
	"context"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// EscalationTarget names a role to invoke after failures (spec-my-D §14).
type EscalationTarget string

const (
	EscalationInvestigator       EscalationTarget = "investigator"
	EscalationArchitectureReview EscalationTarget = "architecture_review"
)

// EscalationHandler applies retry and escalation rules.
type EscalationHandler interface {
	ShouldRetry(ctx context.Context, nodeID string, attempt int, maxAttempts int) bool
	Escalate(ctx context.Context, graphID, nodeID string, failureCount int) (EscalationTarget, error)
}

// ConfigEscalationHandler reads retry/escalation from coordination config.
type ConfigEscalationHandler struct {
	implMaxAttempts    int
	AfterFailure       EscalationTarget
	AfterSecondFailure EscalationTarget
}

// EscalationHandlerFromConfig builds a handler from config.
func EscalationHandlerFromConfig(cfg *config.Config) *ConfigEscalationHandler {
	h := &ConfigEscalationHandler{
		implMaxAttempts:    2,
		AfterFailure:       EscalationInvestigator,
		AfterSecondFailure: EscalationArchitectureReview,
	}
	if cfg == nil {
		return h
	}
	co := cfg.Coordination
	if co.Retry.Implementation.MaxAttempts > 0 {
		h.implMaxAttempts = co.Retry.Implementation.MaxAttempts
	}
	if t := parseEscalationTarget(co.Escalation.AfterFailure); t != "" {
		h.AfterFailure = t
	}
	if t := parseEscalationTarget(co.Escalation.AfterSecondFailure); t != "" {
		h.AfterSecondFailure = t
	}
	return h
}

// ShouldRetry returns true while attempt is below maxAttempts.
func (h *ConfigEscalationHandler) ShouldRetry(_ context.Context, _ string, attempt, maxAttempts int) bool {
	if maxAttempts <= 0 {
		maxAttempts = h.implementationMaxAttempts()
	}
	if maxAttempts <= 0 {
		return false
	}
	return attempt < maxAttempts
}

// Escalate picks the escalation target from failure count.
func (h *ConfigEscalationHandler) Escalate(_ context.Context, _, _ string, failureCount int) (EscalationTarget, error) {
	if failureCount >= 2 && h.AfterSecondFailure != "" {
		return h.AfterSecondFailure, nil
	}
	if failureCount >= 1 && h.AfterFailure != "" {
		return h.AfterFailure, nil
	}
	return EscalationInvestigator, nil
}

// ImplementationMaxAttempts exposes the configured retry cap.
func (h *ConfigEscalationHandler) ImplementationMaxAttempts() int {
	return h.implementationMaxAttempts()
}

func (h *ConfigEscalationHandler) implementationMaxAttempts() int {
	if h == nil || h.implMaxAttempts <= 0 {
		return 2
	}
	return h.implMaxAttempts
}

func parseEscalationTarget(raw string) EscalationTarget {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "investigator", "":
		return EscalationInvestigator
	case "architecture_review", "architect", "architecture":
		return EscalationArchitectureReview
	default:
		return EscalationTarget(strings.TrimSpace(raw))
	}
}
