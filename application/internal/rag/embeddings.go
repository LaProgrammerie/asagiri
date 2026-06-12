package rag

import (
	"github.com/LaProgrammerie/asagiri/application/internal/memory"
)

// marshalEmbedding stores a vector in the index database.
func marshalEmbedding(v []float32) string {
	return memory.MarshalEmbedding(v)
}

// unmarshalEmbedding loads a stored vector.
func unmarshalEmbedding(raw string) []float32 {
	return memory.UnmarshalEmbedding(raw)
}

// cosineSimilarity ranks chunks against a query vector.
func cosineSimilarity(a, b []float32) float64 {
	return memory.CosineSimilarity(a, b)
}
