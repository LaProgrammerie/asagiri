package embedder

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/embedutil"
)

const hashName = "hash"

// HashEmbedder uses deterministic bag-of-words vectors (legacy embedutil behaviour).
type HashEmbedder struct{}

// NewHash returns the local hash embedder.
func NewHash() *HashEmbedder {
	return &HashEmbedder{}
}

func (h *HashEmbedder) Embed(_ context.Context, text string) ([]float32, error) {
	return embedutil.Vector(text), nil
}

func (h *HashEmbedder) Dimensions() int {
	return embedutil.Dims
}

func (h *HashEmbedder) Name() string {
	return hashName
}
