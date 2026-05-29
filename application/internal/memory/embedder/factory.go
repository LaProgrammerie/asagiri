package embedder

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// NewFromConfig selects the configured runtime memory embedder.
func NewFromConfig(mc config.RuntimeMemoryConfig) (Embedder, error) {
	kind := strings.ToLower(strings.TrimSpace(mc.Embedder))
	switch kind {
	case "", "hash":
		return NewHash(), nil
	case "ollama":
		return NewOllama(OllamaConfig{
			BaseURL: mc.Ollama.BaseURL,
			Model:   mc.Ollama.Model,
		}), nil
	case "cloud":
		return NewCloud(CloudConfig{
			Enabled:  mc.Cloud.Enabled,
			Provider: mc.Cloud.Provider,
			BaseURL:  mc.Cloud.BaseURL,
			Model:    mc.Cloud.Model,
			TokenEnv: mc.Cloud.TokenEnv,
		})
	default:
		return nil, fmt.Errorf("memory embedder: unknown %q (use hash, ollama, or cloud)", mc.Embedder)
	}
}
