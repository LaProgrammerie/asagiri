package workcli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func copyMinimalProductForWorkTrust(t *testing.T, repo string) {
	t.Helper()
	src := filepath.Join("..", "trust", "checks", "testdata", "minimal-product")
	dest := filepath.Join(repo, ".asagiri", "products", "minimal-product")
	require.NoError(t, copyDirTrust(src, dest))
}

func copyDirTrust(src, dst string) error {
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
		in, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, in, 0o644)
	})
}

func TestResolveStrictTrustRequiresTrustFlow(t *testing.T) {
	repo := t.TempDir()
	_, _, err := ResolveStrictTrust(repo, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--trust-flow")
}

func TestResolveStrictTrustFromTrustFlowFlag(t *testing.T) {
	repo := t.TempDir()
	copyMinimalProductForWorkTrust(t, repo)
	flow, product, err := ResolveStrictTrust(repo, "workspace-onboarding", "")
	require.NoError(t, err)
	require.Equal(t, "workspace-onboarding", flow)
	require.Equal(t, "minimal-product", product)
}

func TestResolveStrictTrustFromFeatureHint(t *testing.T) {
	repo := t.TempDir()
	copyMinimalProductForWorkTrust(t, repo)
	flow, product, err := ResolveStrictTrust(repo, "", "workspace-onboarding")
	require.NoError(t, err)
	require.Equal(t, "workspace-onboarding", flow)
	require.Equal(t, "minimal-product", product)
}

func TestResolveStrictTrustFeatureNoMatch(t *testing.T) {
	repo := t.TempDir()
	copyMinimalProductForWorkTrust(t, repo)
	_, _, err := ResolveStrictTrust(repo, "", "agentflow-test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--trust-flow")
}

func TestResolveStrictTrustUnknownTrustFlow(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri", "products", "demo", "flows"), 0o755))
	_, _, err := ResolveStrictTrust(repo, "missing", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--trust-flow")
}
