package bus

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/replay"
	"github.com/stretchr/testify/require"
)

func TestGetPaletteEntriesDynamicFlowsReportsReplays(t *testing.T) {
	repoRoot := t.TempDir()
	product := filepath.Join(repoRoot, ".asagiri", "products", "demo", "flows")
	require.NoError(t, os.MkdirAll(product, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(product, "checkout.flow.yaml"), []byte("id: checkout\n"), 0o644))

	trustDir := filepath.Join(repoRoot, ".asagiri", "trust", "run-2026-05-29")
	require.NoError(t, os.MkdirAll(trustDir, 0o755))

	replayDir := filepath.Join(repoRoot, replay.RelDir, "pkg-001")
	require.NoError(t, os.MkdirAll(replayDir, 0o755))

	qb := NewQueryBus(Deps{RepoRoot: repoRoot})
	raw, err := qb.Query(context.Background(), GetPaletteEntriesQuery{Screen: "mission", Limit: 200})
	require.NoError(t, err)
	res, ok := raw.(PaletteEntriesResult)
	require.True(t, ok)

	ids := map[string]bool{}
	for _, entry := range res.Entries {
		ids[entry.ID] = true
	}
	require.True(t, ids["flow.open.checkout"], "dynamic flow entry")
	require.True(t, ids["report.trust.run-2026-05-29"], "dynamic trust report entry")
	require.True(t, ids["replay.open.pkg-001"], "dynamic replay open entry")
	require.True(t, ids["replay.run.pkg-001"], "dynamic replay run entry")
}

func TestGetPaletteEntriesContextualGraphScreen(t *testing.T) {
	qb := NewQueryBus(Deps{RepoRoot: t.TempDir()})
	raw, err := qb.Query(context.Background(), GetPaletteEntriesQuery{Screen: "graph", Limit: 200})
	require.NoError(t, err)
	res, ok := raw.(PaletteEntriesResult)
	require.True(t, ok)

	found := false
	for _, entry := range res.Entries {
		if entry.ID == "ctx.graph-rollback" {
			found = true
			break
		}
	}
	require.True(t, found)
}
