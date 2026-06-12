package gates

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"github.com/stretchr/testify/require"
)

func TestHumanReviewPendingSubmitPhase(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
	task := sqlite.Task{
		ID:          "task-1",
		Status:      asagiri.StatusImplemented,
		PayloadJSON: `{}`,
	}
	pg, ok := BlockingPendingForTask(repo, cfg, task)
	require.True(t, ok)
	require.Equal(t, HumanReviewGateName, pg.Gate)
	require.Equal(t, PendingPhaseSubmit, pg.Phase)
	require.True(t, pg.Blocking)
}

func TestHumanReviewPendingResumePhase(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
	task := sqlite.Task{ID: "task-2", Status: asagiri.StatusImplemented, PayloadJSON: `{}`}
	dir := filepath.Join(repo, ".asagiri", "logs", "task-2", "gates")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, DefaultHumanReviewVerdictFile), []byte("human_review:\n  status: pass\n"), 0o644))

	pg, ok := BlockingPendingForTask(repo, cfg, task)
	require.True(t, ok)
	require.Equal(t, PendingPhaseResume, pg.Phase)
}

func TestHumanReviewNotPendingWhenSatisfied(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
	payload := `{"gates":{"history":[{"gate":"human_review","status":"pass","confidence":1}]}}`
	task := sqlite.Task{ID: "task-3", Status: asagiri.StatusImplemented, PayloadJSON: payload}
	_, ok := BlockingPendingForTask(repo, cfg, task)
	require.False(t, ok)
}

func TestHumanReviewPendingOnVerifiedWithoutHistory(t *testing.T) {
	repo := t.TempDir()
	cfg := &config.Config{
		Work: config.WorkConfig{
			Gates: config.WorkGatesConfig{
				HumanReview: config.WorkHumanReviewGateConfig{
					Enabled: true,
					Mode:    config.GovernanceModePerTask,
				},
			},
		},
	}
	task := sqlite.Task{ID: "task-verified", Status: asagiri.StatusVerified, PayloadJSON: `{}`}
	pg, ok := BlockingPendingForTask(repo, cfg, task)
	require.True(t, ok)
	require.Equal(t, HumanReviewGateName, pg.Gate)
	require.Equal(t, PendingPhaseSubmit, pg.Phase)
}

func TestFormatPendingActionSubmit(t *testing.T) {
	pg := PendingGate{Gate: HumanReviewGateName, Scope: "task-1", Blocking: true, Phase: PendingPhaseSubmit}
	msg := FormatPendingAction(pg, "feat")
	require.Contains(t, msg, "Gate human_review requires action")
	require.Contains(t, msg, "asa gates submit human_review --task task-1")
	require.Contains(t, msg, "asa continue --yes")
}

func TestGateEntrySatisfiedWarnAdvisory(t *testing.T) {
	entry := asagiri.GateHistoryEntry{Status: "warn"}
	require.True(t, GateEntrySatisfied(true, entry))
	require.False(t, GateEntrySatisfied(false, entry))
}
