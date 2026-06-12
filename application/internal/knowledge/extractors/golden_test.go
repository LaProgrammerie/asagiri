package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestGoldenMissingTestsBuildWarns(t *testing.T) {
	fixtureRoot := filepath.Join("..", "testdata", "knowledge-graph", "missing-tests", "fixture")
	repo := t.TempDir()
	copyDir(t, fixtureRoot, repo)

	result, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:     repo,
		Scope:        "demo",
		IncludeFlows: true,
		IncludeCode:  true,
		IncludeTests: true,
	})
	require.NoError(t, err)

	var warned bool
	for _, w := range result.Warnings {
		if strings.Contains(w, "archive_workspace") && strings.Contains(w, "no linked test") {
			warned = true
		}
	}
	require.True(t, warned, "warnings: %v", result.Warnings)
}

func copyDir(t *testing.T, src, dst string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(dst, 0o755))
	entries, err := os.ReadDir(src)
	require.NoError(t, err)
	for _, ent := range entries {
		srcPath := filepath.Join(src, ent.Name())
		dstPath := filepath.Join(dst, ent.Name())
		if ent.IsDir() {
			copyDir(t, srcPath, dstPath)
			continue
		}
		body, err := os.ReadFile(srcPath)
		require.NoError(t, err)
		require.NoError(t, os.WriteFile(dstPath, body, 0o644))
	}
}
