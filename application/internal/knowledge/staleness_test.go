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

func TestGraphStalenessDetectorNoBuild(t *testing.T) {
	t.Parallel()
	repo := t.TempDir()
	report, err := knowledge.DefaultStalenessDetector().Check(context.Background(), repo)
	require.NoError(t, err)
	require.True(t, report.Stale)
	require.Contains(t, report.RecommendCommand, "knowledge build")
}

func TestGraphStalenessDetectorFreshAfterBuild(t *testing.T) {
	fixtureRoot := filepath.Join("testdata", "knowledge-graph", "onboarding-flow", "fixture")
	repo := t.TempDir()
	copyDir(t, fixtureRoot, repo)

	_, err := knowledge.DefaultBuilder().Build(context.Background(), knowledge.BuildRequest{
		RepoRoot:         repo,
		IncludeFlows:     true,
		IncludeContracts: true,
		Scope:            "workspace-saas",
	})
	require.NoError(t, err)

	report, err := knowledge.DefaultStalenessDetector().Check(context.Background(), repo)
	require.NoError(t, err)
	require.False(t, report.Stale)
	require.Zero(t, report.FilesChanged)
}

func TestGraphStalenessDetectorStaleAfterFileChange(t *testing.T) {
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
	require.NoError(t, os.Chtimes(flowPath, time.Now().Add(time.Hour), time.Now().Add(time.Hour)))

	report, err := knowledge.DefaultStalenessDetector().Check(context.Background(), repo)
	require.NoError(t, err)
	require.True(t, report.Stale)
	require.GreaterOrEqual(t, report.FilesChanged, 1)
	require.Equal(t, "asa knowledge build --incremental", report.RecommendCommand)
}

func TestFormatStalenessTemplate(t *testing.T) {
	t.Parallel()
	out := knowledge.FormatStaleness(knowledge.StalenessReport{
		Stale:            true,
		FilesChanged:     3,
		EdgesOutdated:    2,
		RecommendCommand: "asa knowledge build --incremental",
	})
	require.Contains(t, out, "Knowledge graph stale")
	require.Contains(t, out, "3 files changed")
	require.Contains(t, out, "2 edges may be outdated")
}

func TestStubStalenessDetectorReturnsNotImplemented(t *testing.T) {
	t.Parallel()
	_, err := (knowledge.StubStalenessDetector{}).Check(t.Context(), t.TempDir())
	require.ErrorIs(t, err, knowledge.ErrNotImplemented)
}
