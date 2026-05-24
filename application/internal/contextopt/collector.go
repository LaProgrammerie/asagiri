package contextopt

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

type CollectOpts struct {
	MaxFiles int
	MaxBytes int64
}

// Collect walks repo root, feature-related paths, and configured spec paths (specv3 §8.3).
func Collect(repoRoot, feature string, cfg *config.Config, opts CollectOpts) ([]FileEntry, error) {
	if cfg == nil {
		return nil, fmt.Errorf("contextopt: config nil")
	}
	if opts.MaxFiles <= 0 {
		opts.MaxFiles = 500
	}
	if opts.MaxBytes <= 0 {
		opts.MaxBytes = 2 << 20
	}
	seen := map[string]struct{}{}
	var paths []string
	paths = append(paths, ".")
	if feature != "" {
		paths = append(paths,
			filepath.Join(".kiro", "specs", feature),
			filepath.Join(".asagiri", "specs", feature),
		)
	}
	for _, p := range cfg.Sources.Local.Paths {
		paths = append(paths, p)
	}
	// de-dup path list
	uniqPaths := make([]string, 0)
	up := map[string]struct{}{}
	for _, p := range paths {
		c := filepath.Clean(p)
		if _, ok := up[c]; ok {
			continue
		}
		up[c] = struct{}{}
		uniqPaths = append(uniqPaths, c)
	}

	var out []FileEntry
	var total int64
	for _, rel := range uniqPaths {
		abs := filepath.Join(repoRoot, rel)
		entries, err := readTreeFiles(abs, rel, seen, repoRoot, opts.MaxFiles-len(out), opts.MaxBytes-total)
		if err != nil {
			continue // skip missing optional dirs
		}
		for _, e := range entries {
			out = append(out, e)
			total += e.Size
			if len(out) >= opts.MaxFiles || total >= opts.MaxBytes {
				return out, nil
			}
		}
	}
	return out, nil
}

func readTreeFiles(absDir, relPrefix string, seen map[string]struct{}, repoRoot string, maxFiles int, budget int64) ([]FileEntry, error) {
	if maxFiles <= 0 || budget <= 0 {
		return nil, nil
	}
	// single file
	if st, err := statFile(absDir); err == nil && !st.IsDir() {
		return readOneFile(absDir, relPrefix, seen, repoRoot, budget)
	}
	var out []FileEntry
	err := walkShallow(absDir, repoRoot, func(path, rel string, info fs.FileInfo) error {
		if len(out) >= maxFiles {
			return errStopWalk
		}
		if info.IsDir() {
			if skipDir(rel) {
				return filepath.SkipDir
			}
			return nil
		}
		if !includeFile(rel) {
			return nil
		}
		entries, err := readOneFile(path, rel, seen, repoRoot, budget-int64(totalSize(out)))
		if err != nil || len(entries) == 0 {
			return nil
		}
		out = append(out, entries...)
		return nil
	})
	if err != nil && err != errStopWalk {
		return out, err
	}
	return out, nil
}

func totalSize(entries []FileEntry) int64 {
	var n int64
	for _, e := range entries {
		n += e.Size
	}
	return n
}

func readOneFile(abs, rel string, seen map[string]struct{}, repoRoot string, budget int64) ([]FileEntry, error) {
	rel = filepath.ToSlash(rel)
	if _, ok := seen[rel]; ok {
		return nil, nil
	}
	b, info, err := readFileLimited(abs, budget)
	if err != nil {
		return nil, err
	}
	seen[rel] = struct{}{}
	return []FileEntry{{
		Path:     abs,
		RelPath:  rel,
		Size:     info.Size(),
		Content:  string(b),
		Language: classifyPathKind(rel),
	}}, nil
}

func includeFile(rel string) bool {
	ext := strings.ToLower(filepath.Ext(rel))
	switch ext {
	case ".go", ".md", ".yaml", ".yml", ".json", ".mod", ".sum", ".sql", ".php", ".ts", ".tsx", ".js", ".jsx", ".rs", ".toml":
		return true
	default:
		return false
	}
}

func skipDir(rel string) bool {
	parts := strings.Split(filepath.ToSlash(rel), "/")
	for _, p := range parts {
		switch p {
		case "node_modules", "vendor", ".git", "dist", "build", ".asagiri", "terminals":
			return true
		}
	}
	return false
}

func classifyPathKind(rel string) ContentKind {
	switch strings.ToLower(filepath.Ext(rel)) {
	case ".go", ".rs", ".php", ".ts", ".tsx", ".js", ".jsx":
		return KindCode
	case ".md", ".mdx":
		return KindMarkdown
	case ".json", ".yaml", ".yml":
		return KindJSON
	default:
		return KindDefault
	}
}

var errStopWalk = fmt.Errorf("stop walk")
