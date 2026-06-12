package extractors

import (
	"io/fs"
	"path/filepath"
	"strings"
)

func shouldSkipCodeDir(name string) bool {
	switch name {
	case "vendor", "node_modules", ".git", ".asagiri":
		return true
	default:
		return false
	}
}

func shouldSkipCodePath(rel string) bool {
	rel = filepath.ToSlash(rel)
	if strings.Contains(rel, "/vendor/") || strings.Contains(rel, "/testdata/") {
		return true
	}
	if strings.HasPrefix(rel, "testdata/") {
		return true
	}
	return false
}

func walkGoFiles(repoRoot string, testFiles bool, walkFn func(rel string) error) error {
	return filepath.WalkDir(repoRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if shouldSkipCodeDir(d.Name()) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") {
			return nil
		}
		isTest := strings.HasSuffix(d.Name(), "_test.go")
		if testFiles != isTest {
			return nil
		}
		rel, err := filepath.Rel(repoRoot, path)
		if err != nil {
			return err
		}
		rel = filepath.ToSlash(rel)
		if shouldSkipCodePath(rel) {
			return nil
		}
		return walkFn(rel)
	})
}

func pathToStableKey(rel string) string {
	rel = filepath.ToSlash(rel)
	rel = strings.TrimSuffix(rel, ".go")
	rel = strings.ReplaceAll(rel, "/", "_")
	rel = strings.ReplaceAll(rel, ".", "_")
	rel = strings.ReplaceAll(rel, "-", "_")
	return sanitizeStableKey(rel)
}
