package cloud_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
	"github.com/stretchr/testify/require"
)

func TestFetchConfigBundleHeadersAndBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/v1/projects/p1/config/bundle", r.URL.Path)
		w.Header().Set("ETag", "\"hash123\"")
		w.Header().Set("X-Release-Version", "2")
		w.Header().Set("X-Config-Api-Version", "asagiri.cloud/v1")
		w.Header().Set("X-Registry-Schema-Version", "1")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"contentHash": "hash123",
			"registry": map[string]any{
				"schemaVersion": 1,
			},
		})
	}))
	t.Cleanup(srv.Close)

	client := cloud.NewClient(cloud.ClientOptions{BaseURL: srv.URL, Token: "tok"})
	got, err := client.FetchConfigBundle(context.Background(), "p1", nil)
	require.NoError(t, err)
	require.Equal(t, "hash123", got.ContentHash)
	require.Equal(t, 2, got.ReleaseVersion)
	require.Equal(t, 1, got.RegistrySchema)
}
