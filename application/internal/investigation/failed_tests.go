package investigation

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type goTestEvent struct {
	Action  string `json:"Action"`
	Package string `json:"Package"`
	Test    string `json:"Test"`
}

// ParseFailedTests runs `go test -json` and returns repo-relative paths to failing test files.
func ParseFailedTests(ctx context.Context, repoRoot string) ([]string, error) {
	modDir := filepath.Join(repoRoot, "application")
	if st, err := os.Stat(modDir); err != nil || !st.IsDir() {
		modDir = repoRoot
	}
	runCtx, cancel := context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(runCtx, "go", "test", "-json", "./...")
	cmd.Dir = modDir
	out, _ := cmd.CombinedOutput()

	var failedPackages []string
	sc := bufio.NewScanner(bytes.NewReader(out))
	for sc.Scan() {
		var ev goTestEvent
		if json.Unmarshal(sc.Bytes(), &ev) != nil {
			continue
		}
		if ev.Action == "fail" && ev.Package != "" {
			failedPackages = append(failedPackages, ev.Package)
		}
	}

	var files []string
	seen := map[string]struct{}{}
	for _, pkg := range dedupeKeepOrder(failedPackages) {
		rel := packageToTestFile(repoRoot, modDir, pkg)
		if rel == "" {
			continue
		}
		if _, ok := seen[rel]; ok {
			continue
		}
		seen[rel] = struct{}{}
		files = append(files, rel)
	}
	return files, nil
}

func packageToTestFile(repoRoot, modDir, pkg string) string {
	pkg = strings.TrimPrefix(pkg, "github.com/LaProgrammerie/asagiri/application/")
	if pkg == "" {
		return ""
	}
	candidate := filepath.Join(modDir, filepath.FromSlash(pkg))
	entries, err := os.ReadDir(candidate)
	if err != nil {
		return ""
	}
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), "_test.go") {
			rel, err := filepath.Rel(repoRoot, filepath.Join(candidate, e.Name()))
			if err != nil {
				return ""
			}
			return filepath.ToSlash(rel)
		}
	}
	return ""
}
