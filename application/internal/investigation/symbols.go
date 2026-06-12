package investigation

import (
	"bufio"
	"regexp"
	"strings"
)

var funcDecl = regexp.MustCompile(`(?m)^\s*func\s+(\([^)]+\)\s+)?([A-Za-z0-9_]+)\s*\(`)
var typeDecl = regexp.MustCompile(`(?m)^\s*type\s+([A-Za-z0-9_]+)\s+`)

// ExtractGoSymbols returns simple func/type names found in Go source (specv3 — no tree-sitter).
func ExtractGoSymbols(src string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, m := range funcDecl.FindAllStringSubmatch(src, -1) {
		name := m[2]
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	for _, m := range typeDecl.FindAllStringSubmatch(src, -1) {
		name := m[1]
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		out = append(out, name)
	}
	return out
}

// ParseGoModImports extracts module import paths from go.mod (best-effort, ignores replace).
func ParseGoModImports(goMod string) []string {
	sc := bufio.NewScanner(strings.NewReader(goMod))
	var imports []string
	inRequire := false
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if strings.HasPrefix(line, "require (") {
			inRequire = true
			continue
		}
		if inRequire && line == ")" {
			inRequire = false
			continue
		}
		if strings.HasPrefix(line, "require ") && !strings.Contains(line, "(") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				imports = append(imports, parts[1])
			}
			continue
		}
		if inRequire {
			parts := strings.Fields(line)
			if len(parts) >= 1 && strings.Contains(parts[0], "/") {
				imports = append(imports, parts[0])
			}
		}
	}
	return imports
}
