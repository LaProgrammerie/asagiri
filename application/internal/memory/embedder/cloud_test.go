package embedder_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/stretchr/testify/require"
)

func TestCloudDisabledRejectsEvenWithAPIKey(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "sk-test-key-should-not-be-used")
	_, err := embedder.NewFromConfig(config.RuntimeMemoryConfig{
		Embedder: "cloud",
		Cloud:    config.RuntimeMemoryCloudConfig{Enabled: false},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "disabled")
}

func TestCloudEnabledMockHTTP(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/embeddings", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"data": []map[string]any{{"embedding": []float32{0.5, 0.5}}},
		})
	}))
	t.Cleanup(srv.Close)
	client, err := embedder.NewCloudWithClient(embedder.CloudConfig{
		Enabled: true,
		BaseURL: srv.URL,
		Model:   "text-embedding-3-small",
	}, "token", srv.Client())
	require.NoError(t, err)
	out, err := client.Embed(context.Background(), "hello")
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func TestCloudFactoryRequiresEnabled(t *testing.T) {
	t.Setenv("OPENAI_API_KEY", "x")
	_, err := embedder.NewCloud(embedder.CloudConfig{Enabled: false})
	require.Error(t, err)
}
