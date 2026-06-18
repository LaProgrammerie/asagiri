package workflow

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestDevOneTaskFailedWithoutForceBlocked(t *testing.T) {
	svc, store := devEnrichGateService(t, false)
	task := seedTask(t, store, "feat", "task-failed", asagiri.StatusFailed)
	run := &sqlite.Run{ID: "run-failed", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "transition d'état invalide")

	fresh, err := store.GetTask("task-failed")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusFailed, fresh.Status)
}

func TestDevOneTaskFailedWithForceRetries(t *testing.T) {
	svc, store := devEnrichGateService(t, false)
	task := seedTask(t, store, "feat", "task-failed-retry", asagiri.StatusFailed)
	run := &sqlite.Run{ID: "run-failed-retry", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-failed-retry")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}
