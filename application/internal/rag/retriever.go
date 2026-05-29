package rag

import (
	"context"
	"database/sql"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
)

// SearchOptions configures chunk retrieval (PF-X-03).
type SearchOptions struct {
	Limit       int
	KeywordOnly bool
	Memory      config.RuntimeMemoryConfig
}

// Retriever finds context files for a task objective.
type Retriever struct {
	db *sql.DB
}

func NewRetriever(db *sql.DB) *Retriever {
	return &Retriever{db: db}
}

// Search returns up to limit paths matching the query.
func (r *Retriever) Search(query string, limit int) ([]string, error) {
	return r.SearchWithOptions(context.Background(), query, SearchOptions{Limit: limit})
}

// SearchWithOptions uses semantic cosine ranking when embeddings exist and KeywordOnly is false;
// otherwise falls back to SQL LIKE.
func (r *Retriever) SearchWithOptions(ctx context.Context, query string, opts SearchOptions) ([]string, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	limit := opts.Limit
	if limit <= 0 {
		limit = 8
	}
	if opts.KeywordOnly {
		return r.searchKeyword(query, limit)
	}
	hasEmb, err := indexHasEmbeddings(r.db)
	if err != nil {
		return nil, err
	}
	if hasEmb {
		if err := embedder.ConfigureFromConfig(opts.Memory); err != nil {
			return r.searchKeyword(query, limit)
		}
		paths, semErr := r.searchSemantic(ctx, query, limit)
		if semErr != nil {
			return nil, semErr
		}
		if len(paths) > 0 {
			return paths, nil
		}
	}
	return r.searchKeyword(query, limit)
}

func (r *Retriever) searchKeyword(query string, limit int) ([]string, error) {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil, nil
	}
	pattern := "%" + strings.Join(terms, "%") + "%"
	rows, err := r.db.Query(
		`SELECT DISTINCT path FROM chunks WHERE lower(content) LIKE ? LIMIT ?`,
		pattern, limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			return paths, err
		}
		paths = append(paths, p)
	}
	return paths, rows.Err()
}

func (r *Retriever) searchSemantic(ctx context.Context, query string, limit int) ([]string, error) {
	qv, err := embedder.EmbedText(ctx, query)
	if err != nil {
		return nil, err
	}
	rows, err := r.db.Query(
		`SELECT path, embedding FROM chunks WHERE embedding IS NOT NULL AND embedding != ''`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type scored struct {
		path  string
		score float64
	}
	bestByPath := make(map[string]float64)
	for rows.Next() {
		var path, embRaw string
		if err := rows.Scan(&path, &embRaw); err != nil {
			return nil, err
		}
		ev := unmarshalEmbedding(embRaw)
		if len(ev) == 0 {
			continue
		}
		score := cosineSimilarity(qv, ev)
		if prev, ok := bestByPath[path]; !ok || score > prev {
			bestByPath[path] = score
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	ranked := make([]scored, 0, len(bestByPath))
	for path, score := range bestByPath {
		ranked = append(ranked, scored{path: path, score: score})
	}
	sort.Slice(ranked, func(i, j int) bool {
		if ranked[i].score == ranked[j].score {
			return ranked[i].path < ranked[j].path
		}
		return ranked[i].score > ranked[j].score
	})
	var paths []string
	for i := 0; i < len(ranked) && i < limit; i++ {
		paths = append(paths, ranked[i].path)
	}
	return paths, nil
}

// HeuristicContextFiles returns paths without an index (dry-run / fallback).
func HeuristicContextFiles(repoRoot, feature string) []string {
	_ = repoRoot
	candidates := []string{
		"docs/ai/active/handoff.md",
		"docs/ai/active/current-spec.md",
		".kiro/specs/" + feature + "/tasks.md",
		"spec.md",
	}
	return candidates
}
