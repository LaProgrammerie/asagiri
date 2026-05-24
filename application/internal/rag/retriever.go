package rag

import (
	"database/sql"
	"strings"
)

// Retriever finds context files for a task objective.
type Retriever struct {
	db *sql.DB
}

func NewRetriever(db *sql.DB) *Retriever {
	return &Retriever{db: db}
}

// Search returns up to limit paths matching query terms (simple LIKE).
func (r *Retriever) Search(query string, limit int) ([]string, error) {
	if r == nil || r.db == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 8
	}
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
