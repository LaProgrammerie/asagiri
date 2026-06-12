package agentexternal_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentexternal"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func testExternalSyncSetup(t *testing.T) (repoRoot, extPath string, cfg *config.Config) {
	t.Helper()
	repoRoot = t.TempDir()
	extPath = filepath.Join(repoRoot, "profiles", "dev.md")
	registry := filepath.Join(repoRoot, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(registry, 0o755))

	specYAML := `id: dev
version: "1.0.0"
role: dev
provider_targets:
  - kiro-cli
system_prompt: |
  Prompt dev export
output_contract:
  format: asagiri-v1
external:
  kind: kiro-agent
  path: ` + extPath + `
`
	require.NoError(t, os.WriteFile(filepath.Join(registry, "dev.yaml"), []byte(specYAML), 0o644))

	cfg = &config.Config{
		Agents: map[string]config.Agent{
			"dev": {Provider: "kiro-cli"},
		},
		Providers: map[string]config.ProviderConfig{
			"kiro-cli": {Type: "kiro-cli", Command: "nonexistent-kiro-for-test"},
		},
	}
	return repoRoot, extPath, cfg
}

func TestSyncDryRunDoesNotWrite(t *testing.T) {
	repoRoot, extPath, cfg := testExternalSyncSetup(t)

	report, err := agentexternal.Sync(repoRoot, cfg, agentexternal.SyncOptions{AgentID: "dev"})
	require.NoError(t, err)
	require.Equal(t, "check", report.Mode)
	require.False(t, report.Wrote)
	require.Len(t, report.Items, 1)
	require.Equal(t, agentexternal.SyncActionCreate, report.Items[0].Action)
	_, err = os.Stat(extPath)
	require.True(t, os.IsNotExist(err))
}

func TestSyncWriteCreatesFile(t *testing.T) {
	repoRoot, extPath, cfg := testExternalSyncSetup(t)

	report, err := agentexternal.Sync(repoRoot, cfg, agentexternal.SyncOptions{AgentID: "dev", Write: true})
	require.NoError(t, err)
	require.True(t, report.Wrote)
	require.Equal(t, agentexternal.SyncActionCreate, report.Items[0].Action)
	require.True(t, report.Items[0].SpecUpdated)

	data, err := os.ReadFile(extPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "asagiri: true")
	require.Contains(t, string(data), "Orchestrated prompt")

	specData, err := os.ReadFile(filepath.Join(repoRoot, agentspec.RegistryDir, "dev.yaml"))
	require.NoError(t, err)
	require.Contains(t, string(specData), "last_synced_hash:")
}

func TestSyncConflictBlocksWithoutForce(t *testing.T) {
	repoRoot, extPath, cfg := testExternalSyncSetup(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(extPath), 0o755))
	require.NoError(t, os.WriteFile(extPath, []byte("# manual edit\n"), 0o644))

	report, err := agentexternal.Sync(repoRoot, cfg, agentexternal.SyncOptions{AgentID: "dev", Write: true})
	require.NoError(t, err)
	require.Equal(t, agentexternal.SyncActionConflict, report.Items[0].Action)
	require.True(t, agentexternal.HasBlockingSyncConflicts(report))

	data, err := os.ReadFile(extPath)
	require.NoError(t, err)
	require.Equal(t, "# manual edit\n", string(data))
}

func TestSyncForceOverwrites(t *testing.T) {
	repoRoot, extPath, cfg := testExternalSyncSetup(t)
	require.NoError(t, os.MkdirAll(filepath.Dir(extPath), 0o755))
	require.NoError(t, os.WriteFile(extPath, []byte("# manual edit\n"), 0o644))

	report, err := agentexternal.Sync(repoRoot, cfg, agentexternal.SyncOptions{AgentID: "dev", Write: true, Force: true})
	require.NoError(t, err)
	require.True(t, report.Wrote)
	require.Equal(t, agentexternal.SyncActionUpdate, report.Items[0].Action)

	data, err := os.ReadFile(extPath)
	require.NoError(t, err)
	require.Contains(t, string(data), "asagiri: true")
}

func TestSyncRejectsMissingPath(t *testing.T) {
	repoRoot := t.TempDir()
	cfg := &config.Config{
		Agents: map[string]config.Agent{
			"dev": {Provider: "kiro-cli"},
		},
		Providers: map[string]config.ProviderConfig{
			"kiro-cli": {Type: "kiro-cli", Command: "x"},
		},
	}

	report, err := agentexternal.Sync(repoRoot, cfg, agentexternal.SyncOptions{AgentID: "dev"})
	require.NoError(t, err)
	require.Equal(t, agentexternal.SyncActionReject, report.Items[0].Action)
}

func TestRenderProviderMarkdownFrontmatter(t *testing.T) {
	spec := agentspec.Spec{
		ID:           "dev",
		Version:      "1.0.0",
		Role:         agentspec.RoleDev,
		SystemPrompt: "hello",
		OutputContract: agentspec.OutputContract{
			Format:         agentspec.OutputAsagiriV1,
			RequiredFields: []string{"status", "summary"},
		},
	}
	out := agentexternal.RenderProviderMarkdown(spec, "cursor-agent")
	require.Contains(t, out, "---")
	require.Contains(t, out, "agent_id: dev")
	require.Contains(t, out, "content_hash:")
	require.Contains(t, out, "## Orchestrated prompt")
}
