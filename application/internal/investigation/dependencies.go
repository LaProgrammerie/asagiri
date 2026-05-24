package investigation

import (
	"os"
	"path/filepath"
	"strings"
)

// ParseGoModDependencies returns import paths from repo go.mod when present.
func ParseGoModDependencies(repoRoot string) ([]string, error) {
	p := filepath.Join(repoRoot, "go.mod")
	b, err := os.ReadFile(p)
	if err != nil {
		return nil, err
	}
	return ParseGoModImports(string(b)), nil
}

// RelatedTestPaths guesses *_test.go files sharing basename with candidate .go files.
func RelatedTestPaths(candidates []string) []string {
	var out []string
	seen := map[string]struct{}{}
	for _, c := range candidates {
		if !strings.HasSuffix(c, ".go") || strings.HasSuffix(c, "_test.go") {
			continue
		}
		base := strings.TrimSuffix(filepath.Base(c), ".go")
		dir := filepath.Dir(c)
		cand := filepath.ToSlash(filepath.Join(dir, base+"_test.go"))
		if _, ok := seen[cand]; ok {
			continue
		}
		seen[cand] = struct{}{}
		out = append(out, cand)
	}
	return out
}
