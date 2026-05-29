package rag

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/stretchr/testify/require"
)

func TestSemanticSearchRanksRelatedChunk(t *testing.T) {
	repo := t.TempDir()
	writeChunkFile(t, repo, "application/auth.go", "package auth\n// User login session token validation\n")
	writeChunkFile(t, repo, "application/billing.go", "package billing\n// Invoice PDF export monthly\n")

	_, err := Index(IndexOptions{
		RepoRoot: repo,
		Paths:    []string{"application"},
		Memory:   config.RuntimeMemoryConfig{Embedder: "hash"},
	})
	require.NoError(t, err)

	db, err := OpenDB(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	paths, err := NewRetriever(db).SearchWithOptions(context.Background(), "user authentication session", SearchOptions{
		Limit:  4,
		Memory: config.RuntimeMemoryConfig{Embedder: "hash"},
	})
	require.NoError(t, err)
	require.NotEmpty(t, paths)
	require.Equal(t, "application/auth.go", paths[0])
}

func TestKeywordSearchIgnoresEmbeddings(t *testing.T) {
	repo := t.TempDir()
	writeChunkFile(t, repo, "application/zebra.go", "// zebra stripe pattern uniquekeywordxyz\n")
	writeChunkFile(t, repo, "application/other.go", "// unrelated content\n")

	_, err := Index(IndexOptions{
		RepoRoot: repo,
		Paths:    []string{"application"},
		Memory:   config.RuntimeMemoryConfig{Embedder: "hash"},
	})
	require.NoError(t, err)

	db, err := OpenDB(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	paths, err := NewRetriever(db).SearchWithOptions(context.Background(), "uniquekeywordxyz", SearchOptions{
		Limit:       4,
		KeywordOnly: true,
		Memory:      config.RuntimeMemoryConfig{Embedder: "hash"},
	})
	require.NoError(t, err)
	require.Equal(t, []string{"application/zebra.go"}, paths)
}

func TestSemanticFallsBackToKeywordWithoutEmbeddings(t *testing.T) {
	repo := t.TempDir()
	writeChunkFile(t, repo, "application/findme.go", "// onlymatch token abcdef\n")

	_, err := Index(IndexOptions{
		RepoRoot:       repo,
		Paths:          []string{"application"},
		SkipEmbeddings: true,
	})
	require.NoError(t, err)

	db, err := OpenDB(repo)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })

	paths, err := NewRetriever(db).SearchWithOptions(context.Background(), "onlymatch", SearchOptions{Limit: 4})
	require.NoError(t, err)
	require.Equal(t, []string{"application/findme.go"}, paths)
}

func writeChunkFile(t *testing.T, repo, rel, body string) {
	t.Helper()
	dir := filepath.Dir(filepath.Join(repo, rel))
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(repo, rel), []byte(body), 0o644))
}
