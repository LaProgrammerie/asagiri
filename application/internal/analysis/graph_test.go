package analysis_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/stretchr/testify/require"
)

func TestBuildSymbolGraph(t *testing.T) {
	t.Parallel()
	g := analysis.BuildSymbolGraph([]string{"InviteMember", "OnboardingService"})
	require.Len(t, g.Nodes, 2)
}

func TestBuildDependencyGraph(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.go")
	require.NoError(t, os.WriteFile(path, []byte(`package sample
import "fmt"
func Main() { fmt.Println("x") }
`), 0o644))
	g, err := analysis.BuildDependencyGraph(dir, []string{"sample.go"})
	require.NoError(t, err)
	require.NotEmpty(t, g.Nodes)
}
