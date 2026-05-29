package executiongraph

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewGraphID(t *testing.T) {
	id := NewGraphID()
	require.Regexp(t, `^graph-\d{4}-\d{2}-\d{2}-[a-f0-9]{8}$`, id)
	require.NoError(t, ValidateGraphID(id))
}

func TestNewGraphIDUnique(t *testing.T) {
	a := NewGraphID()
	b := NewGraphID()
	require.NotEqual(t, a, b)
}

func TestValidateGraphID(t *testing.T) {
	require.NoError(t, ValidateGraphID("graph-2026-05-27-001"))
	require.NoError(t, ValidateGraphID("graph-2026-05-27-a1b2c3d4"))
	require.Error(t, ValidateGraphID(""))
	require.Error(t, ValidateGraphID("../bad"))
	require.Error(t, ValidateGraphID(`graph-2026-05-27-bad\path`))
	require.Error(t, ValidateGraphID("graph-bad-format"))
	require.ErrorIs(t, ValidateGraphID("../bad"), ErrInvalidGraphID)
}

func TestRepositorySaveLoadRoundTrip(t *testing.T) {
	repoRoot := t.TempDir()
	repo := NewRepository(repoRoot)

	body := readGoldenFixture(t, "simple-linear")
	graph, err := ParseYAML(body)
	require.NoError(t, err)

	yamlPath, jsonPath, err := repo.Save(graph)
	require.NoError(t, err)
	require.FileExists(t, yamlPath)
	require.FileExists(t, jsonPath)

	loaded, err := repo.Load(graph.ID)
	require.NoError(t, err)
	require.Equal(t, graph.ID, loaded.ID)
	require.Equal(t, graph.Product, loaded.Product)
	require.Equal(t, graph.Status, loaded.Status)
	require.Len(t, loaded.Nodes, len(graph.Nodes))
	require.Len(t, loaded.Edges, len(graph.Edges))
	require.NoError(t, loaded.Validate())

	jsonBody, err := os.ReadFile(jsonPath)
	require.NoError(t, err)
	var decoded ExecutionGraph
	require.NoError(t, json.Unmarshal(jsonBody, &decoded))
	require.Equal(t, graph.ID, decoded.ID)
}

func TestRepositoryLoadInvalidID(t *testing.T) {
	repo := NewRepository(t.TempDir())
	_, err := repo.Load("../bad")
	require.Error(t, err)
}

func TestRepositoryLoadIDMismatch(t *testing.T) {
	repoRoot := t.TempDir()
	repo := NewRepository(repoRoot)
	graphID := "graph-2026-05-27-001"
	dir := filepath.Join(repoRoot, ".asagiri", "graphs", graphID)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := []byte(`id: graph-2026-05-27-002
product: workspace-saas
status: planned
strategy:
  max_parallel: 1
nodes: []
edges: []
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "execution-graph.yaml"), body, 0o644))

	_, err := repo.Load(graphID)
	require.Error(t, err)
	require.Contains(t, err.Error(), "id mismatch")
}

func TestRepositoryLoadInvalidGraphOnDisk(t *testing.T) {
	repoRoot := t.TempDir()
	repo := NewRepository(repoRoot)
	graphID := "graph-2026-05-27-001"
	dir := filepath.Join(repoRoot, ".asagiri", "graphs", graphID)
	require.NoError(t, os.MkdirAll(dir, 0o755))
	body := []byte(`id: graph-2026-05-27-001
status: planned
strategy:
  max_parallel: 1
nodes: []
edges: []
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "execution-graph.yaml"), body, 0o644))

	_, err := repo.Load(graphID)
	require.Error(t, err)
	require.ErrorIs(t, err, ErrInvalidGraph)
}

func TestRepositorySaveRejectsInvalidGraph(t *testing.T) {
	repo := NewRepository(t.TempDir())
	_, _, err := repo.Save(ExecutionGraph{ID: "graph-2026-05-27-a1b2c3d4"})
	require.Error(t, err)
}

func TestRepositoryArtifactPaths(t *testing.T) {
	repoRoot := t.TempDir()
	repo := NewRepository(repoRoot)

	graph, err := ParseYAML(readGoldenFixture(t, "simple-linear"))
	require.NoError(t, err)

	yamlPath, jsonPath, err := repo.Save(graph)
	require.NoError(t, err)

	wantDir := filepath.Join(repoRoot, ".asagiri", "graphs", graph.ID)
	require.Equal(t, filepath.Join(wantDir, "execution-graph.yaml"), yamlPath)
	require.Equal(t, filepath.Join(wantDir, "execution-graph.json"), jsonPath)
}

func TestRepositoryLoadJSONFallback(t *testing.T) {
	repoRoot := t.TempDir()
	repo := NewRepository(repoRoot)

	graph, err := ParseYAML(readGoldenFixture(t, "simple-linear"))
	require.NoError(t, err)

	dir := filepath.Join(repoRoot, ".asagiri", "graphs", graph.ID)
	require.NoError(t, os.MkdirAll(dir, 0o755))

	jsonBody, err := json.MarshalIndent(graph, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "execution-graph.json"), jsonBody, 0o644))

	loaded, err := repo.Load(graph.ID)
	require.NoError(t, err)
	require.Equal(t, graph.ID, loaded.ID)
	require.Equal(t, graph.Product, loaded.Product)
	require.NoError(t, loaded.Validate())
}

func readGoldenFixture(t *testing.T, scenario string) []byte {
	t.Helper()
	path := filepath.Join("testdata", "execution-graph", scenario, "execution-graph.yaml")
	body, err := os.ReadFile(path)
	require.NoError(t, err)
	return body
}

func goldenScenarios() []string {
	return []string{
		"simple-linear",
		"parallel-independent",
		"blocked-by-contract",
		"high-risk-security",
		"rollback-required",
	}
}
