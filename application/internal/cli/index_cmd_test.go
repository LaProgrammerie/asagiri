package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func writeIndexTestConfig(t *testing.T, repo string) {
	t.Helper()
	cfgDir := filepath.Join(repo, ".asagiri")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	body := `project:
  name: index-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: hash
`
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(body), 0o644))
}

func TestIndexSearchKeyword(t *testing.T) {
	repo := t.TempDir()
	initGitRepo(t, repo)
	writeIndexTestConfig(t, repo)
	appDir := filepath.Join(repo, "application")
	require.NoError(t, os.MkdirAll(appDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(appDir, "needle.go"), []byte("// specialneedle token\n"), 0o644))

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := RootCommand()
	var buildBuf bytes.Buffer
	root.SetOut(&buildBuf)
	root.SetErr(&buildBuf)
	root.SetArgs([]string{"index", "--skip-embeddings"})
	require.NoError(t, root.Execute())

	var searchBuf bytes.Buffer
	root.SetOut(&searchBuf)
	root.SetErr(&searchBuf)
	root.SetArgs([]string{"index", "search", "specialneedle", "--keyword"})
	require.NoError(t, root.Execute())
	require.Contains(t, searchBuf.String(), "application/needle.go")
}
