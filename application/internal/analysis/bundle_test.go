package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadBundle(t *testing.T) {
	repo := t.TempDir()
	product := "minimal"
	dir := filepath.Join(repo, analysisRel, product)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	b := Bundle{
		Product: product,
		Graphs: map[string]Graph{
			"api": {Kind: "api", Nodes: []Node{{ID: "n1", Kind: "route", Name: "POST /api/workspaces"}}},
		},
	}
	raw, err := json.Marshal(b)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "graphs.json"), raw, 0o644))

	loaded, err := LoadBundle(repo, product)
	require.NoError(t, err)
	require.Equal(t, product, loaded.Product)
	require.Contains(t, loaded.Graphs, "api")
}

func TestLoadBundleMissing(t *testing.T) {
	_, err := LoadBundle(t.TempDir(), "missing")
	require.Error(t, err)
}
