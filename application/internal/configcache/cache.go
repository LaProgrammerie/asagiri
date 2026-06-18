package configcache

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/cloud"
)

const cacheRootRel = ".asagiri/cache/config"

// ActivePointer tracks the currently selected cached release.
type ActivePointer struct {
	ProjectID      string    `json:"projectId"`
	ProjectSlug    string    `json:"projectSlug"`
	ReleaseVersion int       `json:"releaseVersion"`
	ContentHash    string    `json:"contentHash"`
	FetchedAt      time.Time `json:"fetchedAt"`
	Source         string    `json:"source"`
}

// ReleaseMeta stores immutable release metadata beside the snapshot.
type ReleaseMeta struct {
	ReleaseVersion      int       `json:"releaseVersion"`
	PublishedAt         string    `json:"publishedAt,omitempty"`
	ETag                string    `json:"etag"`
	APIVersion          string    `json:"apiVersion,omitempty"`
	RegistrySchemaVersion int     `json:"registrySchemaVersion,omitempty"`
	CachedAt            time.Time `json:"cachedAt"`
}

// Store manages ~/.asagiri/cache/config/{projectSlug}/.
type Store struct {
	RootDir string
}

func NewStore(homeDir string) Store {
	return Store{RootDir: filepath.Join(strings.TrimSpace(homeDir), filepath.FromSlash(cacheRootRel))}
}

func (s Store) projectDir(projectSlug string) string {
	return filepath.Join(s.RootDir, sanitizeSlug(projectSlug))
}

func (s Store) releasesDir(projectSlug string) string {
	return filepath.Join(s.projectDir(projectSlug), "releases")
}

// LoadActive reads active.json for a project slug.
func (s Store) LoadActive(projectSlug string) (ActivePointer, bool, error) {
	path := filepath.Join(s.projectDir(projectSlug), "active.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ActivePointer{}, false, nil
		}
		return ActivePointer{}, false, err
	}
	var ptr ActivePointer
	if err := json.Unmarshal(raw, &ptr); err != nil {
		return ActivePointer{}, false, fmt.Errorf("configcache: decode active.json: %w", err)
	}
	return ptr, true, nil
}

// LoadBundle returns a cached bundle by content hash.
func (s Store) LoadBundle(projectSlug, contentHash string) (cloud.ConfigBundle, ReleaseMeta, bool, error) {
	contentHash = strings.TrimSpace(contentHash)
	if contentHash == "" {
		return nil, ReleaseMeta{}, false, nil
	}
	bundlePath := filepath.Join(s.releasesDir(projectSlug), contentHash+".json")
	metaPath := filepath.Join(s.releasesDir(projectSlug), contentHash+".meta.json")
	raw, err := os.ReadFile(bundlePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, ReleaseMeta{}, false, nil
		}
		return nil, ReleaseMeta{}, false, err
	}
	var bundle cloud.ConfigBundle
	if err := json.Unmarshal(raw, &bundle); err != nil {
		return nil, ReleaseMeta{}, false, fmt.Errorf("configcache: decode bundle: %w", err)
	}
	metaRaw, err := os.ReadFile(metaPath)
	if err != nil && !os.IsNotExist(err) {
		return nil, ReleaseMeta{}, false, err
	}
	var meta ReleaseMeta
	if len(metaRaw) > 0 {
		if err := json.Unmarshal(metaRaw, &meta); err != nil {
			return nil, ReleaseMeta{}, false, fmt.Errorf("configcache: decode meta: %w", err)
		}
	}
	return bundle, meta, true, nil
}

// SaveBundle writes an immutable release snapshot and updates active.json.
func (s Store) SaveBundle(projectSlug string, fetched cloud.BundleFetchResult) (ActivePointer, error) {
	projectSlug = sanitizeSlug(projectSlug)
	if projectSlug == "" {
		return ActivePointer{}, fmt.Errorf("configcache: project slug requis")
	}
	hash := strings.TrimSpace(fetched.ContentHash)
	if hash == "" {
		return ActivePointer{}, fmt.Errorf("configcache: content hash requis")
	}
	if err := os.MkdirAll(s.releasesDir(projectSlug), 0o755); err != nil {
		return ActivePointer{}, err
	}

	bundlePath := filepath.Join(s.releasesDir(projectSlug), hash+".json")
	metaPath := filepath.Join(s.releasesDir(projectSlug), hash+".meta.json")

	bundleRaw, err := json.MarshalIndent(fetched.Bundle, "", "  ")
	if err != nil {
		return ActivePointer{}, err
	}
	if err := os.WriteFile(bundlePath, append(bundleRaw, '\n'), 0o644); err != nil {
		return ActivePointer{}, err
	}

	publishedAt := ""
	if prov, ok := fetched.Bundle["provenance"].(map[string]any); ok {
		if v, ok := prov["publishedAt"].(string); ok {
			publishedAt = v
		}
	}
	meta := ReleaseMeta{
		ReleaseVersion:        fetched.ReleaseVersion,
		PublishedAt:           publishedAt,
		ETag:                  fetched.ETag,
		APIVersion:            fetched.APIVersion,
		RegistrySchemaVersion: fetched.RegistrySchema,
		CachedAt:              time.Now().UTC(),
	}
	metaRaw, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		return ActivePointer{}, err
	}
	if err := os.WriteFile(metaPath, append(metaRaw, '\n'), 0o644); err != nil {
		return ActivePointer{}, err
	}

	projectID := ""
	projectSlugFromBundle := projectSlug
	if v, ok := fetched.Bundle["projectId"].(string); ok {
		projectID = v
	}
	if v, ok := fetched.Bundle["projectSlug"].(string); ok && strings.TrimSpace(v) != "" {
		projectSlugFromBundle = sanitizeSlug(v)
	}

	ptr := ActivePointer{
		ProjectID:      projectID,
		ProjectSlug:    projectSlugFromBundle,
		ReleaseVersion: fetched.ReleaseVersion,
		ContentHash:    hash,
		FetchedAt:      time.Now().UTC(),
		Source:         "cloud",
	}
	ptrRaw, err := json.MarshalIndent(ptr, "", "  ")
	if err != nil {
		return ActivePointer{}, err
	}
	if err := os.MkdirAll(s.projectDir(projectSlugFromBundle), 0o755); err != nil {
		return ActivePointer{}, err
	}
	if err := os.WriteFile(filepath.Join(s.projectDir(projectSlugFromBundle), "active.json"), append(ptrRaw, '\n'), 0o644); err != nil {
		return ActivePointer{}, err
	}
	return ptr, nil
}

// CurrentBundle loads the active cached bundle for a project slug.
func (s Store) CurrentBundle(projectSlug string) (cloud.ConfigBundle, ActivePointer, ReleaseMeta, bool, error) {
	ptr, ok, err := s.LoadActive(projectSlug)
	if err != nil || !ok {
		return nil, ActivePointer{}, ReleaseMeta{}, false, err
	}
	bundle, meta, ok, err := s.LoadBundle(projectSlug, ptr.ContentHash)
	if err != nil || !ok {
		return nil, ptr, ReleaseMeta{}, false, err
	}
	return bundle, ptr, meta, true, nil
}

func sanitizeSlug(slug string) string {
	slug = strings.TrimSpace(slug)
	if slug == "" {
		return "default"
	}
	return slug
}
