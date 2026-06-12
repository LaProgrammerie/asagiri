package extractors_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	"github.com/stretchr/testify/require"
)

func TestAnalyticsExtractorParsesMetricsAndEvents(t *testing.T) {
	repo := t.TempDir()
	product := "demo"
	dir := filepath.Join(repo, ".asagiri", "products", product, "contracts")
	require.NoError(t, os.MkdirAll(dir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "analytics.yaml"), []byte(`events:
  - metric:onboarding_completion_rate
  - dashboard:onboarding_completion_rate
  - event:onboarding.invite_member
`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "observability.yaml"), []byte(`metrics:
  - name: onboarding_completion_rate
`), 0o644))

	nodes, edges, _, err := extractors.AnalyticsExtractor{}.Extract(context.Background(), repo, product)
	require.NoError(t, err)
	ids := nodeIDs(nodes)
	require.Contains(t, ids, "metric:onboarding_completion_rate")
	require.Contains(t, ids, "event:onboarding.invite_member")
	require.NotEmpty(t, edges)
}
