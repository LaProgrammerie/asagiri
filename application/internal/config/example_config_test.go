package config_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func repoRoot(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	require.True(t, ok)
	dir := filepath.Dir(file)
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, ".asagiri", "config.yaml.example")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("repo root not found")
		}
		dir = parent
	}
}

func TestExampleConfigLoadsAndResolves(t *testing.T) {
	root := repoRoot(t)
	path := config.ExamplePath(root)
	cfg, err := config.Load(path, root)
	require.NoError(t, err)
	require.NoError(t, cfg.Validate(root))

	workAgents := []string{
		cfg.Work.DefaultSpecAgent,
		cfg.Work.DefaultAgent,
		cfg.Work.DefaultReviewer,
		cfg.Work.DefaultEnricher,
	}
	for _, name := range workAgents {
		require.NotEmpty(t, name)
		_, err := cfg.LookupAgent(name)
		require.NoError(t, err, "work agent %q", name)
		providerType, merged, err := cfg.MergedAgentRuntime(name)
		require.NoError(t, err, "merged runtime %q", name)
		require.NotEmpty(t, providerType)
		require.NotEmpty(t, merged.Command, "command for work agent %q", name)
	}

	for id := range cfg.Providers {
		p, err := cfg.LookupProvider(id)
		require.NoError(t, err, "provider %q", id)
		require.NotEmpty(t, p.Type)
		require.True(t, config.IsKnownProviderType(p.Type), "provider %q type %q", id, p.Type)
	}

	for agentName := range cfg.Agents {
		if cfg.Agents[agentName].Provider == "" {
			continue
		}
		_, err := cfg.AgentProviderType(agentName)
		require.NoError(t, err, "agent %q", agentName)
	}
}
