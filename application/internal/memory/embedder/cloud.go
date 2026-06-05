package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

const cloudName = "cloud"

// CloudConfig configures an OpenAI-compatible embeddings API (opt-in only).
type CloudConfig struct {
	Enabled   bool
	Provider  string
	BaseURL   string
	Model     string
	TokenEnv  string
}

// CloudEmbedder calls a provider embeddings endpoint when explicitly enabled.
type CloudEmbedder struct {
	baseURL string
	model   string
	token   string
	client  *http.Client
}

// NewCloud returns a cloud embedder or an error when disabled.
func NewCloud(cfg CloudConfig) (*CloudEmbedder, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("cloud embedder: disabled (set runtime.memory.cloud.enabled: true to opt in)")
	}
	tokenEnv := strings.TrimSpace(cfg.TokenEnv)
	if tokenEnv == "" {
		tokenEnv = "OPENAI_API_KEY"
	}
	token := strings.TrimSpace(os.Getenv(tokenEnv))
	if token == "" {
		return nil, fmt.Errorf("cloud embedder: missing env %s", tokenEnv)
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		switch strings.ToLower(strings.TrimSpace(cfg.Provider)) {
		case "openai", "":
			base = "https://api.openai.com/v1"
		default:
			return nil, fmt.Errorf("cloud embedder: base_url required for provider %q", cfg.Provider)
		}
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "text-embedding-3-small"
	}
	return &CloudEmbedder{
		baseURL: base,
		model:   model,
		token:   token,
		client:  &http.Client{Timeout: 2 * time.Minute},
	}, nil
}

// NewCloudWithClient is for tests (token required).
func NewCloudWithClient(cfg CloudConfig, token string, client *http.Client) (*CloudEmbedder, error) {
	cfg.Enabled = true
	if strings.TrimSpace(token) == "" {
		return nil, fmt.Errorf("cloud embedder: token required for test client")
	}
	base := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "text-embedding-3-small"
	}
	e := &CloudEmbedder{
		baseURL: base,
		model:   model,
		token:   token,
		client:  &http.Client{Timeout: 2 * time.Minute},
	}
	if client != nil {
		e.client = client
	}
	return e, nil
}

func (c *CloudEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	body, err := json.Marshal(map[string]any{
		"model": c.model,
		"input": text,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/embeddings", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)
	res, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("cloud embed: %w", err)
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, fmt.Errorf("cloud embed: HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(raw)))
	}
	var parsed struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
		} `json:"data"`
	}
	if err := json.Unmarshal(raw, &parsed); err != nil {
		return nil, fmt.Errorf("cloud embed: decode: %w", err)
	}
	if len(parsed.Data) == 0 || len(parsed.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("cloud embed: empty embedding")
	}
	return parsed.Data[0].Embedding, nil
}

func (c *CloudEmbedder) Dimensions() int {
	return 0
}

func (c *CloudEmbedder) Name() string {
	return cloudName
}
