package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
)

func TestADRExtractor(t *testing.T) {
	repo := t.TempDir()
	dir := filepath.Join(repo, "docs", "decisions")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "024-engineering-knowledge-graph.md"),
		[]byte("# ADR-024 — Engineering Knowledge Graph\n"), 0o644))

	nodes, edges, warnings, err := extractors.ADRExtractor{}.Extract(context.Background(), repo, "")
	require.NoError(t, err)
	require.Empty(t, edges)
	require.Empty(t, warnings)
	require.Len(t, nodes, 1)
	require.Equal(t, "adr:024_engineering-knowledge-graph", nodes[0].ID)
}
