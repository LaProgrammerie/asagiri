package reportsink

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// SaveOptions controls optional history retention on save.
type SaveOptions struct {
	// KeepHistory archives the previous latest snapshot before overwrite (default true).
	KeepHistory bool
}

// DefaultSaveOptions enables history archival.
func DefaultSaveOptions() SaveOptions {
	return SaveOptions{KeepHistory: true}
}

// HistoryEntry is one timestamped snapshot under a history/ directory.
type HistoryEntry struct {
	RelPath string    `json:"rel_path"`
	AbsPath string    `json:"abs_path"`
	ModTime time.Time `json:"mod_time"`
}

func archiveBeforeOverwrite(repoRoot, relUnderRuntime, absLatest string, opts SaveOptions) error {
	if !opts.KeepHistory {
		return nil
	}
	if _, err := os.Stat(absLatest); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reports: stat latest: %w", err)
	}
	historyRel := historyDirRel(relUnderRuntime)
	historyAbs := filepath.Join(repoRoot, runtimeDir, historyRel)
	if err := os.MkdirAll(historyAbs, 0o755); err != nil {
		return fmt.Errorf("reports: mkdir history: %w", err)
	}
	name := archiveFileName(relUnderRuntime, time.Now().UTC())
	dst := filepath.Join(historyAbs, name)
	return copyFile(absLatest, dst)
}

func historyDirRel(relUnderRuntime string) string {
	dir := filepath.Dir(relUnderRuntime)
	return filepath.Join(dir, "history")
}

func archiveFileName(relUnderRuntime string, at time.Time) string {
	base := strings.TrimSuffix(filepath.Base(relUnderRuntime), ".json")
	ts := at.UTC().Format("20060102T150405Z")
	if base == "latest" {
		return ts + ".json"
	}
	return base + "_" + ts + ".json"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("reports: open %s: %w", src, err)
	}
	defer func() { _ = in.Close() }()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return fmt.Errorf("reports: create history %s: %w", dst, err)
	}
	defer func() {
		_ = out.Close()
	}()
	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("reports: copy history: %w", err)
	}
	return out.Close()
}

// ListHistory returns history entries newest-first for a stable report path.
func ListHistory(repoRoot, relUnderRuntime string) ([]HistoryEntry, error) {
	if err := RequireInitialized(repoRoot); err != nil {
		return nil, err
	}
	historyRel := historyDirRel(relUnderRuntime)
	historyAbs := filepath.Join(repoRoot, runtimeDir, historyRel)
	entries, err := os.ReadDir(historyAbs)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("reports: read history: %w", err)
	}
	prefix := historyPrefix(relUnderRuntime)
	out := make([]HistoryEntry, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		if prefix != "" && !strings.HasPrefix(e.Name(), prefix) {
			continue
		}
		abs := filepath.Join(historyAbs, e.Name())
		info, err := e.Info()
		if err != nil {
			continue
		}
		out = append(out, HistoryEntry{
			RelPath: relRepo(repoRoot, abs),
			AbsPath: abs,
			ModTime: info.ModTime(),
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].ModTime.After(out[j].ModTime)
	})
	return out, nil
}

func historyPrefix(relUnderRuntime string) string {
	base := strings.TrimSuffix(filepath.Base(relUnderRuntime), ".json")
	if base == "latest" {
		return ""
	}
	return base + "_"
}

// DiffPairPaths resolves before/after paths for diff: latest vs previous history entry.
func DiffPairPaths(repoRoot, relUnderRuntime string) (beforeRel, afterRel string, err error) {
	if err := RequireInitialized(repoRoot); err != nil {
		return "", "", err
	}
	afterAbs := filepath.Join(repoRoot, runtimeDir, relUnderRuntime)
	if _, err := os.Stat(afterAbs); err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("reports: aucun snapshot latest — lancez avec --save d'abord")
		}
		return "", "", fmt.Errorf("reports: stat latest: %w", err)
	}
	afterRel = relRepo(repoRoot, afterAbs)
	history, err := ListHistory(repoRoot, relUnderRuntime)
	if err != nil {
		return "", "", err
	}
	if len(history) == 0 {
		return "", "", fmt.Errorf("reports: historique insuffisant — au moins deux --save requis pour diff")
	}
	return history[0].RelPath, afterRel, nil
}
