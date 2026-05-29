package cli_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cli"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func TestMemoryReindexOnCorpus(t *testing.T) {
	dir := t.TempDir()
	cfgDir := filepath.Join(dir, ".asagiri")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(`
project:
  name: mem-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: hash
`), 0o644))

	store, err := runtime.Open(dir)
	require.NoError(t, err)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:     runtime.ScopeProject,
		Type:      "note",
		Summary:   "checkout payment timeout",
		Relevance: 0.8,
	}))
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:     runtime.ScopeFlow,
		Type:      "note",
		Summary:   "onboarding invitation email",
		Relevance: 0.7,
	}))
	require.NoError(t, store.Close())

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "reindex"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "reindexed: 2")
}
