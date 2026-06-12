package agentspec_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/stretchr/testify/require"
)

func TestLoadAllEmbeddedDefaultsWhenRegistryMissing(t *testing.T) {
	dir := t.TempDir()
	loader := agentspec.NewLoader(dir)
	require.True(t, loader.UsingEmbeddedDefaults())

	specs, err := loader.LoadAll()
	require.NoError(t, err)
	require.Len(t, specs, 5)

	ids := map[string]struct{}{}
	for _, spec := range specs {
		ids[spec.ID] = struct{}{}
		require.Equal(t, "embedded", spec.Source)
		require.NotEmpty(t, spec.ContentHash)
	}
	require.Contains(t, ids, "dev")
	require.Contains(t, ids, "reviewer")
	require.Contains(t, ids, "enricher")
	require.Contains(t, ids, "governance")
	require.Contains(t, ids, "gate")
}

func TestLoadDiskOnlyMissingFile(t *testing.T) {
	repo := t.TempDir()
	loader := agentspec.NewLoader(repo)
	_, err := loader.LoadDiskOnly("dev")
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")
}

func TestLoadDiskOnlyWithoutEmbeddedFallback(t *testing.T) {
	repo := t.TempDir()
	agentsDir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(agentsDir, 0o755))

	data, err := os.ReadFile(filepath.Join("testdata", "agents", "valid-dev.yaml"))
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "valid-dev.yaml"), data, 0o644))

	loader := agentspec.NewLoader(repo)
	spec, err := loader.LoadDiskOnly("valid-dev")
	require.NoError(t, err)
	require.Equal(t, "valid-dev", spec.ID)
	require.NotEqual(t, "embedded", spec.Source)

	_, err = loader.Load("missing-embedded")
	require.Error(t, err)
}

func TestLoadFromDiskRegistry(t *testing.T) {
	repo := t.TempDir()
	agentsDir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(agentsDir, 0o755))

	src := filepath.Join("testdata", "agents", "valid-dev.yaml")
	data, err := os.ReadFile(src)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "valid-dev.yaml"), data, 0o644))

	loader := agentspec.NewLoader(repo)
	require.False(t, loader.UsingEmbeddedDefaults())

	spec, err := loader.Load("valid-dev")
	require.NoError(t, err)
	require.Equal(t, "dev", spec.Role)
	require.Equal(t, agentspec.OutputAsagiriV1, spec.OutputContract.Format)
}

func TestLoadInvalidYAML(t *testing.T) {
	_, err := agentspec.Parse([]byte("id: [unclosed"), "broken.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "parse YAML")
}

func TestLoadInvalidSpecReadableError(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "agents", "invalid-missing-prompt.yaml"))
	require.NoError(t, err)

	_, err = agentspec.Parse(data, "invalid-missing-prompt.yaml")
	require.Error(t, err)
	require.Contains(t, err.Error(), "system_prompt")
}

func TestDuplicateIDAcrossFiles(t *testing.T) {
	repo := t.TempDir()
	agentsDir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(agentsDir, 0o755))

	for _, name := range []string{"dup-a.yaml", "dup-b.yaml"} {
		data, err := os.ReadFile(filepath.Join("testdata", "agents", name))
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(filepath.Join(agentsDir, name), data, 0o644))
	}

	loader := agentspec.NewLoader(repo)
	_, err := loader.LoadAll()
	require.Error(t, err)
	require.Contains(t, err.Error(), "dupliqué")
}

func TestLoadMissingID(t *testing.T) {
	loader := agentspec.NewLoader(t.TempDir())
	_, err := loader.Load("missing-agent")
	require.Error(t, err)
	require.Contains(t, err.Error(), "introuvable")
}

func TestListReturnsMeta(t *testing.T) {
	loader := agentspec.NewLoader(t.TempDir())
	meta, err := loader.List()
	require.NoError(t, err)
	require.Len(t, meta, 5)
	for _, m := range meta {
		require.NotEmpty(t, m.ID)
		require.NotEmpty(t, m.ContentHash)
	}
}
