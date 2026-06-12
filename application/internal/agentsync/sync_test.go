package agentsync_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/agentsync"
	"github.com/stretchr/testify/require"
)

func TestSyncDryRunNoWrite(t *testing.T) {
	repo := t.TempDir()
	report, err := agentsync.Sync(repo, agentsync.Options{})
	require.NoError(t, err)
	require.Equal(t, "check", report.Mode)
	require.False(t, report.Wrote)
	require.NotEmpty(t, report.Items)

	dir := filepath.Join(repo, agentspec.RegistryDir)
	_, err = os.Stat(dir)
	require.True(t, os.IsNotExist(err))

	for _, item := range report.Items {
		require.Equal(t, agentsync.ActionCreate, item.Action)
	}
}

func TestSyncWriteCreatesFiles(t *testing.T) {
	repo := t.TempDir()
	report, err := agentsync.Sync(repo, agentsync.Options{Write: true})
	require.NoError(t, err)
	require.True(t, report.Wrote)

	dir := filepath.Join(repo, agentspec.RegistryDir)
	entries, err := os.ReadDir(dir)
	require.NoError(t, err)
	require.Len(t, entries, len(report.Items))

	report2, err := agentsync.Sync(repo, agentsync.Options{Write: true})
	require.NoError(t, err)
	for _, item := range report2.Items {
		require.Equal(t, agentsync.ActionSkip, item.Action)
	}
}

func TestSyncConflictWithoutForce(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dev.yaml"), []byte(`id: dev
version: "9.9.9"
role: dev
system_prompt: custom
output_contract:
  format: asagiri-v1
`), 0o644))

	report, err := agentsync.Sync(repo, agentsync.Options{Write: true})
	require.NoError(t, err)
	require.True(t, agentsync.HasBlockingConflicts(report))

	var devItem agentsync.Item
	for _, item := range report.Items {
		if item.ID == "dev" {
			devItem = item
		}
	}
	require.Equal(t, agentsync.ActionConflict, devItem.Action)

	data, err := os.ReadFile(filepath.Join(dir, "dev.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(data), "9.9.9")
}

func TestSyncForceOverwrite(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dev.yaml"), []byte(`id: dev
version: "9.9.9"
role: dev
system_prompt: custom
output_contract:
  format: asagiri-v1
`), 0o644))

	report, err := agentsync.Sync(repo, agentsync.Options{Write: true, Force: true, AgentID: "dev"})
	require.NoError(t, err)
	require.False(t, agentsync.HasBlockingConflicts(report))

	var devItem agentsync.Item
	for _, item := range report.Items {
		if item.ID == "dev" {
			devItem = item
		}
	}
	require.Equal(t, agentsync.ActionUpdate, devItem.Action)

	data, err := os.ReadFile(filepath.Join(dir, "dev.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(data), "1.0.0")
	require.NotContains(t, string(data), "9.9.9")
}
