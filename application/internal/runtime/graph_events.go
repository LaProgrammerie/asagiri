package runtime

// Graph runtime event types (spec-my-C §19).
const (
	EventGraphCreated           = "graph.created"
	EventGraphStarted           = "graph.started"
	EventGraphNodeStarted       = "graph.node.started"
	EventGraphNodeCompleted     = "graph.node.completed"
	EventGraphNodeFailed        = "graph.node.failed"
	EventGraphCheckpointCreated = "graph.checkpoint.created"
	EventGraphBlocked           = "graph.blocked"
	EventGraphCompleted         = "graph.completed"
)

// GraphEmitter publishes execution graph events on the runtime bus.
type GraphEmitter struct {
	Store *Store
}

// Emit records a graph-related runtime event (source: executiongraph).
func (e *GraphEmitter) Emit(eventType, graphID, flowID string, payload map[string]any) error {
	if e == nil || e.Store == nil {
		return nil
	}
	if payload == nil {
		payload = make(map[string]any)
	}
	payload["graph_id"] = graphID
	_, err := e.Store.EmitEvent(eventType, "executiongraph", graphID, flowID, payload)
	return err
}
