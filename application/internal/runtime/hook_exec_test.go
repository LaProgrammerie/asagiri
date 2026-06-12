package runtime_test

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestExecuteHookCommandRejectsNonAsa(t *testing.T) {
	t.Parallel()
	err := runtime.ExecuteHookCommand(context.Background(), t.TempDir(), "rm -rf /")
	require.Error(t, err)
}

func TestEnqueueAndProcessHooksDry(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	store, err := runtime.Open(dir)
	require.NoError(t, err)
	t.Cleanup(func() { _ = store.Close() })

	require.NoError(t, store.EnqueueHook("test.event", "asa version"))
	jobs, err := store.DequeueHooks(5)
	require.NoError(t, err)
	require.Len(t, jobs, 1)
}
