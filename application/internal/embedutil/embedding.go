package embedutil

import (
	"encoding/json"
	"hash/fnv"
	"math"
	"strings"
)

const Dims = 64

// Vector returns a normalized bag-of-words embedding.
func Vector(text string) []float32 {
	vec := make([]float32, Dims)
	for _, tok := range tokenize(text) {
		h := fnv.New32a()
		_, _ = h.Write([]byte(tok))
		idx := int(h.Sum32() % uint32(Dims))
		vec[idx] += 1
	}
	var norm float64
	for _, v := range vec {
		norm += float64(v * v)
	}
	if norm == 0 {
		return vec
	}
	norm = math.Sqrt(norm)
	for i := range vec {
		vec[i] /= float32(norm)
	}
	return vec
}

// ToJSON marshals a vector for SQLite.
func ToJSON(v []float32) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func tokenize(s string) []string {
	s = strings.ToLower(s)
	var out []string
	for _, w := range strings.FieldsFunc(s, func(r rune) bool {
		return (r < 'a' || r > 'z') && (r < '0' || r > '9')
	}) {
		if len(w) >= 2 {
			out = append(out, w)
		}
	}
	return out
}
