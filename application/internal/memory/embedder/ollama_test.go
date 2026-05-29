package embedder_test

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/memory"
	"github.com/LaProgrammerie/asagiri/application/internal/memory/embedder"
	"github.com/stretchr/testify/require"
)

func TestOllamaGoldenSynonymSimilarity(t *testing.T) {
	t.Parallel()
	vLogin := normalizeVec([]float32{1, 0, 0, 0})
	vAuth := normalizeVec([]float32{0.8, 0.6, 0, 0})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Prompt string `json:"prompt"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&req))
		var emb []float32
		switch req.Prompt {
		case "user login failed":
			emb = vLogin
		case "authentication error for user":
			emb = vAuth
		default:
			emb = vLogin
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]any{"embedding": emb}))
	}))
	t.Cleanup(srv.Close)

	client := embedder.NewOllamaWithClient(embedder.OllamaConfig{
		BaseURL: srv.URL,
		Model:   "test-model",
	}, srv.Client())

	a, err := client.Embed(context.Background(), "user login failed")
	require.NoError(t, err)
	b, err := client.Embed(context.Background(), "authentication error for user")
	require.NoError(t, err)
	require.Greater(t, memory.CosineSimilarity(a, b), 0.7)
}

func TestOllamaMockHTTP(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/embeddings", r.URL.Path)
		_ = json.NewEncoder(w).Encode(map[string]any{"embedding": []float32{0.1, 0.2}})
	}))
	t.Cleanup(srv.Close)
	client := embedder.NewOllamaWithClient(embedder.OllamaConfig{BaseURL: srv.URL, Model: "m"}, srv.Client())
	out, err := client.Embed(context.Background(), "hello")
	require.NoError(t, err)
	require.Len(t, out, 2)
}

func normalizeVec(v []float32) []float32 {
	var norm float64
	for _, x := range v {
		norm += float64(x * x)
	}
	if norm == 0 {
		return v
	}
	norm = math.Sqrt(norm)
	out := make([]float32, len(v))
	for i, x := range v {
		out[i] = float32(float64(x) / norm)
	}
	return out
}
