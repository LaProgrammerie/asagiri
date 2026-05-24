package investigation

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Run performs a bounded local investigation pass (specv3 §9).
func Run(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (InvestigationResult, error) {
	var res InvestigationResult
	if cfg == nil {
		return res, fmt.Errorf("investigation: config nil")
	}
	inv := cfg.MCP.Investigation
	pattern := feature
	if pattern == "" {
		pattern = taskID
	}
	if pattern != "" {
		hits, err := Grep(ctx, repoRoot, pattern, inv)
		if err != nil {
			res.Errors = append(res.Errors, err.Error())
		} else {
			res.GrepHits = hits
			for _, h := range hits {
				if idx := strings.Index(h, ":"); idx > 0 {
					res.CandidateFiles = append(res.CandidateFiles, filepath.ToSlash(h[:idx]))
				}
			}
		}
	}
	cand, large, err := ScanRepo(repoRoot, inv, 500)
	if err != nil {
		res.Errors = append(res.Errors, err.Error())
	} else {
		res.CandidateFiles = uniqMerge(res.CandidateFiles, cand)
		res.LargeFiles = uniqMerge(res.LargeFiles, large)
	}
	sens, err := FindSensitivePaths(repoRoot, cfg.MCP.SecretPathDenylist)
	if err != nil {
		res.Errors = append(res.Errors, err.Error())
	} else {
		res.SensitivePaths = sens
	}
	res.RelatedTests = RelatedTestPaths(res.CandidateFiles)
	for _, rel := range res.CandidateFiles {
		if !strings.HasSuffix(strings.ToLower(rel), ".go") {
			continue
		}
		b, err := ReadFileSnippet(repoRoot, rel, int(inv.LargeFileBytes))
		if err != nil {
			continue
		}
		syms := ExtractGoSymbols(string(b))
		res.Symbols = append(res.Symbols, syms...)
		if len(res.Symbols) > 200 {
			break
		}
	}
	if mods, err := ParseGoModDependencies(repoRoot); err == nil {
		res.Imports = mods
	}
	res.CandidateFiles = dedupeKeepOrder(res.CandidateFiles)
	return res, nil
}

func uniqMerge(a, b []string) []string {
	return dedupeKeepOrder(append(append([]string{}, a...), b...))
}

func dedupeKeepOrder(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}
