package embedder

import (
	"context"
	"sync"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

var (
	activeMu sync.RWMutex
	active   Embedder = NewHash()
)

// Configure replaces the process-wide embedder.
func Configure(e Embedder) {
	if e == nil {
		e = NewHash()
	}
	activeMu.Lock()
	active = e
	activeMu.Unlock()
}

// ConfigureFromConfig builds and installs the embedder from runtime.memory config.
func ConfigureFromConfig(mc config.RuntimeMemoryConfig) error {
	e, err := NewFromConfig(mc)
	if err != nil {
		return err
	}
	Configure(e)
	return nil
}

// Current returns the active embedder.
func Current() Embedder {
	activeMu.RLock()
	defer activeMu.RUnlock()
	return active
}

// EmbedText embeds text with the active embedder.
func EmbedText(ctx context.Context, text string) ([]float32, error) {
	return Current().Embed(ctx, text)
}
