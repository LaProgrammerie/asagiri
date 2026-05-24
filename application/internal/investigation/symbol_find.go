package investigation

import (
	"path/filepath"
	"strings"
)

// FindSymbolFiles lists .go candidate files declaring symbol.
func FindSymbolFiles(repoRoot string, candidates []string, symbol string) []string {
	var out []string
	want := strings.TrimSpace(symbol)
	if want == "" {
		return out
	}
	for _, rel := range candidates {
		if !strings.HasSuffix(strings.ToLower(rel), ".go") {
			continue
		}
		b, err := ReadFileSnippet(repoRoot, rel, 256*1024)
		if err != nil {
			continue
		}
		for _, s := range ExtractGoSymbols(string(b)) {
			if s == want {
				out = append(out, filepath.ToSlash(rel))
				break
			}
		}
	}
	return out
}
