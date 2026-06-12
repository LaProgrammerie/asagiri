package reportsink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ReadJSONFile loads and unmarshals a report JSON from an absolute or repo-relative path.
func ReadJSONFile(repoRoot, relOrAbs string, dest any) error {
	relOrAbs = strings.TrimSpace(relOrAbs)
	if relOrAbs == "" {
		return fmt.Errorf("reports: path required")
	}
	abs := relOrAbs
	if !filepath.IsAbs(relOrAbs) {
		abs = filepath.Join(repoRoot, filepath.FromSlash(strings.TrimPrefix(relOrAbs, "/")))
	}
	body, err := os.ReadFile(abs)
	if err != nil {
		return fmt.Errorf("reports: read %s: %w", relOrAbs, err)
	}
	if err := json.Unmarshal(body, dest); err != nil {
		return fmt.Errorf("reports: decode %s: %w", relOrAbs, err)
	}
	return nil
}

// TrustTaskRel returns the stable latest path for a task report.
func TrustTaskRel(taskID string) (string, error) {
	return trustTaskRel(taskID)
}

// TrustFeatureRel returns the stable latest path for a feature report.
func TrustFeatureRel(feature string) (string, error) {
	return trustFeatureRel(feature)
}

// TrustRunRel returns the stable latest path for a run report.
func TrustRunRel(runID string) (string, error) {
	return trustRunRel(runID)
}

// DoctorLatestRel returns the stable doctor latest path.
func DoctorLatestRel() string {
	return doctorLatestRel()
}
