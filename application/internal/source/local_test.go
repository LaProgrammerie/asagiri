package source

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestLocalSourceListAndSyncDryRun(t *testing.T) {
	repo := t.TempDir()
	specDir := filepath.Join(repo, ".kiro", "specs", "payment-routing")
	require.NoError(t, os.MkdirAll(specDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(specDir, "requirements.md"), []byte("# Pay"), 0o644))

	src := &LocalSource{
		RepoRoot: repo,
		Config: config.LocalSourceConfig{
			Enabled: true,
			Paths:   []string{".kiro/specs"},
		},
	}
	items, err := src.List(context.Background())
	require.NoError(t, err)
	require.NotEmpty(t, items)

	doc, err := src.Fetch(context.Background(), SourceRef{Name: "payment-routing"})
	require.NoError(t, err)
	res, err := WriteLocalSpec(repo, ".agentflow/specs", "payment-routing", doc, SyncOptions{Force: true})
	require.NoError(t, err)
	require.DirExists(t, res.Path)
	require.FileExists(t, filepath.Join(res.Path, "spec.md"))
}
