package ast

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
)

// FileSymbols holds parsed symbols from a Go file.
type FileSymbols struct {
	Path    string
	Package string
	Funcs   []string
	Types   []string
	Imports []string
}

// ParseGoFile parses a Go source file for package, symbols, and imports.
func ParseGoFile(path string) (FileSymbols, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return FileSymbols{}, err
	}
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, raw, parser.ParseComments)
	if err != nil {
		return FileSymbols{}, err
	}
	out := FileSymbols{Path: path, Package: f.Name.Name}
	for _, imp := range f.Imports {
		if imp.Path != nil {
			p := imp.Path.Value
			if len(p) >= 2 {
				p = p[1 : len(p)-1]
			}
			out.Imports = append(out.Imports, p)
		}
	}
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			if x.Name != nil {
				out.Funcs = append(out.Funcs, x.Name.Name)
			}
		case *ast.TypeSpec:
			if x.Name != nil {
				out.Types = append(out.Types, x.Name.Name)
			}
		}
		return true
	})
	return out, nil
}
