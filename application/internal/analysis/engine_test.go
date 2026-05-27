package analysis_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/stretchr/testify/require"
)

func TestBuildAllWorkspaceSaaS(t *testing.T) {
	t.Parallel()
	repo, err := os.Getwd()
	require.NoError(t, err)
	for filepath.Base(repo) != "hyper-fast-builder" && repo != "/" {
		repo = filepath.Dir(repo)
	}
	if filepath.Base(repo) != "hyper-fast-builder" {
		t.Skip("repo root not found")
	}
	b, err := analysis.BuildAll(repo, "workspace-saas")
	require.NoError(t, err)
	require.Equal(t, "workspace-saas", b.Product)
	require.NotEmpty(t, b.Graphs["flow"])
	require.NotEmpty(t, b.Graphs["symbol"])
}

func TestWriteBundle(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	productDir := filepath.Join(dir, ".asagiri/products/demo/flows")
	require.NoError(t, os.MkdirAll(productDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "demo.flow.yaml"), []byte(`id: demo
title: Demo
steps:
  - id: s1
    screen: home
`), 0o644))
	b, err := analysis.BuildAll(dir, "demo")
	require.NoError(t, err)
	path, err := analysis.WriteBundle(dir, b)
	require.NoError(t, err)
	require.FileExists(t, path)
}
