package spec

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestReadFeatureFromKiroFiles(t *testing.T) {
	repo := t.TempDir()
	featureDir := filepath.Join(repo, ".kiro", "specs", "feature-a")
	require.NoError(t, os.MkdirAll(featureDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "requirements.md"), []byte("req"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(featureDir, "tasks.md"), []byte("- [ ] task"), 0o644))

	reader := NewReader(repo, &config.Config{
		Specs: config.Specs{
			KiroPath:       ".kiro/specs",
			ActiveSpecPath: "docs/ai/active/current-spec.md",
		},
	})
	doc, err := reader.ReadFeature("feature-a")
	require.NoError(t, err)
	require.Equal(t, "kiro", doc.Source)
	require.Equal(t, "req", doc.Requirements)
	require.Contains(t, doc.Tasks, "task")
}

func TestReadFeatureFallbackToActiveSpec(t *testing.T) {
	repo := t.TempDir()
	activePath := filepath.Join(repo, "docs", "ai", "active")
	require.NoError(t, os.MkdirAll(activePath, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(activePath, "current-spec.md"), []byte("active-spec"), 0o644))

	reader := NewReader(repo, &config.Config{
		Specs: config.Specs{
			KiroPath:       ".kiro/specs",
			ActiveSpecPath: "docs/ai/active/current-spec.md",
		},
	})
	doc, err := reader.ReadFeature("missing-feature")
	require.NoError(t, err)
	require.Equal(t, "active", doc.Source)
	require.Equal(t, "active-spec", doc.Active)
}
