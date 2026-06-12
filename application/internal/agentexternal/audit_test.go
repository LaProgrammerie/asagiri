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

func TestAuditMissingPathEmbeddedDefaults(t *testing.T) {
	dir := t.TempDir()
	report, err := agentexternal.Audit(dir, nil)
	require.NoError(t, err)
	require.Equal(t, agentexternal.ReportVersion, report.ReportVersion)
	require.True(t, report.ReadOnly)
	require.NotEmpty(t, report.Targets)
	for _, target := range report.Targets {
		require.NotEmpty(t, target.AgentID)
		require.NotEmpty(t, target.DesiredHash)
		if target.ConfiguredPath == "" {
			require.Equal(t, agentexternal.StatusMissingPath, target.Status)
		}
	}
}

func TestAuditExternalFileDrift(t *testing.T) {
	dir := t.TempDir()
	registry := filepath.Join(dir, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(registry, 0o755))

	extFile := filepath.Join(dir, "external-dev.md")
	require.NoError(t, os.WriteFile(extFile, []byte("# external prompt\n"), 0o644))

	specYAML := `id: dev
version: "1.0.0"
role: dev
system_prompt: inline
output_contract:
  format: asagiri-v1
external:
  kind: cursor-agent
  path: ` + extFile + `
  last_synced_hash: "0000000000000000000000000000000000000000000000000000000000000000"
`
	require.NoError(t, os.WriteFile(filepath.Join(registry, "dev.yaml"), []byte(specYAML), 0o644))

	report, err := agentexternal.Audit(dir, nil)
	require.NoError(t, err)

	var dev *agentexternal.ExternalTarget
	for i := range report.Targets {
		if report.Targets[i].AgentID == "dev" {
			dev = &report.Targets[i]
			break
		}
	}
	require.NotNil(t, dev)
	require.Equal(t, extFile, dev.DetectedPath)
	require.NotEmpty(t, dev.InstalledHash)
	require.Equal(t, agentexternal.StatusDrift, dev.Status)
	require.Equal(t, "spec.external.path", dev.PathSource)
}

func TestAuditConfigExternalPath(t *testing.T) {
	dir := t.TempDir()
	extFile := filepath.Join(dir, "kiro-dev.md")
	require.NoError(t, os.WriteFile(extFile, []byte("kiro profile"), 0o644))

	cfg := &config.Config{
		Agents: map[string]config.Agent{
			"dev": {
				Provider:     "kiro-cli",
				ExternalPath: extFile,
			},
		},
		Providers: map[string]config.ProviderConfig{
			"kiro-cli": {Type: "kiro-cli", Command: "nonexistent-kiro-cli-for-test"},
		},
	}

	report, err := agentexternal.Audit(dir, cfg)
	require.NoError(t, err)

	var dev *agentexternal.ExternalTarget
	for i := range report.Targets {
		if report.Targets[i].AgentID == "dev" {
			dev = &report.Targets[i]
			break
		}
	}
	require.NotNil(t, dev)
	require.Equal(t, extFile, dev.DetectedPath)
	require.Equal(t, "config.agents.external_path", dev.PathSource)
	require.Equal(t, agentexternal.StatusCLIMissing, dev.Status)
}
