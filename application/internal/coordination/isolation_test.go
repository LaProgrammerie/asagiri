package coordination_test

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/coordination"
)

func TestWorktreePathIsolated(t *testing.T) {
	repo := t.TempDir()
	path, err := coordination.WorktreePath(repo, "graph-2026-05-29-a1b2c3d4", "implement-node")
	require.NoError(t, err)
	require.Contains(t, path, filepath.Join(".asagiri", "worktrees", "graph-2026-05-29-a1b2c3d4"))
	require.NotContains(t, path, "..")
}

func TestWorktreePathRejectsTraversal(t *testing.T) {
	_, err := coordination.WorktreePath(t.TempDir(), "../bad", "node")
	require.Error(t, err)
}
