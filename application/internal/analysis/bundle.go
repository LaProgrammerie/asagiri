package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// LoadBundle reads persisted graphs from .asagiri/analysis/<product>/graphs.json.
func LoadBundle(repoRoot, productID string) (Bundle, error) {
	if productID == "" {
		return Bundle{}, fmt.Errorf("analysis: product id required")
	}
	path := filepath.Join(repoRoot, analysisRel, productID, "graphs.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return Bundle{}, fmt.Errorf("load analysis bundle %s: %w", path, err)
	}
	var b Bundle
	if err := json.Unmarshal(raw, &b); err != nil {
		return Bundle{}, fmt.Errorf("decode analysis bundle: %w", err)
	}
	if b.Product == "" {
		b.Product = productID
	}
	return b, nil
}
