package reportsink

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRequireInitializedMissing(t *testing.T) {
	dir := t.TempDir()
	err := RequireInitialized(dir)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrRuntimeAbsent)
}

func TestSaveTrustTaskOverwriteStable(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	first := map[string]string{"verdict": "acceptable", "n": "1"}
	rel, err := SaveTrustTask(repo, "task-save-1", first)
	require.NoError(t, err)
	require.Equal(t, ".asagiri/reports/trust/tasks/task-save-1.json", rel)

	abs, err := TrustTaskAbs(repo, "task-save-1")
	require.NoError(t, err)
	raw, err := os.ReadFile(abs)
	require.NoError(t, err)
	require.Contains(t, string(raw), `"acceptable"`)

	second := map[string]string{"verdict": "risky", "n": "2"}
	rel2, err := SaveTrustTask(repo, "task-save-1", second)
	require.NoError(t, err)
	require.Equal(t, rel, rel2)

	raw2, err := os.ReadFile(abs)
	require.NoError(t, err)
	require.Contains(t, string(raw2), `"risky"`)
	require.NotContains(t, string(raw2), `"acceptable"`)
}

func TestSaveDoctorLatest(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	report := map[string]any{"report_version": "doctor-v1", "ready": true}
	rel, err := SaveDoctor(repo, report)
	require.NoError(t, err)
	require.Equal(t, ".asagiri/reports/doctor/latest.json", rel)

	raw, err := os.ReadFile(DoctorLatestAbs(repo))
	require.NoError(t, err)
	var decoded map[string]any
	require.NoError(t, json.Unmarshal(raw, &decoded))
	require.Equal(t, "doctor-v1", decoded["report_version"])
}

func TestSaveRejectsUnsafeID(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri"), 0o755))

	_, err := SaveTrustTask(repo, "../escape", map[string]string{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "invalid report id")
}
