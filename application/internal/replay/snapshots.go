package replay

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Snapshot copies a replay package (re-export for snapshots.go consumers).
func Snapshot(ctx context.Context, req SnapshotRequest) (SnapshotResult, error) {
	return DefaultSnapshotter().Snapshot(ctx, req)
}

// RestoreSnapshot copies a snapshot back into a new replay id directory.
func RestoreSnapshot(repoRoot, snapshotName, replayID string) (string, error) {
	if err := ValidateReplayID(replayID); err != nil {
		return "", err
	}
	src := filepath.Join(repoRoot, RelDir, SnapshotsRelDir, filepath.Base(snapshotName))
	if _, err := os.Stat(src); err != nil {
		return "", fmt.Errorf("replay snapshot restore: %w", err)
	}
	dst := replayDir(repoRoot, replayID)
	if err := copyDir(src, dst); err != nil {
		return "", err
	}
	// refresh manifest id
	pkg, err := LoadPackage(repoRoot, replayID)
	if err == nil {
		pkg.Manifest.ID = replayID
		_ = writeManifest(dst, pkg.Manifest)
	}
	return dst, nil
}

// ListSnapshots returns snapshot names under .asagiri/replays/snapshots/.
func ListSnapshots(repoRoot string) ([]string, error) {
	root := filepath.Join(repoRoot, RelDir, SnapshotsRelDir)
	entries, err := os.ReadDir(root)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	return names, nil
}

// SnapshotPath returns the absolute path for a named snapshot.
func SnapshotPath(repoRoot, name string) string {
	return filepath.Join(repoRoot, RelDir, SnapshotsRelDir, strings.TrimSpace(name))
}
