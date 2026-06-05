package checks

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
)

func setupMinimalProductRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	src := filepath.Join("testdata", "minimal-product")
	dest := filepath.Join(repo, ".asagiri", "products", "minimal-product")
	require.NoError(t, copyDir(src, dest))

	graphsSrc := filepath.Join("testdata", "graphs-minimal.json")
	analysisDir := filepath.Join(repo, ".asagiri", "analysis", "minimal-product")
	require.NoError(t, os.MkdirAll(analysisDir, 0o755))
	require.NoError(t, copyFile(graphsSrc, filepath.Join(analysisDir, "graphs.json")))
	return repo
}

func testDeps(t *testing.T, repo string) Dependencies {
	t.Helper()
	deps := DefaultDependencies()
	deps.LoadBundle = func(repoRoot, productID string) (analysis.Bundle, error) {
		return analysis.LoadBundle(repoRoot, productID)
	}
	deps.ParseFailedTests = func(ctx context.Context, repoRoot string) ([]string, error) {
		return nil, nil
	}
	deps.Investigate = func(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (investigation.InvestigationResult, error) {
		return investigation.InvestigationResult{
			CandidateFiles: []string{"application/internal/trust/engine.go"},
		}, nil
	}
	return deps
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	_, err = io.Copy(out, in)
	return err
}

func copyDir(src, dst string) error {
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}
