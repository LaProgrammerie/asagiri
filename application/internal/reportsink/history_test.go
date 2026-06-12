package reportsink

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestSaveTrustTaskArchivesHistory(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	first := map[string]string{"score": "70"}
	_, err := SaveTrustTask(repo, "task-h1", first)
	require.NoError(t, err)

	second := map[string]string{"score": "80"}
	_, err = SaveTrustTask(repo, "task-h1", second)
	require.NoError(t, err)

	history, err := ListHistory(repo, "reports/trust/tasks/task-h1.json")
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Contains(t, history[0].RelPath, "reports/trust/tasks/history/task-h1_")
	require.Contains(t, history[0].RelPath, ".json")

	before, after, err := DiffPairPaths(repo, "reports/trust/tasks/task-h1.json")
	require.NoError(t, err)
	require.Equal(t, history[0].RelPath, before)
	require.Equal(t, ".asagiri/reports/trust/tasks/task-h1.json", after)
}

func TestSaveDoctorArchivesHistory(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	_, err := SaveDoctor(repo, map[string]any{"ready": true, "n": 1})
	require.NoError(t, err)
	_, err = SaveDoctor(repo, map[string]any{"ready": false, "n": 2})
	require.NoError(t, err)

	history, err := ListHistory(repo, doctorLatestRel())
	require.NoError(t, err)
	require.Len(t, history, 1)
	require.Contains(t, history[0].RelPath, "reports/doctor/history/")
}

func TestDiffPairPathsRequiresTwoSaves(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	_, err := SaveTrustTask(repo, "task-once", map[string]string{"v": "1"})
	require.NoError(t, err)

	_, _, err = DiffPairPaths(repo, "reports/trust/tasks/task-once.json")
	require.Error(t, err)
	require.Contains(t, err.Error(), "historique insuffisant")
}

func TestSaveWithoutHistory(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	opts := SaveOptions{KeepHistory: false}
	_, err := SaveTrustTaskWithOptions(repo, "task-nh", map[string]string{"v": "1"}, opts)
	require.NoError(t, err)
	_, err = SaveTrustTaskWithOptions(repo, "task-nh", map[string]string{"v": "2"}, opts)
	require.NoError(t, err)

	history, err := ListHistory(repo, "reports/trust/tasks/task-nh.json")
	require.NoError(t, err)
	require.Empty(t, history)
}
