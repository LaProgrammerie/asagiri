package ast_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/ast"
	"github.com/stretchr/testify/require"
)

func TestParseGoFile(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "x.go")
	require.NoError(t, os.WriteFile(path, []byte(`package x
import "fmt"
func Hello() { fmt.Println("hi") }
type Widget struct{}
`), 0o644))
	parsed, err := ast.ParseGoFile(path)
	require.NoError(t, err)
	require.Equal(t, "x", parsed.Package)
	require.Contains(t, parsed.Funcs, "Hello")
	require.Contains(t, parsed.Types, "Widget")
	require.Contains(t, parsed.Imports, "fmt")
}
