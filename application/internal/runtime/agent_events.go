package runtime

// Agent coordination runtime event types (spec-my-D §10).
const (
	EventAgentStarted         = "agent.started"
	EventAgentCompleted       = "agent.completed"
	EventAgentFailed          = "agent.failed"
	EventAgentBlocked         = "agent.blocked"
	EventAgentReviewRequested = "agent.review_requested"
	EventAgentReviewRejected  = "agent.review_rejected"
	EventAgentHandoffCreated  = "agent.handoff.created"
	EventAgentContextReduced  = "agent.context_reduced"
)

// AgentEmitter publishes agent.* events on the runtime bus.
type AgentEmitter struct {
	Store *Store
}

// Emit records an agent-related runtime event (source: coordination).
func (e *AgentEmitter) Emit(eventType, graphID, flowID string, payload map[string]any) error {
	if e == nil || e.Store == nil {
		return nil
	}
	if payload == nil {
		payload = make(map[string]any)
	}
	payload["graph_id"] = graphID
	_, err := e.Store.EmitEvent(eventType, "coordination", graphID, flowID, payload)
	return err
}
