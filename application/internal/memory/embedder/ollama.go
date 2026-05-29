package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const ollamaName = "ollama"

// OllamaConfig tunes the local Ollama embeddings endpoint.
type OllamaConfig struct {
	BaseURL string
	Model   string
}

// OllamaEmbedder calls POST /api/embeddings on Ollama.
type OllamaEmbedder struct {
	baseURL string
	model   string
	client  *http.Client
}

// NewOllama builds an Ollama embedder.
func NewOllama(cfg OllamaConfig) *OllamaEmbedder {
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = "http://127.0.0.1:11434"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "nomic-embed-text"
	}
	return &OllamaEmbedder{
		baseURL: base,
		model:   model,
		client:  &http.Client{Timeout: 2 * time.Minute},
	}
}

// NewOllamaWithClient is for tests and custom transports.
func NewOllamaWithClient(cfg OllamaConfig, client *http.Client) *OllamaEmbedder {
	e := NewOllama(cfg)
	if client != nil {
		e.client = client
	}
	return e
}

func (o *OllamaEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(map[string]string{
		"model":  o.model,
		"prompt": text,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, o.baseURL+"/api/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	res, err := o.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ollama embed: %w", err)
	}
	defer res.Body.Close()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("ollama embed: HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Embedding []float32 `json:"embedding"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("ollama embed: decode: %w", err)
	}
	if len(parsed.Embedding) == 0 {
		return nil, fmt.Errorf("ollama embed: empty embedding")
	}
	return parsed.Embedding, nil
}

func (o *OllamaEmbedder) Dimensions() int {
	return 0
}

func (o *OllamaEmbedder) Name() string {
	return ollamaName
}
