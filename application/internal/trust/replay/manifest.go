package replay

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/safeid"
)

// Manifest captures inputs required to replay a verification (spec §21).
type Manifest struct {
	TrustID    string   `yaml:"trust_id"`
	Checks     []string `yaml:"checks,omitempty"`
	RepoCommit string   `yaml:"repo_commit,omitempty"`
	Flow       string   `yaml:"flow,omitempty"`
	Branch     string   `yaml:"branch,omitempty"`
	Commands   []string `yaml:"commands,omitempty"`
}

// WriteReplay writes a stub replay.yaml under .asagiri/trust/<id>/ (lot 5 expands content).
func WriteReplay(repoRoot, trustID string, m Manifest) error {
	if err := safeid.Validate(trustID); err != nil {
		return fmt.Errorf("write replay manifest: %w", err)
	}
	dir := filepath.Join(repoRoot, ".asagiri", "trust", trustID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create trust replay dir: %w", err)
	}
	if m.TrustID == "" {
		m.TrustID = trustID
	}
	body, err := yaml.Marshal(m)
	if err != nil {
		return fmt.Errorf("marshal replay manifest: %w", err)
	}
	path := filepath.Join(dir, "replay.yaml")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		return fmt.Errorf("write replay manifest: %w", err)
	}
	return nil
}
