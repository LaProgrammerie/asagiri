package cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// ConfigBundle is the published project configuration snapshot.
type ConfigBundle map[string]any

// BundleFetchResult includes HTTP metadata from GET /config/bundle.
type BundleFetchResult struct {
	Bundle         ConfigBundle
	ContentHash    string
	ReleaseVersion int
	ETag           string
	RegistrySchema int
	APIVersion     string
}

// FetchConfigBundle downloads the active or pinned release bundle.
func (c *Client) FetchConfigBundle(ctx context.Context, projectID string, releaseVersion *int) (BundleFetchResult, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return BundleFetchResult{}, fmt.Errorf("cloud: project_id requis")
	}
	path := "/api/v1/projects/" + projectID + "/config/bundle"
	if releaseVersion != nil && *releaseVersion > 0 {
		path += "?release=" + strconv.Itoa(*releaseVersion)
	}

	_, raw, headers, err := c.doWithHeaders(ctx, http.MethodGet, path, nil)
	if err != nil {
		return BundleFetchResult{}, err
	}

	var bundle ConfigBundle
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &bundle); err != nil {
			return BundleFetchResult{}, fmt.Errorf("cloud api decode bundle: %w", err)
		}
	}

	result := BundleFetchResult{Bundle: bundle}
	if etag := strings.Trim(headers.Get("ETag"), "\""); etag != "" {
		result.ETag = etag
	}
	if v := strings.TrimSpace(headers.Get("X-Release-Version")); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			result.ReleaseVersion = n
		}
	}
	result.APIVersion = strings.TrimSpace(headers.Get("X-Config-Api-Version"))
	if s := strings.TrimSpace(headers.Get("X-Registry-Schema-Version")); s != "" {
		if n, err := strconv.Atoi(s); err == nil {
			result.RegistrySchema = n
		}
	}
	if hash, ok := bundle["contentHash"].(string); ok {
		result.ContentHash = strings.TrimSpace(hash)
	}
	if result.ContentHash == "" {
		result.ContentHash = result.ETag
	}

	return result, nil
}

// ActivateConfigRelease sets the active published release on the cloud project.
func (c *Client) ActivateConfigRelease(ctx context.Context, projectID string, version int) (map[string]any, error) {
	projectID = strings.TrimSpace(projectID)
	if projectID == "" {
		return nil, fmt.Errorf("cloud: project_id requis")
	}
	if version <= 0 {
		return nil, fmt.Errorf("cloud: version invalide")
	}
	path := fmt.Sprintf("/api/v1/projects/%s/config/releases/%d/activate", projectID, version)
	var out map[string]any
	_, _, err := c.do(ctx, http.MethodPost, path, nil, &out)
	return out, err
}
