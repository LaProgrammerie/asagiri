package knowledge

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// KnowledgeRelDir is the knowledge artefact root under a repository.
const KnowledgeRelDir = ".asagiri/knowledge"

// GraphDBName is the SQLite database file name.
const GraphDBName = "graph.sqlite"

// GraphJSONName is the exported graph snapshot file name.
const GraphJSONName = "graph.json"

// GraphStore persists and queries the engineering knowledge graph.
type GraphStore interface {
	UpsertNode(ctx context.Context, node GraphNode) error
	UpsertEdge(ctx context.Context, edge GraphEdge) error
	GetNode(ctx context.Context, id string) (GraphNode, error)
	GetEdge(ctx context.Context, id string) (GraphEdge, error)
	ListNodes(ctx context.Context, filter NodeFilter) ([]GraphNode, error)
	ListEdges(ctx context.Context, filter EdgeFilter) ([]GraphEdge, error)
	LoadGraph(ctx context.Context) (KnowledgeGraph, error)
	SetIndexMetadata(ctx context.Context, key string, value any) error
	GetIndexMetadata(ctx context.Context, key string) (map[string]any, error)
	Close() error
}

// NodeFilter narrows node listing.
type NodeFilter struct {
	ID       string
	Type     NodeType
	Path     string
	PathLike string
}

// EdgeFilter narrows edge listing.
type EdgeFilter struct {
	ID         string
	Type       EdgeType
	FromNodeID string
	ToNodeID   string
}

var sqliteStoreOpener func(string) (GraphStore, error)

// RegisterSQLiteStore wires the SQLite backend (called from knowledge/sqlite init).
func RegisterSQLiteStore(opener func(string) (GraphStore, error)) {
	sqliteStoreOpener = opener
}

// OpenStore opens the knowledge graph SQLite store under repoRoot/.asagiri/knowledge/.
func OpenStore(repoRoot string) (GraphStore, error) {
	if sqliteStoreOpener == nil {
		return nil, fmt.Errorf("open knowledge store: sqlite backend not registered")
	}
	return sqliteStoreOpener(repoRoot)
}

// OpenStoreIfExists opens the store only when graph.sqlite is present at the repo root.
func OpenStoreIfExists(repoRoot string) (GraphStore, error) {
	if _, err := os.Stat(DBPath(repoRoot)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return OpenStore(repoRoot)
}

// DBPath returns the absolute SQLite path for a repository root.
func DBPath(repoRoot string) string {
	return filepath.Join(repoRoot, KnowledgeRelDir, GraphDBName)
}

// JSONPath returns the absolute JSON export path for a repository root.
func JSONPath(repoRoot string) string {
	return filepath.Join(repoRoot, KnowledgeRelDir, GraphJSONName)
}
