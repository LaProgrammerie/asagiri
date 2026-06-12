package investigation

import (
	"go/parser"
	"go/token"
	"os"
)

// ExtractImports returns import paths from a Go source file.
func ExtractImports(path string) ([]string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, raw, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	var out []string
	for _, imp := range f.Imports {
		if imp == nil || imp.Path == nil {
			continue
		}
		p := imp.Path.Value
		if len(p) >= 2 {
			p = p[1 : len(p)-1]
		}
		out = append(out, p)
	}
	return out, nil
}
