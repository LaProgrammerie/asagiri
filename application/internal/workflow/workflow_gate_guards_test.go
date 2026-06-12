package workflow

import (
	"context"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestVerifyFeatureBlockedWhenHumanReviewPending(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	svc.dryRun = true
	seedTask(t, store, "feat", "task-hr-verify", asagiri.StatusImplemented)

	_, err := svc.VerifyFeature(context.Background(), "feat", "task-hr-verify", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Gate human_review requires action")
	require.Contains(t, err.Error(), "asa gates submit human_review --task task-hr-verify")
	require.Contains(t, err.Error(), "asa continue --yes")

	fresh, err := store.GetTask("task-hr-verify")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
}

func TestVerifyFeatureAllowedWhenHumanReviewSatisfied(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	svc.dryRun = true
	task := seedTaskWithGateHistory(t, store, "feat", "task-hr-ok", asagiri.StatusImplemented, gates.HumanReviewGateName, "pass")

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
}

func TestReviewFeatureBlockedWhenHumanReviewPendingOnVerified(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	svc.dryRun = true
	seedTask(t, store, "feat", "task-hr-review", asagiri.StatusVerified)

	_, err := svc.ReviewFeature(context.Background(), "feat", "task-hr-review", "reviewer", true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "Gate human_review requires action")
	require.Contains(t, err.Error(), "asa gates submit human_review --task task-hr-review")

	fresh, err := store.GetTask("task-hr-review")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
}

func TestReviewFeatureAllowedWhenHumanReviewSatisfied(t *testing.T) {
	svc, store := humanReviewTestService(t, true, false)
	svc.dryRun = true
	task := seedTaskWithGateHistory(t, store, "feat", "task-hr-review-ok", asagiri.StatusVerified, gates.HumanReviewGateName, "pass")

	_, err := svc.ReviewFeature(context.Background(), "feat", task.ID, "reviewer", true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusReviewed, fresh.Status)
}

func seedTaskWithGateHistory(t *testing.T, store *sqlite.Store, feature, id, status, gateName, gateStatus string) sqlite.Task {
	t.Helper()
	task := seedTask(t, store, feature, id, status)
	canonical, err := payloadToCanonical(task.PayloadJSON)
	require.NoError(t, err)
	if canonical.Gates == nil {
		canonical.Gates = &asagiri.TaskGates{}
	}
	canonical.Gates.History = append(canonical.Gates.History, asagiri.GateHistoryEntry{
		Gate:   gateName,
		Status: gateStatus,
		At:     time.Now().UTC().Format(time.RFC3339),
	})
	payload, err := canonicalToPayload(canonical)
	require.NoError(t, err)
	require.NoError(t, store.UpdateTask(&sqlite.Task{ID: id, PayloadJSON: payload}))
	task.PayloadJSON = payload
	return task
}
