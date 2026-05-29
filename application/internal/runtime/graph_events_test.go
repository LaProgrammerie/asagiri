package runtime_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

func TestGraphEmitterEmitsAllEventTypes(t *testing.T) {
	repo := t.TempDir()
	store, err := runtime.Open(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	emitter := &runtime.GraphEmitter{Store: store}
	graphID := "graph-test-001"
	flowID := "onboarding"

	events := []string{
		runtime.EventGraphCreated,
		runtime.EventGraphStarted,
		runtime.EventGraphNodeStarted,
		runtime.EventGraphNodeCompleted,
		runtime.EventGraphNodeFailed,
		runtime.EventGraphCheckpointCreated,
		runtime.EventGraphBlocked,
		runtime.EventGraphCompleted,
	}
	for _, eventType := range events {
		require.NoError(t, emitter.Emit(eventType, graphID, flowID, map[string]any{
			"node_id": "node-a",
		}))
	}

	listed, err := store.ListEvents(20)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(listed), len(events))

	seen := make(map[string]struct{})
	for _, ev := range listed {
		if ev.Payload["graph_id"] == graphID {
			seen[ev.Type] = struct{}{}
		}
	}
	for _, eventType := range events {
		require.Contains(t, seen, eventType)
	}
}

func TestGraphEmitterNilStoreNoOp(t *testing.T) {
	var emitter runtime.GraphEmitter
	require.NoError(t, emitter.Emit(runtime.EventGraphCreated, "g1", "f1", nil))
}
