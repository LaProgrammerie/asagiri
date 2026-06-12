package replay

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/safeid"
)

// Load reads replay.yaml for a trust run id.
func Load(repoRoot, trustID string) (Manifest, error) {
	if err := safeid.Validate(trustID); err != nil {
		return Manifest{}, fmt.Errorf("load replay manifest: %w", err)
	}
	path := filepath.Join(repoRoot, ".asagiri", "trust", trustID, "replay.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		return Manifest{}, fmt.Errorf("read replay manifest: %w", err)
	}
	var m Manifest
	if err := yaml.Unmarshal(data, &m); err != nil {
		return Manifest{}, fmt.Errorf("parse replay manifest: %w", err)
	}
	if m.TrustID == "" {
		m.TrustID = trustID
	}
	return m, nil
}
