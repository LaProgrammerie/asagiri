package sqlite

import "github.com/LaProgrammerie/asagiri/application/internal/knowledge"

func init() {
	knowledge.RegisterSQLiteStore(func(repoRoot string) (knowledge.GraphStore, error) {
		return Open(repoRoot)
	})
}
