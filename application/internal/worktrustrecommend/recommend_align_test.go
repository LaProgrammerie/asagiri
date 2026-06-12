package worktrustrecommend

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/worktrust"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestRecommendationAlignsWithIntent(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Work: config.WorkConfig{
			DefaultAgent:    "dev-a",
			DefaultEnricher: "enrich-a",
			DefaultReviewer: "review-a",
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{Enabled: true, Mode: config.GovernanceModePerTask},
				VerifyEvidence: config.WorkVerifyEvidenceGateConfig{
					Enabled: true, Mode: config.GovernanceModePerTask,
				},
			},
		},
	}
	cases := []struct {
		name  string
		task  sqlite.Task
		setup func(t *testing.T, dir string, taskID string)
	}{
		{
			name: "planned enrich required",
			task: sqlite.Task{ID: "t-plan", Feature: "feat", Status: asagiri.StatusPlanned, PayloadJSON: `{}`},
		},
		{
			name: "enriched dev",
			task: sqlite.Task{
				ID: "t-dev", Feature: "feat", Status: asagiri.StatusEnriched,
				PayloadJSON: payloadWithGate(gates.EnrichGateName, "pass", 0.9),
			},
		},
		{
			name: "implemented verify",
			task: sqlite.Task{ID: "t-impl", Feature: "feat", Status: asagiri.StatusImplemented, PayloadJSON: `{}`},
		},
		{
			name: "verified review",
			task: sqlite.Task{
				ID: "t-ver", Feature: "feat", Status: asagiri.StatusVerified,
				PayloadJSON: payloadWithGate("verify_evidence", "pass", 0.85),
			},
			setup: func(t *testing.T, dir, taskID string) {
				writeValidationResults(t, dir, taskID)
			},
		},
		{
			name: "verify failed force",
			task: sqlite.Task{ID: "t-vf", Feature: "feat", Status: asagiri.StatusVerifyFailed, PayloadJSON: `{}`},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setup != nil {
				tc.setup(t, dir, tc.task.ID)
			}
			want, err := intent.RecommendForTask(dir, cfg, tc.task)
			require.NoError(t, err)
			report, err := worktrust.BuildTaskReport(dir, cfg, tc.task)
			require.NoError(t, err)
			rec := RecommendationFromIntent(dir, cfg, tc.task, report)
			require.Equal(t, want.Primitive, rec.Command)
		})
	}
}

func TestFeatureNextAlignsWithIntent(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{Work: config.WorkConfig{DefaultAgent: "cursor"}}
	tasks := []sqlite.Task{
		{ID: "t1", Feature: "feat", Status: asagiri.StatusEnriched, PayloadJSON: payloadWithGate(gates.EnrichGateName, "pass", 0.9)},
		{ID: "t2", Feature: "feat", Status: asagiri.StatusReviewed, PayloadJSON: `{}`},
	}
	want, err := intent.RecommendNextFromTasks(dir, cfg, "feat", tasks)
	require.NoError(t, err)
	report, err := worktrust.BuildFeatureReport(dir, cfg, "feat", tasks)
	require.NoError(t, err)
	actions := FeatureNextFromIntent(dir, cfg, "feat", tasks, report.Tasks)
	require.Len(t, actions, 1)
	require.Equal(t, want.Primitive, actions[0].Command)
}

func payloadWithGate(gate, status string, confidence float64) string {
	payload := map[string]any{
		"gates": map[string]any{
			"history": []map[string]any{{
				"gate": gate, "status": status, "at": "2026-06-06T12:00:00Z", "confidence": confidence,
			}},
		},
	}
	b, _ := json.Marshal(payload)
	return string(b)
}

func writeValidationResults(t *testing.T, dir, taskID string) {
	t.Helper()
	logDir := filepath.Join(dir, ".asagiri", "logs", taskID, "validation")
	require.NoError(t, os.MkdirAll(logDir, 0o755))
	doc := map[string]any{
		"task_id": taskID,
		"at":      "2026-06-06T12:00:00Z",
		"commands": []map[string]any{
			{"name": "test", "command": "go test", "exit_code": 0},
		},
	}
	b, err := json.Marshal(doc)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(logDir, "results.json"), b, 0o644))
}
