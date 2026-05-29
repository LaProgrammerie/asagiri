package knowledge_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/stretchr/testify/require"
)

func TestGraphSnapshotter(t *testing.T) {
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

	result, err := knowledge.DefaultSnapshotter().Snapshot(context.Background(), knowledge.SnapshotRequest{
		RepoRoot: repo,
		Name:     "before-refactor",
	})
	require.NoError(t, err)
	require.Equal(t, "before-refactor", result.Name)
	require.NotEmpty(t, result.ID)

	snapDir := filepath.Join(repo, ".asagiri", "knowledge", "snapshots", "before-refactor")
	_, err = os.Stat(filepath.Join(snapDir, "metadata.json"))
	require.NoError(t, err)
	_, err = os.Stat(filepath.Join(snapDir, "graph.json"))
	require.NoError(t, err)
}

func TestGraphSnapshotterRequiresName(t *testing.T) {
	t.Parallel()
	_, err := knowledge.DefaultSnapshotter().Snapshot(context.Background(), knowledge.SnapshotRequest{
		RepoRoot: t.TempDir(),
		Name:     "",
	})
	require.Error(t, err)
}

func TestStubSnapshotterReturnsNotImplemented(t *testing.T) {
	t.Parallel()
	_, err := (knowledge.StubSnapshotter{}).Snapshot(t.Context(), knowledge.SnapshotRequest{})
	require.ErrorIs(t, err, knowledge.ErrNotImplemented)
}
