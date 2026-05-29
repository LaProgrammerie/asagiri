package rag

import (
	"context"
	"database/sql"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	_ "modernc.org/sqlite"
)

// DefaultIndexDir is the local RAG index root (spec §10.3).
const DefaultIndexDir = ".asagiri/index"

// IndexOptions configures what to index.
type IndexOptions struct {
	RepoRoot string
	Paths    []string
	Exclude  []string
	DryRun   bool
	// Memory selects the embedder (runtime.memory); zero value uses hash after ConfigureFromConfig.
	Memory config.RuntimeMemoryConfig
	// SkipEmbeddings stores chunks without vectors (keyword-only retrieval).
	SkipEmbeddings bool
}

// IndexResult summarizes an index run.
type IndexResult struct {
	Files            int
	Chunks           int
	EmbeddedChunks   int
	DBPath           string
	EmbedderConfigured bool
}

var defaultExclude = []string{".git", "vendor", "node_modules", ".asagiri/index"}

// Index walks sources and stores chunks in chunks.sqlite.
func Index(opts IndexOptions) (IndexResult, error) {
	indexDir := filepath.Join(opts.RepoRoot, DefaultIndexDir)
	dbPath := filepath.Join(indexDir, "chunks.sqlite")
	if opts.DryRun {
		count := 0
		chunks := 0
		for _, rel := range opts.Paths {
			_ = filepath.WalkDir(filepath.Join(opts.RepoRoot, rel), func(path string, d fs.DirEntry, err error) error {
				if err != nil || d.IsDir() {
					return nil
				}
				if shouldSkip(path, opts.RepoRoot, append(defaultExclude, opts.Exclude...)) {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil
				}
				count++
				chunks += 1
				return nil
			})
		}
		return IndexResult{Files: count, Chunks: chunks, DBPath: dbPath}, nil
	}

	if err := os.MkdirAll(indexDir, 0o755); err != nil {
		return IndexResult{}, err
	}
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return IndexResult{}, err
	}
	defer db.Close()

	if err := initSchema(db); err != nil {
		return IndexResult{}, err
	}
	if _, err := db.Exec(`DELETE FROM chunks`); err != nil {
		return IndexResult{}, err
	}

	embedChunks := !opts.SkipEmbeddings
	if embedChunks {
		if err := embedder.ConfigureFromConfig(opts.Memory); err != nil {
			return IndexResult{}, fmt.Errorf("rag index embedder: %w", err)
		}
	}

	res := IndexResult{DBPath: dbPath, EmbedderConfigured: embedChunks}
	ctx := context.Background()
	for _, rel := range opts.Paths {
		root := filepath.Join(opts.RepoRoot, rel)
		if _, statErr := os.Stat(root); statErr != nil {
			if os.IsNotExist(statErr) {
				continue
			}
			return res, statErr
		}
		err := filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if d.IsDir() {
				if shouldSkip(path, opts.RepoRoot, append(defaultExclude, opts.Exclude...)) {
					return filepath.SkipDir
				}
				return nil
			}
			if shouldSkip(path, opts.RepoRoot, append(defaultExclude, opts.Exclude...)) {
				return nil
			}
			if isSecretOrBinary(path) {
				return nil
			}
			body, readErr := os.ReadFile(path)
			if readErr != nil {
				return nil
			}
			relPath, _ := filepath.Rel(opts.RepoRoot, path)
			parts := SplitText(relPath, string(body), 0)
			for _, ch := range parts {
				var embJSON *string
				if embedChunks {
					vec, embErr := embedder.EmbedText(ctx, ch.Content)
					if embErr != nil {
						return embErr
					}
					s := marshalEmbedding(vec)
					embJSON = &s
					res.EmbeddedChunks++
				}
				if _, insErr := db.Exec(
					`INSERT INTO chunks(path, offset, content, embedding) VALUES(?,?,?,?)`,
					ch.Path, ch.Offset, ch.Content, embJSON,
				); insErr != nil {
					return insErr
				}
				res.Chunks++
			}
			res.Files++
			return nil
		})
		if err != nil {
			return res, err
		}
	}
	return res, nil
}

func shouldSkip(path, repoRoot string, excludes []string) bool {
	rel, err := filepath.Rel(repoRoot, path)
	if err != nil {
		rel = path
	}
	rel = filepath.ToSlash(rel)
	for _, ex := range excludes {
		ex = strings.Trim(ex, "/")
		if rel == ex || strings.HasPrefix(rel, ex+"/") {
			return true
		}
	}
	return false
}

func isSecretOrBinary(path string) bool {
	base := strings.ToLower(filepath.Base(path))
	if strings.Contains(base, "secret") || strings.Contains(base, ".env") {
		return true
	}
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".png", ".jpg", ".jpeg", ".gif", ".zip", ".sqlite", ".db", ".exe", ".so", ".dylib":
		return true
	default:
		return false
	}
}

// DefaultIndexPaths returns Go-template index sources (spec §10.3, ADR-001).
func DefaultIndexPaths() []string {
	return []string{
		"application",
		"docs",
		".kiro",
		"spec.md",
		"go.mod",
	}
}

// OpenDB opens the chunks database for retrieval.
func OpenDB(repoRoot string) (*sql.DB, error) {
	dbPath := filepath.Join(repoRoot, DefaultIndexDir, "chunks.sqlite")
	if _, err := os.Stat(dbPath); err != nil {
		return nil, fmt.Errorf("index absent — lancez asa index: %w", err)
	}
	return sql.Open("sqlite", dbPath)
}
