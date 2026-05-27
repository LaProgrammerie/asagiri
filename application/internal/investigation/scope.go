package investigation

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// ResolvedScope narrows investigation perimeter (spec-my-A §25.6).
type ResolvedScope struct {
	Instruction     string   `json:"instruction"`
	Flow            string   `json:"flow,omitempty"`
	Step            string   `json:"step,omitempty"`
	Action          string   `json:"action,omitempty"`
	TaskID          string   `json:"task_id,omitempty"`
	RunID           string   `json:"run_id,omitempty"`
	LikelyModules   []string `json:"likely_modules,omitempty"`
	Contracts       []string `json:"contracts,omitempty"`
	SearchPatterns  []string `json:"search_patterns,omitempty"`
	ChangedFiles    []string `json:"changed_files,omitempty"`
	FailedTestPaths []string `json:"failed_test_paths,omitempty"`
}

// ResolveScope builds scope from request and product artefacts when possible.
func ResolveScope(req Request) ResolvedScope {
	scope := ResolvedScope{
		Instruction: req.Symptom,
		Flow:        req.Flow,
		TaskID:      req.TaskID,
		RunID:       req.RunID,
	}
	patterns := tokenizeSymptom(req.Symptom)
	if req.Feature != "" {
		patterns = append(patterns, req.Feature)
	}
	if req.Flow != "" {
		patterns = append(patterns, req.Flow)
		scope.LikelyModules = append(scope.LikelyModules, "flows/"+req.Flow, "src/"+titleCase(req.Flow))
		scope.Contracts = append(scope.Contracts, req.Flow+".*")
	}
	if req.TaskID != "" {
		patterns = append(patterns, req.TaskID)
		scope.LikelyModules = append(scope.LikelyModules, ".asagiri/tasks/**/"+req.TaskID+"*")
	}
	if req.FromFailedTests && req.RepoRoot != "" {
		paths, err := ParseFailedTests(context.Background(), req.RepoRoot)
		if err != nil || len(paths) == 0 {
			paths = discoverFailedTestHints(req.RepoRoot)
		}
		scope.FailedTestPaths = paths
		scope.LikelyModules = append(scope.LikelyModules, scope.FailedTestPaths...)
	}
	scope.SearchPatterns = dedupeKeepOrder(patterns)
	if req.Flow == "" && len(patterns) > 0 {
		for _, p := range patterns {
			if looksLikeFlowID(p) {
				scope.Flow = p
				break
			}
		}
	}
	if req.Symptom != "" {
		lower := strings.ToLower(req.Symptom)
		if strings.Contains(lower, "invite") {
			scope.Step = "invite_team"
			scope.Action = "invite_member"
			scope.Contracts = append(scope.Contracts, "member.invite", "invitation_email_sent")
		}
	}
	scope.Contracts = dedupeKeepOrder(scope.Contracts)
	scope.LikelyModules = dedupeKeepOrder(scope.LikelyModules)
	return scope
}

func tokenizeSymptom(s string) []string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	var out []string
	for _, w := range strings.FieldsFunc(s, func(r rune) bool {
		return r == ' ' || r == ',' || r == '.' || r == ';' || r == ':' || r == '"' || r == '\''
	}) {
		w = strings.TrimSpace(w)
		if len(w) < 3 {
			continue
		}
		out = append(out, w)
	}
	return dedupeKeepOrder(out)
}

func looksLikeFlowID(s string) bool {
	return strings.Contains(s, "-") || strings.Contains(strings.ToLower(s), "onboarding")
}

func titleCase(s string) string {
	if s == "" {
		return s
	}
	parts := strings.Split(s, "-")
	for i, p := range parts {
		if len(p) == 0 {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, "")
}

func discoverFailedTestHints(repoRoot string) []string {
	// V1: scan common test output locations and *_test.go under application/
	var paths []string
	appTests := filepath.Join(repoRoot, "application")
	_ = filepath.WalkDir(appTests, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, "_test.go") {
			rel, _ := filepath.Rel(repoRoot, path)
			paths = append(paths, filepath.ToSlash(rel))
		}
		return nil
	})
	if len(paths) > 10 {
		paths = paths[:10]
	}
	return paths
}
