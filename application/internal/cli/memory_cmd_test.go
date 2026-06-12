package cli_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cli"
	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/stretchr/testify/require"
)

func writeMemoryConfig(t *testing.T, dir string, embedderBlock string) {
	t.Helper()
	cfgDir := filepath.Join(dir, ".asagiri")
	require.NoError(t, os.MkdirAll(cfgDir, 0o755))
	body := `project:
  name: mem-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: hash
`
	if embedderBlock != "" {
		body = embedderBlock
	}
	require.NoError(t, os.WriteFile(filepath.Join(cfgDir, "config.yaml"), []byte(body), 0o644))
}

func TestMemoryReindexOnCorpus(t *testing.T) {
	embedder.Configure(embedder.NewHash())
	dir := t.TempDir()
	writeMemoryConfig(t, dir, "")

	store, err := runtime.Open(dir)
	require.NoError(t, err)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:     runtime.ScopeProject,
		Type:      "note",
		Summary:   "checkout payment timeout",
		Relevance: 0.8,
	}))
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:     runtime.ScopeFlow,
		Type:      "note",
		Summary:   "onboarding invitation email",
		Relevance: 0.7,
	}))
	require.NoError(t, store.Close())

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "reindex"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "reindexed: 2")
}

func TestMemoryDoctorHealthyHash(t *testing.T) {
	embedder.Configure(embedder.NewHash())
	dir := t.TempDir()
	writeMemoryConfig(t, dir, "")

	store, err := runtime.Open(dir)
	require.NoError(t, err)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:     runtime.ScopeProject,
		Type:      "note",
		Summary:   "ok",
		Relevance: 0.8,
	}))
	require.NoError(t, store.Close())

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "doctor"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "✓ ollama")
	require.Contains(t, buf.String(), "✓ dimensions")
	require.Contains(t, buf.String(), "✓ orphans")
	require.Contains(t, buf.String(), "Mémoire runtime saine.")
}

func TestMemoryDoctorOllamaUnreachable(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			http.Error(w, "down", http.StatusServiceUnavailable)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	writeMemoryConfig(t, dir, `project:
  name: mem-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: ollama
    ollama:
      base_url: `+srv.URL+`
      model: test-model
`)

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "doctor"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	err := root.Execute()
	require.Error(t, err)
	require.Contains(t, buf.String(), "✗ ollama")
}

func TestMemoryDoctorOllamaReachableAndDimensions(t *testing.T) {
	const dim = 4
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"models": []any{}}))
		case "/api/embeddings":
			var req struct {
				Prompt string `json:"prompt"`
			}
			require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
			vec := make([]float32, dim)
			vec[0] = 1
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"embedding": vec}))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	writeMemoryConfig(t, dir, `project:
  name: mem-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: ollama
    ollama:
      base_url: `+srv.URL+`
      model: test-model
`)

	store, err := runtime.Open(dir)
	require.NoError(t, err)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:         runtime.ScopeProject,
		Type:          "note",
		Summary:       "aligned",
		EmbeddingJSON: memory.MarshalEmbedding([]float32{1, 0, 0, 0}),
		Relevance:     0.8,
	}))
	require.NoError(t, store.Close())

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "doctor"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "✓ ollama")
	require.Contains(t, buf.String(), "✓ dimensions")
}

func TestMemoryDoctorDimensionMismatchOllama(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/tags":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"models": []any{}}))
		case "/api/embeddings":
			require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"embedding": []float32{1, 0, 0, 0}}))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	dir := t.TempDir()
	writeMemoryConfig(t, dir, `project:
  name: mem-test
state:
  backend: sqlite
  path: .asagiri/state.sqlite
runtime:
  memory:
    embedder: ollama
    ollama:
      base_url: `+srv.URL+`
      model: test-model
`)

	store, err := runtime.Open(dir)
	require.NoError(t, err)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:         runtime.ScopeProject,
		Type:          "note",
		Summary:       "wrong dim",
		EmbeddingJSON: memory.MarshalEmbedding([]float32{1, 2}),
		Relevance:     0.5,
	}))
	require.NoError(t, store.Close())

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "doctor"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	require.Error(t, root.Execute())
	require.Contains(t, buf.String(), "✗ dimensions")
}

func TestMemoryDoctorOrphanFlow(t *testing.T) {
	embedder.Configure(embedder.NewHash())
	dir := t.TempDir()
	writeMemoryConfig(t, dir, "")

	store, err := runtime.Open(dir)
	require.NoError(t, err)
	require.NoError(t, store.UpsertMemory(runtime.MemoryEntry{
		Scope:       runtime.ScopeFlow,
		Type:        "note",
		Summary:     "detached",
		LinkedFlows: []string{"ghost-flow"},
		Relevance:   0.6,
	}))
	require.NoError(t, store.Close())

	oldWD, _ := os.Getwd()
	require.NoError(t, os.Chdir(dir))
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	root := cli.RootCommand()
	root.SetArgs([]string{"memory", "doctor"})
	var buf bytes.Buffer
	root.SetOut(&buf)
	require.Error(t, root.Execute())
	require.Contains(t, buf.String(), "✗ orphans")
}
