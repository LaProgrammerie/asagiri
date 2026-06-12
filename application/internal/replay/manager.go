package replay

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReplayManager creates and runs replay packages (spec §25).
type ReplayManager interface {
	Create(ctx context.Context, req ReplayCreateRequest) (ReplayPackage, error)
	Run(ctx context.Context, req ReplayRunRequest) (ReplayResult, error)
}

// ReplayComparator compares two replay packages (spec §25).
type ReplayComparator interface {
	Compare(ctx context.Context, a string, b string) (ReplayComparison, error)
}

// Manager is the default ReplayManager implementation.
type Manager struct {
	RepoRoot string
	Policies CapturePolicies
}

// DefaultManager returns a manager for repoRoot with config-derived policies.
func DefaultManager(repoRoot string, policies CapturePolicies) *Manager {
	return &Manager{RepoRoot: repoRoot, Policies: policies}
}

// Create captures a replay package.
func (m *Manager) Create(ctx context.Context, req ReplayCreateRequest) (ReplayPackage, error) {
	if req.RepoRoot == "" {
		req.RepoRoot = m.RepoRoot
	}
	if req.Config == (CapturePolicies{}) {
		req.Config = m.Policies
	}
	return CapturePackage(ctx, req)
}

// Run executes a replay session.
func (m *Manager) Run(ctx context.Context, req ReplayRunRequest) (ReplayResult, error) {
	if req.RepoRoot == "" {
		req.RepoRoot = m.RepoRoot
	}
	return ExecuteReplay(ctx, req)
}

// comparatorImpl implements ReplayComparator for a repo.
type comparatorImpl struct {
	repoRoot string
}

// NewComparator returns a ReplayComparator bound to repoRoot.
func NewComparator(repoRoot string) ReplayComparator {
	return &comparatorImpl{repoRoot: repoRoot}
}

// Compare compares two replay IDs.
func (c *comparatorImpl) Compare(ctx context.Context, a, b string) (ReplayComparison, error) {
	return DefaultComparator().Compare(ctx, c.repoRoot, a, b)
}

// SnapshotRequest names a replay snapshot (spec §21).
type SnapshotRequest struct {
	RepoRoot string
	ReplayID string
	Name     string
}

// SnapshotResult describes a persisted snapshot.
type SnapshotResult struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Path string `json:"path"`
}

// Snapshotter captures replay snapshots.
type Snapshotter interface {
	Snapshot(ctx context.Context, req SnapshotRequest) (SnapshotResult, error)
}

// PackageSnapshotter copies a replay package into snapshots/<name>/.
type PackageSnapshotter struct{}

// DefaultSnapshotter returns the standard snapshotter.
func DefaultSnapshotter() *PackageSnapshotter {
	return &PackageSnapshotter{}
}

// Snapshot copies replay artefacts to .asagiri/replays/snapshots/<name>/ (spec §21).
func (s *PackageSnapshotter) Snapshot(ctx context.Context, req SnapshotRequest) (SnapshotResult, error) {
	_ = ctx
	name := strings.TrimSpace(req.Name)
	if name == "" || name == "." || name == ".." {
		return SnapshotResult{}, ErrSnapshotName
	}
	if err := ValidateReplayID(req.ReplayID); err != nil {
		return SnapshotResult{}, err
	}
	src, err := LoadPackage(req.RepoRoot, req.ReplayID)
	if err != nil {
		return SnapshotResult{}, err
	}
	dst := filepath.Join(req.RepoRoot, RelDir, SnapshotsRelDir, filepath.Base(name))
	if err := os.MkdirAll(dst, 0o755); err != nil {
		return SnapshotResult{}, fmt.Errorf("replay snapshot dir: %w", err)
	}
	if err := copyDir(src.Path, dst); err != nil {
		return SnapshotResult{}, err
	}
	return SnapshotResult{
		ID:   req.ReplayID,
		Name: name,
		Path: dst,
	}, nil
}

func copyDir(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		body, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, body, 0o644)
	})
}
