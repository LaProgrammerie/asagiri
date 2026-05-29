package investigation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

// ContextPack is a bounded agent context (spec-my-A §25, spec-my-E §15).
type ContextPack struct {
	Files             []string `json:"files"`
	Tests             []string `json:"tests"`
	Symbols           []string `json:"symbols,omitempty"`
	Contracts         []string `json:"contracts,omitempty"`
	APIs              []string `json:"apis,omitempty"`
	Events            []string `json:"events,omitempty"`
	Metrics           []string `json:"metrics,omitempty"`
	Flows             []string `json:"flows,omitempty"`
	Risks             []string `json:"risks,omitempty"`
	ExcludedSensitive []string `json:"excluded_sensitive,omitempty"`
	MaxFiles          int      `json:"max_files"`
}

// BuildContextPack selects minimal files excluding sensitive paths.
func BuildContextPack(repoRoot string, res InvestigationResult, scope ResolvedScope, maxFiles int) (ContextPack, error) {
	if maxFiles <= 0 {
		maxFiles = 80
	}
	sensitive := map[string]struct{}{}
	for _, p := range res.SensitivePaths {
		sensitive[p] = struct{}{}
	}
	var files []string
	for _, f := range res.CandidateFiles {
		if isSensitivePath(f, sensitive) {
			continue
		}
		if strings.Contains(strings.ToLower(f), ".env") {
			continue
		}
		files = append(files, f)
		if len(files) >= maxFiles {
			break
		}
	}
	pack := ContextPack{
		Files:             files,
		Tests:             res.RelatedTests,
		Symbols:           trimSlice(res.Symbols, 50),
		ExcludedSensitive: res.SensitivePaths,
		MaxFiles:          maxFiles,
	}
	return pack, nil
}

// WriteContextPack saves context-pack.json next to the report.
func WriteContextPack(dir string, pack ContextPack) (string, error) {
	path := filepath.Join(dir, "context-pack.json")
	raw, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, raw, 0o644)
}

func isSensitivePath(path string, set map[string]struct{}) bool {
	if _, ok := set[path]; ok {
		return true
	}
	lower := strings.ToLower(path)
	for _, frag := range []string{".env", "secret", "credential", "private_key", "id_rsa"} {
		if strings.Contains(lower, frag) {
			return true
		}
	}
	return false
}

func trimSlice(in []string, n int) []string {
	if len(in) <= n {
		return in
	}
	return in[:n]
}
