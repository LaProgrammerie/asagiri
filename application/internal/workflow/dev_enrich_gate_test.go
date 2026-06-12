package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func seedTaskWithEnrichGateHistory(t *testing.T, store *sqlite.Store, feature, id, status, gateStatus string) sqlite.Task {
	t.Helper()
	task := seedTask(t, store, feature, id, status)
	if gateStatus == "" {
		return task
	}
	canonical, err := payloadToCanonical(task.PayloadJSON)
	require.NoError(t, err)
	if canonical.Gates == nil {
		canonical.Gates = &asagiri.TaskGates{}
	}
	canonical.Gates.History = append(canonical.Gates.History, asagiri.GateHistoryEntry{
		Gate:   enrichGateName,
		Status: gateStatus,
		At:     time.Now().UTC().Format(time.RFC3339),
	})
	payload, err := canonicalToPayload(canonical)
	require.NoError(t, err)
	require.NoError(t, store.UpdateTask(&sqlite.Task{ID: id, PayloadJSON: payload}))
	task.PayloadJSON = payload
	return task
}

func devEnrichGateService(t *testing.T, gateEnabled bool) (*Service, *sqlite.Store) {
	t.Helper()
	svc, store := newEnrichGateService(t, gateEnabled, nil)
	svc.cfg.Agents["dev"] = config.Agent{Command: "echo", Args: []string{"ok"}}
	enableWorktreeDryRun(svc)
	svc.dryRun = true
	return svc, store
}

func TestDevOneTaskEnrichGateEnrichedWithoutHistoryBlocked(t *testing.T) {
	svc, store := devEnrichGateService(t, true)
	task := seedTask(t, store, "feat", "task-enriched-bypass", asagiri.StatusEnriched)
	run := &sqlite.Run{ID: "run-enriched-bypass", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate required before dev")
	require.Contains(t, err.Error(), "asa enrich feat --task task-enriched-bypass --force")

	fresh, err := store.GetTask("task-enriched-bypass")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusEnriched, fresh.Status)
}

func TestDevOneTaskEnrichGateEnrichedWithPassHistoryAllowed(t *testing.T) {
	svc, store := devEnrichGateService(t, true)
	task := seedTaskWithEnrichGateHistory(t, store, "feat", "task-enriched-pass", asagiri.StatusEnriched, "pass")
	run := &sqlite.Run{ID: "run-enriched-pass", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-enriched-pass")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}

func TestDevOneTaskEnrichGateActiveNoHistoryBlocked(t *testing.T) {
	svc, store := devEnrichGateService(t, true)
	task := seedTask(t, store, "feat", "task-block", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-block", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate required before dev")
	require.Contains(t, err.Error(), "asa enrich feat --task task-block --force")

	fresh, err := store.GetTask("task-block")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusPlanned, fresh.Status)
}

func TestDevOneTaskEnrichGatePassHistoryAllowed(t *testing.T) {
	svc, store := devEnrichGateService(t, true)
	task := seedTaskWithEnrichGateHistory(t, store, "feat", "task-pass", asagiri.StatusPlanned, "pass")
	run := &sqlite.Run{ID: "run-pass", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-pass")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}

func TestDevOneTaskEnrichGateInactivePlannedPromoted(t *testing.T) {
	svc, store := devEnrichGateService(t, false)
	task := seedTask(t, store, "feat", "task-legacy", asagiri.StatusPlanned)
	run := &sqlite.Run{ID: "run-legacy", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-legacy")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}

func TestDevOneTaskEnrichGateWarnNonAdvisoryBlocked(t *testing.T) {
	f := false
	svc, store := newEnrichGateService(t, true, &f)
	svc.cfg.Agents["dev"] = config.Agent{Command: "echo", Args: []string{"ok"}}
	enableWorktreeDryRun(svc)
	svc.dryRun = true

	task := seedTaskWithEnrichGateHistory(t, store, "feat", "task-warn", asagiri.StatusPlanned, "warn")
	run := &sqlite.Run{ID: "run-warn", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.Error(t, err)

	fresh, err := store.GetTask("task-warn")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusPlanned, fresh.Status)
}

func TestDevOneTaskEnrichGateWarnAdvisoryAllowed(t *testing.T) {
	svc, store := devEnrichGateService(t, true)
	task := seedTaskWithEnrichGateHistory(t, store, "feat", "task-warn-ok", asagiri.StatusPlanned, "warn")
	run := &sqlite.Run{ID: "run-warn-ok", Feature: "feat"}
	a, err := svc.ensureAgent("dev")
	require.NoError(t, err)

	_, err = svc.devOneTask(context.Background(), run, "feat", task, a, true)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-warn-ok")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}
