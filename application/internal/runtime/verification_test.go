package runtime

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVerificationEmitterEmit(t *testing.T) {
	dir := t.TempDir()
	store, err := Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	emitter := &VerificationEmitter{Store: store}
	require.NoError(t, emitter.Emit(EventVerificationStarted, "onboarding", map[string]any{"trust_id": "trust-1"}))

	events, err := store.ListEvents(10)
	require.NoError(t, err)
	require.NotEmpty(t, events)
	require.Equal(t, EventVerificationStarted, events[0].Type)
	require.Equal(t, "trust", events[0].Source)
	require.Equal(t, "onboarding", events[0].FlowID)
}
