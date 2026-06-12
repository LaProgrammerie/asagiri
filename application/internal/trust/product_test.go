package trust

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestResolveProductIDFromFlowFile(t *testing.T) {
	repo := t.TempDir()
	productDir := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows")
	require.NoError(t, os.MkdirAll(productDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "workspace-onboarding.flow.yaml"),
		[]byte("id: workspace-onboarding\nentry_screen: s1\nsteps:\n  - id: s1\n    screen: s1\n    action: go\n"), 0o644))

	id, err := ResolveProductID(repo, "workspace-onboarding")
	require.NoError(t, err)
	require.Equal(t, "minimal-product", id)
}

func TestResolveProductFlowExact(t *testing.T) {
	repo := t.TempDir()
	productDir := filepath.Join(repo, ".asagiri", "products", "minimal-product", "flows")
	require.NoError(t, os.MkdirAll(productDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(productDir, "workspace-onboarding.flow.yaml"),
		[]byte("id: workspace-onboarding\nentry_screen: s1\nsteps:\n  - id: s1\n    screen: s1\n    action: go\n"), 0o644))

	flow, product, err := ResolveProductFlow(repo, "workspace-onboarding")
	require.NoError(t, err)
	require.Equal(t, "workspace-onboarding", flow)
	require.Equal(t, "minimal-product", product)
}

func TestResolveProductFlowMissing(t *testing.T) {
	repo := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(repo, ".asagiri", "products", "demo", "flows"), 0o755))
	_, _, err := ResolveProductFlow(repo, "missing-flow")
	require.Error(t, err)
	require.Contains(t, err.Error(), "no product flow")
}
