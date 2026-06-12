package intent

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestRecommendForTask_MatchesRecommendNextFromTasks(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{
		Work: config.WorkConfig{
			DefaultAgent:    "dev-agent",
			DefaultEnricher: "enrich-agent",
			DefaultReviewer: "review-agent",
			Gates: config.WorkGatesConfig{
				Enrich: config.WorkEnrichGateConfig{Enabled: true, Mode: config.GovernanceModePerTask},
			},
		},
	}
	task := sqlite.Task{
		ID: "t1", Feature: "feat", Status: asagiri.StatusPlanned,
		PayloadJSON: `{}`,
	}
	recTask, err := RecommendForTask(dir, cfg, task)
	require.NoError(t, err)
	recFeat, err := RecommendNextFromTasks(dir, cfg, "feat", []sqlite.Task{task})
	require.NoError(t, err)
	require.Equal(t, recFeat.Primitive, recTask.Primitive)
	require.Contains(t, recTask.Primitive, "asa enrich feat --task t1 --agent enrich-agent --force")
}

func TestRecommendForTask_VerifyFailedMatchesNext(t *testing.T) {
	task := sqlite.Task{ID: "t1", Feature: "feat", Status: asagiri.StatusVerifyFailed}
	rec, err := RecommendForTask(t.TempDir(), &config.Config{}, task)
	require.NoError(t, err)
	require.Equal(t, "asa verify feat --task t1 --force", rec.Primitive)
}

func TestRecommendNextFromTasks_PlannedEnrichedDev(t *testing.T) {
	cfg := &config.Config{Work: config.WorkConfig{DefaultAgent: "cursor"}}
	payload := `{"gates":{"history":[{"gate":"` + gates.EnrichGateName + `","status":"pass","at":"2026-06-06T12:00:00Z","confidence":0.9}]}}`
	task := sqlite.Task{ID: "t2", Feature: "feat", Status: asagiri.StatusEnriched, PayloadJSON: payload}
	rec, err := RecommendNextFromTasks(t.TempDir(), cfg, "feat", []sqlite.Task{task})
	require.NoError(t, err)
	require.Equal(t, "asa dev feat --task t2 --agent cursor", rec.Primitive)
}
