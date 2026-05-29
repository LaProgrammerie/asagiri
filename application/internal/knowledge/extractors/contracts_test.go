package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
)

func TestContractExtractorOpenAPIPaths(t *testing.T) {
	repo := t.TempDir()
	product := "demo"
	dir := filepath.Join(repo, ".asagiri", "products", product, "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	spec := []byte(`openapi: 3.1.0
paths:
  /invitations:
    post:
      operationId: createInvitation
`)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "api.openapi.yaml"), spec, 0o644))

	nodes, edges, _, err := extractors.ContractExtractor{}.Extract(context.Background(), repo, product)
	require.NoError(t, err)
	ids := nodeIDs(nodes)
	require.Contains(t, ids, "api_operation:POST_invitations")
	require.NotEmpty(t, edges)
}
