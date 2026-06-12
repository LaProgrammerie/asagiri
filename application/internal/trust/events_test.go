package trust

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust/checks"
)

func TestRuntimeEmitterVerificationEvents(t *testing.T) {
	repo := t.TempDir()
	store, err := runtime.Open(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	eng := NewEngineWithChecks(repo, checks.NewRegistry())
	eng.Emitter = NewRuntimeEmitter(store)
	eng.Config = &config.Config{
		Validation: config.ValidationConfig{
			Commands: []config.ValidationCommand{{Command: "go test ./..."}},
		},
	}

	_, err = eng.Verify(context.Background(), VerificationRequest{Flow: "f", Product: "demo"})
	require.NoError(t, err)

	events, err := store.ListEvents(20)
	require.NoError(t, err)
	types := make(map[string]bool)
	for _, ev := range events {
		types[ev.Type] = true
	}
	require.True(t, types[runtime.EventVerificationStarted])
	require.True(t, types[runtime.EventVerificationCompleted])
}
