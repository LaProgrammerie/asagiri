package rag

import (
	"database/sql"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestIndexFixtureRepo(t *testing.T) {
	repo := t.TempDir()
	src := filepath.Join(repo, "application")
	if err := os.MkdirAll(src, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	res, err := Index(IndexOptions{RepoRoot: repo, Paths: []string{"application"}, DryRun: false})
	if err != nil {
		t.Fatal(err)
	}
	if res.Files < 1 || res.Chunks < 1 {
		t.Fatalf("files=%d chunks=%d", res.Files, res.Chunks)
	}
}

func TestIndexStoresEmbeddings(t *testing.T) {
	repo := t.TempDir()
	src := filepath.Join(repo, "application")
	require.NoError(t, os.MkdirAll(src, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "main.go"), []byte("package main\n"), 0o644))

	res, err := Index(IndexOptions{
		RepoRoot: repo,
		Paths:    []string{"application"},
		Memory:   config.RuntimeMemoryConfig{Embedder: "hash"},
	})
	require.NoError(t, err)
	require.Equal(t, res.Chunks, res.EmbeddedChunks)
	require.True(t, res.EmbedderConfigured)

	db, err := sql.Open("sqlite", res.DBPath)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	var withEmb int
	require.NoError(t, db.QueryRow(
		`SELECT COUNT(1) FROM chunks WHERE embedding IS NOT NULL AND embedding != ''`,
	).Scan(&withEmb))
	require.Equal(t, res.Chunks, withEmb)
}
