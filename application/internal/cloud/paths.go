package cloud

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// ExpandPath resolves ~ in token_path and returns an absolute path.
func ExpandPath(p string) (string, error) {
	p = strings.TrimSpace(p)
	if p == "" {
		return "", nil
	}
	if strings.HasPrefix(p, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		p = filepath.Join(home, strings.TrimPrefix(p, "~/"))
	}
	return filepath.Clean(p), nil
}

// TokenPath returns the resolved token file path from config defaults.
func TokenPath(cfg *config.Config) (string, error) {
	if cfg == nil {
		return ExpandPath(config.DefaultCloudTokenRel)
	}
	return ExpandPath(cfg.Cloud.TokenPath)
}

// ProjectIRI builds the API Platform IRI for a project UUID.
func ProjectIRI(baseURL, projectID string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	id := strings.TrimSpace(projectID)
	return base + "/api/v1/projects/" + id
}

// RunIRI builds the API Platform IRI for a run UUID.
func RunIRI(baseURL, runID string) string {
	base := strings.TrimRight(strings.TrimSpace(baseURL), "/")
	id := strings.TrimSpace(runID)
	return base + "/api/v1/runs/" + id
}
