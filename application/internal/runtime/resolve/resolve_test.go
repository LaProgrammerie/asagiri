package resolve_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/configcache"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime/resolve"
	"github.com/stretchr/testify/require"
)

func testConfig() *config.Config {
	cfg := config.NewTestConfig("t")
	cfg.Agents = map[string]config.Agent{
		"dev": {Provider: "exec", Model: "local-model"},
	}
	return cfg
}

func writeDevSpec(t *testing.T, repo string) {
	t.Helper()
	dir := filepath.Join(repo, agentspec.RegistryDir)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := []byte(`id: dev
version: "1.0.0"
role: developer
provider_targets:
  - exec
system_prompt: test
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "dev.yaml"), body, 0o644))
}

func TestResolveLocalWinsOverRegistry(t *testing.T) {
	repo := t.TempDir()
	writeDevSpec(t, repo)
	home := t.TempDir()
	store := configcache.NewStore(home)

	_, err := store.SaveBundle("demo", cloud.BundleFetchResult{
		Bundle: cloud.ConfigBundle{
			"contentHash": "abc",
			"registry": map[string]any{
				"agentDefinitions": map[string]any{
					"dev": map[string]any{
						"defaultProviderKey": "openai",
						"defaultModelKey":    "gpt-4o",
					},
				},
			},
		},
		ContentHash:    "abc",
		ReleaseVersion: 1,
	})
	require.NoError(t, err)

	cfg := testConfig()
	cfg.Cloud.RegistryFallback = true
	cfg.Cloud.Project = "demo"

	res := resolve.ResolveAgent(resolve.Options{
		RepoRoot:    repo,
		Config:      cfg,
		AgentKey:    "dev",
		Cache:       store,
		ProjectSlug: "demo",
	})
	require.Equal(t, resolve.SourceLocal, res.Source)
	require.Equal(t, "exec", res.Provider)
	require.Equal(t, "local-model", res.Model)
}

func TestResolveRegistryFallbackWhenNoLocalSpec(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	store := configcache.NewStore(home)

	_, err := store.SaveBundle("demo", cloud.BundleFetchResult{
		Bundle: cloud.ConfigBundle{
			"contentHash": "def",
			"provenance": map[string]any{
				"releaseVersion": float64(2),
			},
			"registry": map[string]any{
				"agentDefinitions": map[string]any{
					"cursor-dev": map[string]any{
						"role":               "developer",
						"defaultProviderKey": "openai",
						"defaultModelKey":    "gpt-4o",
						"activeVersion": map[string]any{
							"version":        float64(3),
							"capabilityKeys": []any{"dev"},
						},
					},
				},
			},
		},
		ContentHash:    "def",
		ReleaseVersion: 2,
	})
	require.NoError(t, err)

	cfg := testConfig()
	cfg.Cloud.RegistryFallback = true
	cfg.Cloud.Project = "demo"

	res := resolve.ResolveAgent(resolve.Options{
		RepoRoot:    repo,
		Config:      cfg,
		AgentKey:    "cursor-dev",
		Cache:       store,
		ProjectSlug: "demo",
	})
	require.Equal(t, resolve.SourceRegistry, res.Source)
	require.Equal(t, "openai", res.Provider)
	require.Equal(t, "gpt-4o", res.Model)
	require.Equal(t, 3, res.RegistryAgentVersion)
	require.Equal(t, 2, res.ConfigReleaseVersion)
}

func TestResolveDefaultsWhenRegistryFallbackDisabled(t *testing.T) {
	repo := t.TempDir()
	cfg := testConfig()
	cfg.Cloud.RegistryFallback = false

	res := resolve.ResolveAgent(resolve.Options{
		RepoRoot: repo,
		Config:   cfg,
		AgentKey: "dev",
		Cache:    configcache.NewStore(t.TempDir()),
	})
	require.Equal(t, resolve.SourceLocal, res.Source)
	require.Equal(t, "exec", res.Provider)
}

func TestResolveOfflineUsesCachedRegistry(t *testing.T) {
	repo := t.TempDir()
	home := t.TempDir()
	store := configcache.NewStore(home)

	_, err := store.SaveBundle("demo", cloud.BundleFetchResult{
		Bundle: cloud.ConfigBundle{
			"contentHash": "ghi",
			"registry": map[string]any{
				"agentDefinitions": map[string]any{
					"cursor-dev": map[string]any{
						"defaultProviderKey": "anthropic",
						"defaultModelKey":    "claude",
					},
				},
			},
		},
		ContentHash:    "ghi",
		ReleaseVersion: 1,
	})
	require.NoError(t, err)

	cfg := testConfig()
	cfg.Cloud.RegistryFallback = true
	cfg.Cloud.Project = "demo"
	cfg.Cloud.RegistryCacheMaxAge = "1h"

	res := resolve.ResolveAgent(resolve.Options{
		RepoRoot:    repo,
		Config:      cfg,
		AgentKey:    "cursor-dev",
		Cache:       store,
		ProjectSlug: "demo",
		Now:         time.Now().UTC().Add(2 * time.Hour),
	})
	require.Equal(t, resolve.SourceRegistry, res.Source)
	require.Equal(t, "anthropic", res.Provider)
}
