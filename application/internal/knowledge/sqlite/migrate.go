package sqlite

import (
	"embed"
	"fmt"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

func (s *Store) migrate() error {
	raw, err := migrationsFS.ReadFile("migrations/001_knowledge.sql")
	if err != nil {
		return fmt.Errorf("read knowledge migration: %w", err)
	}
	if _, err := s.db.Exec(string(raw)); err != nil {
		return fmt.Errorf("apply knowledge migration: %w", err)
	}
	return nil
}
