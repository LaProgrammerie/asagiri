package embedder

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Reachable checks that the Ollama HTTP API responds (GET /api/tags).
func (o *OllamaEmbedder) Reachable(ctx context.Context) error {
	if o == nil {
		return fmt.Errorf("ollama: embedder not configured")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, o.baseURL+"/api/tags", nil)
	if err != nil {
		return err
	}
	res, err := o.client.Do(req)
	if err != nil {
		return fmt.Errorf("ollama unreachable at %s: %w", o.baseURL, err)
	}
	defer func() { _ = res.Body.Close() }()
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		raw, _ := io.ReadAll(res.Body)
		return fmt.Errorf("ollama unreachable at %s: HTTP %d: %s", o.baseURL, res.StatusCode, strings.TrimSpace(string(raw)))
	}
	return nil
}
