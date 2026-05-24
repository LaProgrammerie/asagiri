package worktree

import (
	"context"
	"os"
	osExec "os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := osExec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func TestCreateAndRemoveWorktree(t *testing.T) {
	repo := t.TempDir()
	runGit(t, repo, "init")
	runGit(t, repo, "config", "user.email", "test@example.com")
	runGit(t, repo, "config", "user.name", "Test")
	require.NoError(t, os.WriteFile(filepath.Join(repo, "README.md"), []byte("hello"), 0o644))
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-m", "init")

	basePath := filepath.Join(repo, ".agentflow", "worktrees")
	mgr := New(repo, basePath, "agentflow", "", false)

	path, _, err := mgr.Create(context.Background(), "feature-a", "task-001")
	require.NoError(t, err)
	_, err = os.Stat(path)
	require.NoError(t, err)

	require.NoError(t, mgr.Remove(context.Background(), path))
	_, err = os.Stat(path)
	require.Error(t, err)
}
