package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "modernc.org/sqlite"
)

const driverName = "sqlite"

// Store implements knowledge.GraphStore with SQLite.
type Store struct {
	repoRoot string
	db       *sql.DB
}

// Open creates or opens the knowledge graph database under repoRoot.
func Open(repoRoot string) (*Store, error) {
	if repoRoot == "" {
		return nil, fmt.Errorf("open knowledge store: repo root required")
	}
	if strings.Contains(repoRoot, "..") {
		return nil, fmt.Errorf("open knowledge store: parent segments not allowed in repo root")
	}
	dir := filepath.Join(repoRoot, knowledge.KnowledgeRelDir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create knowledge dir: %w", err)
	}
	dbPath := filepath.Join(dir, knowledge.GraphDBName)
	db, err := sql.Open(driverName, dbPath)
	if err != nil {
		return nil, fmt.Errorf("open knowledge db: %w", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("ping knowledge db: %w", err)
	}
	s := &Store{repoRoot: repoRoot, db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// RepoRoot returns the repository root.
func (s *Store) RepoRoot() string { return s.repoRoot }

// DBPath returns the SQLite file path.
func (s *Store) DBPath() string {
	return knowledge.DBPath(s.repoRoot)
}

// Close closes the database.
func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}
