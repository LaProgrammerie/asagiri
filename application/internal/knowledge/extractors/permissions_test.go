package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
)

func TestPermissionExtractorParsesList(t *testing.T) {
	repo := t.TempDir()
	product := "demo"
	dir := filepath.Join(repo, ".asagiri", "products", product, "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "permissions.yaml"), []byte(`permissions:
  - id: workspace.invite
`), 0o644))

	nodes, _, _, err := extractors.PermissionExtractor{}.Extract(context.Background(), repo, product)
	require.NoError(t, err)
	ids := nodeIDs(nodes)
	require.Contains(t, ids, "permission:workspace.invite")
}
