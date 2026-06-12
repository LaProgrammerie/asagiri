package investigation

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// ScanRepo walks a bounded tree listing candidate source files and large files.
func ScanRepo(repoRoot string, invCfg config.InvestigationConfig, maxFiles int) (candidates, large []string, err error) {
	if maxFiles <= 0 {
		maxFiles = 400
	}
	limitBytes := invCfg.LargeFileBytes
	if limitBytes <= 0 {
		limitBytes = 512 * 1024
	}
	err = filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, e := filepath.Rel(repoRoot, path)
		if e != nil {
			return e
		}
		if d.IsDir() {
			if shouldSkipDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !isCandidateFile(rel) {
			return nil
		}
		if len(candidates) >= maxFiles {
			return nil
		}
		candidates = append(candidates, filepath.ToSlash(rel))
		info, err := d.Info()
		if err == nil && info.Size() >= limitBytes {
			large = append(large, filepath.ToSlash(rel))
		}
		return nil
	})
	return candidates, large, err
}

func shouldSkipDir(rel string) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, p := range parts {
		switch p {
		case ".git", "node_modules", "vendor", "dist", "build":
			return true
		}
	}
	return false
}

func isCandidateFile(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	switch ext {
	case ".go", ".php", ".ts", ".tsx", ".js", ".rs", ".py", ".md":
		return true
	default:
		return false
	}
}

// FindSensitivePaths lists files whose names match denylist or basic patterns.
func FindSensitivePaths(repoRoot string, deny []string) ([]string, error) {
	if len(deny) == 0 {
		deny = []string{".env", "credentials.json"}
	}
	var out []string
	err := filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return nil
		}
		rel, e := filepath.Rel(repoRoot, path)
		if e != nil {
			return e
		}
		rel = filepath.ToSlash(rel)
		base := filepath.Base(rel)
		for _, p := range deny {
			if p == "" {
				continue
			}
			if strings.EqualFold(base, p) || strings.Contains(rel, p) {
				out = append(out, rel)
				break
			}
		}
		return nil
	})
	return out, err
}

// ReadFileSnippet reads up to maxBytes from a repo file (must stay under repoRoot).
func ReadFileSnippet(repoRoot, rel string, maxBytes int) ([]byte, error) {
	if maxBytes <= 0 {
		maxBytes = 64 * 1024
	}
	clean := filepath.Clean(rel)
	abs := filepath.Join(repoRoot, clean)
	if !strings.HasPrefix(abs, filepath.Clean(repoRoot)) {
		return nil, fs.ErrPermission
	}
	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}
	if len(b) > maxBytes {
		b = b[:maxBytes]
	}
	return b, nil
}
