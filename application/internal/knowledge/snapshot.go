package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// SnapshotsRelDir is the snapshots subdirectory under knowledge artefacts.
const SnapshotsRelDir = "snapshots"

// SnapshotRequest names a graph snapshot.
type SnapshotRequest struct {
	RepoRoot string
	Name     string
}

// SnapshotResult describes a persisted snapshot.
type SnapshotResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// Snapshotter captures graph snapshots (spec-my-E §6.5).
type Snapshotter interface {
	Snapshot(ctx context.Context, req SnapshotRequest) (SnapshotResult, error)
}

// GraphSnapshotter copies build metadata and graph.json into snapshots/<name>/.
type GraphSnapshotter struct {
	OpenStore func(string) (GraphStore, error)
}

// DefaultSnapshotter returns a snapshotter using the registered SQLite store.
func DefaultSnapshotter() *GraphSnapshotter {
	return &GraphSnapshotter{OpenStore: OpenStore}
}

// Snapshot persists metadata and graph.json under .asagiri/knowledge/snapshots/<name>/.
func (s *GraphSnapshotter) Snapshot(ctx context.Context, req SnapshotRequest) (SnapshotResult, error) {
	name := filepath.Base(strings.TrimSpace(req.Name))
	if name == "" || name == "." || name == ".." {
		return SnapshotResult{}, fmt.Errorf("knowledge snapshot: name required")
	}
	if req.RepoRoot == "" {
		return SnapshotResult{}, fmt.Errorf("knowledge snapshot: repo root required")
	}

	open := s.OpenStore
	if open == nil {
		open = OpenStore
	}
	store, err := open(req.RepoRoot)
	if err != nil {
		return SnapshotResult{}, err
	}
	defer func() { _ = store.Close() }()

	buildMeta, err := store.GetIndexMetadata(ctx, "build")
	if err != nil {
		return SnapshotResult{}, fmt.Errorf("knowledge snapshot: %w", err)
	}

	snapDir := filepath.Join(req.RepoRoot, KnowledgeRelDir, SnapshotsRelDir, name)
	if err := os.MkdirAll(snapDir, 0o755); err != nil {
		return SnapshotResult{}, fmt.Errorf("knowledge snapshot dir: %w", err)
	}

	metaPath := filepath.Join(snapDir, "metadata.json")
	metaBody, err := json.MarshalIndent(buildMeta, "", "  ")
	if err != nil {
		return SnapshotResult{}, err
	}
	if err := os.WriteFile(metaPath, append(metaBody, '\n'), 0o644); err != nil {
		return SnapshotResult{}, fmt.Errorf("write snapshot metadata: %w", err)
	}

	srcJSON := JSONPath(req.RepoRoot)
	dstJSON := filepath.Join(snapDir, GraphJSONName)
	if body, err := os.ReadFile(srcJSON); err == nil {
		if err := os.WriteFile(dstJSON, body, 0o644); err != nil {
			return SnapshotResult{}, fmt.Errorf("copy graph.json: %w", err)
		}
	} else if !os.IsNotExist(err) {
		return SnapshotResult{}, fmt.Errorf("read graph.json: %w", err)
	}

	id := uuid.NewString()
	record := map[string]any{
		"id":         id,
		"name":       name,
		"created_at": time.Now().UTC().Format(time.RFC3339),
		"path":       filepath.ToSlash(filepath.Join(KnowledgeRelDir, SnapshotsRelDir, name)),
	}
	if err := store.SetIndexMetadata(ctx, snapshotMetaKey(name), record); err != nil {
		return SnapshotResult{}, err
	}

	rel := filepath.ToSlash(filepath.Join(KnowledgeRelDir, SnapshotsRelDir, name))
	return SnapshotResult{ID: id, Name: name, Path: rel}, nil
}

func snapshotMetaKey(name string) string {
	return "snapshot:" + name
}

// StubSnapshotter is a deprecated placeholder.
type StubSnapshotter struct{}

func (StubSnapshotter) Snapshot(_ context.Context, _ SnapshotRequest) (SnapshotResult, error) {
	return SnapshotResult{}, ErrNotImplemented
}
