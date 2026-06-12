package workflow

import (
	"context"
	"os"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func enrichFeatureGateService(t *testing.T, enabled bool, warnAdvisory *bool, dryRun bool) (*Service, *sqlite.Store) {
	t.Helper()
	svc, store := newEnrichGateService(t, enabled, warnAdvisory)
	svc.dryRun = dryRun
	return svc, store
}

func enrichGateHistoryCount(payloadJSON string) int {
	canonical, err := payloadToCanonical(payloadJSON)
	if err != nil || canonical.Gates == nil {
		return 0
	}
	n := 0
	for _, e := range canonical.Gates.History {
		if e.Gate == enrichGateName {
			n++
		}
	}
	return n
}

func TestEnrichFeatureGateDisabledPreservesBehavior(t *testing.T) {
	svc, store := enrichFeatureGateService(t, false, nil, true)
	seedTask(t, store, "feat", "task-off", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-off", "reviewer", false)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-off")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusEnriched, fresh.Status)
	require.Equal(t, 0, enrichGateHistoryCount(fresh.PayloadJSON))
	require.NoFileExists(t, gateLogJSONPath(svc.repoRoot, "task-off", enrichGateName))
}

func TestEnrichFeatureGatePassEnrichedWithHistory(t *testing.T) {
	svc, store := enrichFeatureGateService(t, true, nil, true)
	seedTask(t, store, "feat", "task-pass", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-pass", "reviewer", false)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-pass")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusEnriched, fresh.Status)
	require.Equal(t, 1, enrichGateHistoryCount(fresh.PayloadJSON))
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "pass", entry.Status)
	require.FileExists(t, gateLogJSONPath(svc.repoRoot, "task-pass", enrichGateName))
	require.FileExists(t, gateLogMarkdownPath(svc.repoRoot, "task-pass", enrichGateName))
}

func TestEnrichFeatureGateWarnAdvisoryEnriched(t *testing.T) {
	svc, store := enrichFeatureGateService(t, true, nil, false)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnEnrichGateYAML(), nil
	}
	seedTask(t, store, "feat", "task-warn", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-warn", "reviewer", false)
	require.NoError(t, err)

	fresh, err := store.GetTask("task-warn")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusEnriched, fresh.Status)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "warn", entry.Status)
	require.Equal(t, 1, enrichGateHistoryCount(fresh.PayloadJSON))
}

func TestEnrichFeatureGateFailStaysPlanned(t *testing.T) {
	svc, store := enrichFeatureGateService(t, true, nil, false)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return failEnrichGateYAML(), nil
	}
	seedTask(t, store, "feat", "task-fail", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-fail", "reviewer", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate failed")

	fresh, err := store.GetTask("task-fail")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusPlanned, fresh.Status)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
	require.Equal(t, 1, enrichGateHistoryCount(fresh.PayloadJSON))
	require.FileExists(t, gateLogJSONPath(svc.repoRoot, "task-fail", enrichGateName))
}

func TestEnrichFeatureGateWarnNonAdvisoryStaysPlanned(t *testing.T) {
	f := false
	svc, store := enrichFeatureGateService(t, true, &f, false)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return warnEnrichGateYAML(), nil
	}
	seedTask(t, store, "feat", "task-warn-block", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-warn-block", "reviewer", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate warn (non-advisory)")

	fresh, err := store.GetTask("task-warn-block")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusPlanned, fresh.Status)
}

func TestEnrichFeatureGateParseErrorStaysPlanned(t *testing.T) {
	svc, store := enrichFeatureGateService(t, true, nil, false)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return "not a gate block", nil
	}
	seedTask(t, store, "feat", "task-parse", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-parse", "reviewer", false)
	require.Error(t, err)
	require.Contains(t, err.Error(), "enrich gate failed")

	fresh, err := store.GetTask("task-parse")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusPlanned, fresh.Status)
	entry, ok := lastEnrichHistoryEntry(t, fresh.PayloadJSON)
	require.True(t, ok)
	require.Equal(t, "fail", entry.Status)
}

func TestEnrichFeatureAgentErrorStaysPlanned(t *testing.T) {
	svc, store := newEnrichGateService(t, false, nil)
	svc.cfg.Agents["reviewer"] = config.Agent{Command: "false"}
	seedTask(t, store, "feat", "task-agent", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-agent", "reviewer", false)
	require.Error(t, err)

	fresh, err := store.GetTask("task-agent")
	require.NoError(t, err)
	require.Equal(t, asagiri.StatusPlanned, fresh.Status)
}

func TestEnrichFeatureGatePassLogsWritten(t *testing.T) {
	svc, store := enrichFeatureGateService(t, true, nil, false)
	svc.enrichGateAgentHook = func(_ context.Context, _ string) (string, error) {
		return passEnrichGateYAML(), nil
	}
	seedTask(t, store, "feat", "task-logs", asagiri.StatusPlanned)

	_, err := svc.EnrichFeature(context.Background(), "feat", "task-logs", "reviewer", false)
	require.NoError(t, err)

	logJSON := gateLogJSONPath(svc.repoRoot, "task-logs", enrichGateName)
	require.FileExists(t, logJSON)
	body, err := os.ReadFile(logJSON)
	require.NoError(t, err)
	require.Contains(t, string(body), `"status": "pass"`)
}
