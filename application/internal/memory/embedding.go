package memory

import (
	"encoding/json"
	"math"
	"sort"

	"github.com/LaProgrammerie/asagiri/application/internal/embedutil"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// Embed builds a deterministic bag-of-words vector (spec-my-A §24.10).
func Embed(text string) []float32 {
	return embedutil.Vector(text)
}

// CosineSimilarity returns similarity in [0,1].
func CosineSimilarity(a, b []float32) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0
	}
	var dot, na, nb float64
	for i := range a {
		dot += float64(a[i] * b[i])
		na += float64(a[i] * a[i])
		nb += float64(b[i] * b[i])
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// MarshalEmbedding JSON-encodes a vector for SQLite storage.
func MarshalEmbedding(v []float32) string {
	return embedutil.ToJSON(v)
}

// UnmarshalEmbedding decodes stored vector.
func UnmarshalEmbedding(raw string) []float32 {
	if raw == "" {
		return nil
	}
	var out []float32
	_ = json.Unmarshal([]byte(raw), &out)
	return out
}

// RetrieveByQuery ranks memory entries by embedding similarity.
func (e *Engine) RetrieveByQuery(query string, limit int) ([]runtime.MemoryEntry, error) {
	if e == nil || e.store == nil {
		return nil, nil
	}
	if query == "" {
		return e.Retrieve("", nil, limit)
	}
	qv := Embed(query)
	entries, err := e.store.ListMemory("", 0)
	if err != nil {
		return nil, err
	}
	type scored struct {
		ent   runtime.MemoryEntry
		score float64
	}
	var ranked []scored
	for _, ent := range entries {
		ev := UnmarshalEmbedding(ent.EmbeddingJSON)
		if len(ev) == 0 {
			continue
		}
		ranked = append(ranked, scored{ent: ent, score: CosineSimilarity(qv, ev)})
	}
	sort.Slice(ranked, func(i, j int) bool { return ranked[i].score > ranked[j].score })
	if limit <= 0 {
		limit = 10
	}
	var out []runtime.MemoryEntry
	for i := 0; i < len(ranked) && i < limit; i++ {
		out = append(out, ranked[i].ent)
		e.store.BumpMemoryLookup(ranked[i].score > 0.3)
	}
	return out, nil
}
