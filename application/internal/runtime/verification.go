package runtime

// Verification runtime event types (spec-my-B §18).
const (
	EventVerificationStarted            = "verification.started"
	EventVerificationCompleted          = "verification.completed"
	EventTrustLowConfidence             = "trust.low_confidence"
	EventSecurityIssueDetected          = "security.issue_detected"
	EventFlowIntegrityFailed            = "flow.integrity_failed"
	EventContractBreakingChangeDetected = "contract.breaking_change_detected"
)

// VerificationEmitter publishes trust verification events on the runtime bus.
type VerificationEmitter struct {
	Store *Store
}

// Emit records a verification-related runtime event (source: trust).
func (e *VerificationEmitter) Emit(eventType, flowID string, payload map[string]any) error {
	if e == nil || e.Store == nil {
		return nil
	}
	_, err := e.Store.EmitEvent(eventType, "trust", "", flowID, payload)
	return err
}
