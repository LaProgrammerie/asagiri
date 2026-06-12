package flows_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis/flows"
	"github.com/stretchr/testify/require"
)

func TestBuildFlowGraph(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "onboarding.flow.yaml"), []byte(`id: onboarding
title: Onboarding
steps:
  - id: s1
    screen: landing
  - id: s2
    screen: signup
`), 0o644))
	g, err := flows.Build(dir)
	require.NoError(t, err)
	require.NotEmpty(t, g.Nodes)
	require.NotEmpty(t, g.Edges)
}
