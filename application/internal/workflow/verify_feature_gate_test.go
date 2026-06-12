package workflow

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func verifyFeatureGateService(t *testing.T, enabled bool, warnAdvisory *bool, dryRun bool) (*Service, *sqlite.Store) {
	t.Helper()
	svc, store := newVerifyEvidenceGateService(t, enabled, warnAdvisory)
	svc.dryRun = dryRun
	return svc, store
}

func verifyEvidenceGateHistoryCount(payloadJSON string) int {
	canonical, err := payloadToCanonical(payloadJSON)
	if err != nil || canonical.Gates == nil {
		return 0
	}
	n := 0
	for _, e := range canonical.Gates.History {
		if e.Gate == verifyEvidenceGateName {
			n++
		}
	}
	return n
}

func TestVerifyFeatureGateDisabledPreservesBehavior(t *testing.T) {
	svc, store := verifyFeatureGateService(t, false, nil, false)
	task := seedTaskWithValidationCommands(t, store, "feat", "task-off", asagiri.StatusImplemented, []string{"echo ok"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
	require.Equal(t, 0, verifyEvidenceGateHistoryCount(fresh.PayloadJSON))
	require.NoFileExists(t, gateLogJSONPath(svc.repoRoot, task.ID, verifyEvidenceGateName))
}

func TestVerifyFeatureGatePassVerifiedWithHistory(t *testing.T) {
	svc, store := verifyFeatureGateService(t, true, nil, false)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return passVerifyEvidenceGateYAML(), nil
	}
	task := seedTaskWithValidationCommands(t, store, "feat", "task-pass", asagiri.StatusImplemented, []string{"echo ok"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
	require.Equal(t, 1, verifyEvidenceGateHistoryCount(fresh.PayloadJSON))
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "pass", entry.Status)
	require.FileExists(t, gateLogJSONPath(svc.repoRoot, task.ID, verifyEvidenceGateName))
}

func TestVerifyFeatureGateFailStaysImplemented(t *testing.T) {
	svc, store := verifyFeatureGateService(t, true, nil, false)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return failVerifyEvidenceGateYAML(), nil
	}
	task := seedTaskWithValidationCommands(t, store, "feat", "task-fail", asagiri.StatusImplemented, []string{"echo ok"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.Error(t, err)
	require.Contains(t, err.Error(), "verify evidence gate failed")
	require.Contains(t, err.Error(), "asa verify feat --task task-fail --force")

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusImplemented, fresh.Status)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestVerifyFeatureValidationFailSkipsGate(t *testing.T) {
	svc, store := verifyFeatureGateService(t, true, nil, false)
	svc.verifyEvidenceGateAgentHook = func(_ context.Context, _ string) (string, error) {
		t.Fatal("verify evidence gate must not run when validation commands fail")
		return "", nil
	}
	task := seedTaskWithValidationCommands(t, store, "feat", "task-val-ko", asagiri.StatusImplemented,
		[]string{"echo ok", "__asa_missing_validation_cmd__"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.Error(t, err)
	require.NotContains(t, err.Error(), "verify evidence gate")

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerifyFailed, fresh.Status)
	require.Equal(t, 0, verifyEvidenceGateHistoryCount(fresh.PayloadJSON))
	require.NoFileExists(t, gateLogJSONPath(svc.repoRoot, task.ID, verifyEvidenceGateName))
}

func TestVerifyFeatureGateDryRunVerifiedWithSimulatedPass(t *testing.T) {
	svc, store := verifyFeatureGateService(t, true, nil, true)
	task := seedTaskWithValidationCommands(t, store, "feat", "task-dry-gate", asagiri.StatusImplemented, []string{"echo ok"})

	_, err := svc.VerifyFeature(context.Background(), "feat", task.ID, true)
	require.NoError(t, err)

	fresh, err := store.GetTask(task.ID)
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusVerified, fresh.Status)
	entry, ok := lastVerifyEvidenceHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "pass", entry.Status)
	require.True(t, entry.DryRun)
}
