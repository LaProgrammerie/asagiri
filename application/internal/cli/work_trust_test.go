package cli

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

func TestResolveWorkStrictTrustRequiresTrustFlow(t *testing.T) {
	repo := t.TempDir()
	_, _, err := resolveWorkStrictTrust(repo, "", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--trust-flow")
}

func TestResolveWorkStrictTrustFromTrustFlowFlag(t *testing.T) {
	repo := t.TempDir()
	copyMinimalProductForWorkTrust(t, repo)
	flow, product, err := resolveWorkStrictTrust(repo, "workspace-onboarding", "")
	require.NoError(t, err)
	require.Equal(t, "workspace-onboarding", flow)
	require.Equal(t, "minimal-product", product)
}

func TestResolveWorkStrictTrustFromFeatureHint(t *testing.T) {
	repo := t.TempDir()
	copyMinimalProductForWorkTrust(t, repo)
	flow, product, err := resolveWorkStrictTrust(repo, "", "workspace-onboarding")
	require.NoError(t, err)
	require.Equal(t, "workspace-onboarding", flow)
	require.Equal(t, "minimal-product", product)
}

func TestResolveWorkStrictTrustFeatureNoMatch(t *testing.T) {
	repo := t.TempDir()
	copyMinimalProductForWorkTrust(t, repo)
	_, _, err := resolveWorkStrictTrust(repo, "", "agentflow-test")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--trust-flow")
}

func TestResolveWorkStrictTrustUnknownTrustFlow(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri", "products", "demo", "flows"), 0o755))
	_, _, err := resolveWorkStrictTrust(repo, "missing", "")
	require.Error(t, err)
	require.Contains(t, err.Error(), "--trust-flow")
}
