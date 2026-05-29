package knowledge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestIncrementalBuildSkipsUnchangedExtractors(t *testing.T) {
	fixtureRoot := filepath.Join("testdata", "knowledge-graph", "onboarding-flow", "fixture")
	repo := t.TempDir()
	copyDir(t, fixtureRoot, repo)

	first, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:         repo,
		IncludeFlows:     true,
		IncludeContracts: true,
		Scope:            "workspace-saas",
	})
	require.NoError(t, err)
	require.Greater(t, first.Nodes, 0)

	second, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:         repo,
		Incremental:      true,
		IncludeFlows:     true,
		IncludeContracts: true,
		Scope:            "workspace-saas",
	})
	require.NoError(t, err)
	require.False(t, second.Rebuilt)
	require.Contains(t, second.SkippedExtractors, "flows")
	require.Contains(t, second.SkippedExtractors, "contracts")
	require.Equal(t, first.Nodes, second.Nodes)
}

func TestIncrementalBuildReextractsAfterMtimeChange(t *testing.T) {
	fixtureRoot := filepath.Join("testdata", "knowledge-graph", "stale-graph", "fixture")
	repo := t.TempDir()
	copyDir(t, fixtureRoot, repo)

	_, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:     repo,
		IncludeFlows: true,
		Scope:        "demo",
	})
	require.NoError(t, err)

	flowPath := filepath.Join(repo, ".asagiri", "products", "demo", "flows", "checkout.flow.yaml")
	require.NoError(t, touchFile(flowPath))

	second, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:     repo,
		Incremental:  true,
		IncludeFlows: true,
		Scope:        "demo",
	})
	require.NoError(t, err)
	require.NotContains(t, second.SkippedExtractors, "flows")
}

func touchFile(path string) error {
	now := time.Now().Add(2 * time.Hour)
	return os.Chtimes(path, now, now)
}
