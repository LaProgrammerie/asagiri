package embedder

import "context"

// Embedder produces dense vectors for semantic memory retrieval (spec-phase-finale PF-A-01).
type Embedder interface {
	Embed(ctx context.Context, text string) ([]float32, error)
	Dimensions() int
	Name() string
}
