package rag

import (
	"database/sql"
)

func initSchema(db *sql.DB) error {
	if _, err := db.Exec(`
CREATE TABLE IF NOT EXISTS chunks (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  path TEXT NOT NULL,
  offset INTEGER NOT NULL,
  content TEXT NOT NULL,
  embedding TEXT
);
CREATE INDEX IF NOT EXISTS idx_chunks_path ON chunks(path);
`); err != nil {
		return err
	}
	return migrateChunksEmbedding(db)
}

func migrateChunksEmbedding(db *sql.DB) error {
	rows, err := db.Query(`PRAGMA table_info(chunks)`)
	if err != nil {
		return err
	}
	defer func() { _ = rows.Close() }()
	hasEmbedding := false
	for rows.Next() {
		var cid int
		var name, ctype string
		var notnull, pk int
		var dflt sql.NullString
		if err := rows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk); err != nil {
			return err
		}
		if name == "embedding" {
			hasEmbedding = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	if hasEmbedding {
		return nil
	}
	_, err = db.Exec(`ALTER TABLE chunks ADD COLUMN embedding TEXT`)
	return err
}

func indexHasEmbeddings(db *sql.DB) (bool, error) {
	var n int
	err := db.QueryRow(
		`SELECT COUNT(1) FROM chunks WHERE embedding IS NOT NULL AND embedding != ''`,
	).Scan(&n)
	return n > 0, err
}
