package coordination_test

import (
	"context"
	"os"
	osExec "os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := osExec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestEnsureWorktreeCreatesAndCleansGitWorktree(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "test@example.com")
	runGit(t, repo, "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "README.md"), []byte("hello"), 0o644))
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "init")

	graphID := "graph-2026-05-29-a1b2c3d4"
	nodeID := "implement-node"
	branch := "asagiri/coord/graph-2026-05-29-a1b2c3d4/implement-node"

	path, cleanup, err := coordination.EnsureWorktree(context.Background(), repo, graphID, nodeID, branch)
	require.NoError(t, err)
	require.NotEmpty(t, path)
	_, err = os.Stat(path)
	require.NoError(t, err)

	cleanup()
	_, err = os.Stat(path)
	require.Error(t, err)
}
