package coordination

import (
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// CoordinationEmitter publishes agent.* runtime events (spec-my-D §10).
type CoordinationEmitter struct {
	Store *runtime.Store
}

// EmitAgentStarted records agent.started for a node assignment.
func (e *CoordinationEmitter) EmitAgentStarted(graphID, flowID string, asg AgentAssignment) error {
	return e.emit(runtime.EventAgentStarted, graphID, flowID, asg)
}

// EmitAgentCompleted records agent.completed.
func (e *CoordinationEmitter) EmitAgentCompleted(graphID, flowID string, asg AgentAssignment) error {
	return e.emit(runtime.EventAgentCompleted, graphID, flowID, asg)
}

// EmitAgentFailed records agent.failed.
func (e *CoordinationEmitter) EmitAgentFailed(graphID, flowID string, asg AgentAssignment, reason string) error {
	payload := assignmentPayload(asg)
	payload["reason"] = reason
	return e.emitRaw(runtime.EventAgentFailed, graphID, flowID, payload)
}

// EmitAgentBlocked records agent.blocked.
func (e *CoordinationEmitter) EmitAgentBlocked(graphID, flowID string, asg AgentAssignment, reason string) error {
	payload := assignmentPayload(asg)
	payload["reason"] = reason
	return e.emitRaw(runtime.EventAgentBlocked, graphID, flowID, payload)
}

// EmitReviewRequested records agent.review_requested.
func (e *CoordinationEmitter) EmitReviewRequested(graphID, flowID string, asg AgentAssignment) error {
	return e.emit(runtime.EventAgentReviewRequested, graphID, flowID, asg)
}

// EmitReviewRejected records agent.review_rejected.
func (e *CoordinationEmitter) EmitReviewRejected(graphID, flowID string, asg AgentAssignment, reason string) error {
	payload := assignmentPayload(asg)
	payload["reason"] = reason
	return e.emitRaw(runtime.EventAgentReviewRejected, graphID, flowID, payload)
}

// EmitHandoffCreated records agent.handoff.created.
func (e *CoordinationEmitter) EmitHandoffCreated(graphID, flowID string, h Handoff) error {
	if e == nil || e.Store == nil {
		return nil
	}
	payload := map[string]any{
		"graph_id":   graphID,
		"handoff_id": h.ID,
		"from":       string(h.From),
		"to":         string(h.To),
		"confidence": h.Confidence,
	}
	_, err := e.Store.EmitEvent(runtime.EventAgentHandoffCreated, "coordination", graphID, flowID, payload)
	return err
}

// EmitContextReduced records agent.context_reduced.
func (e *CoordinationEmitter) EmitContextReduced(graphID, flowID, nodeID string, bytesBefore, bytesAfter int) error {
	if e == nil || e.Store == nil {
		return nil
	}
	payload := map[string]any{
		"graph_id":     graphID,
		"node_id":      nodeID,
		"bytes_before": bytesBefore,
		"bytes_after":  bytesAfter,
	}
	_, err := e.Store.EmitEvent(runtime.EventAgentContextReduced, "coordination", graphID, flowID, payload)
	return err
}

func (e *CoordinationEmitter) emit(eventType, graphID, flowID string, asg AgentAssignment) error {
	return e.emitRaw(eventType, graphID, flowID, assignmentPayload(asg))
}

func (e *CoordinationEmitter) emitRaw(eventType, graphID, flowID string, payload map[string]any) error {
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

func assignmentPayload(asg AgentAssignment) map[string]any {
	return map[string]any{
		"node_id":    asg.NodeID,
		"agent_ref":  asg.AgentRef,
		"role":       string(asg.Role),
		"isolation":  string(asg.Isolation),
		"profile_id": asg.ProfileID,
	}
}
