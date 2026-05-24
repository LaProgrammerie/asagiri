package source

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// WriteLocalSpec writes a synced document to .asagiri/specs/<feature>/.
func WriteLocalSpec(repoRoot, importRel string, feature string, doc SourceDocument, opts SyncOptions) (SyncResult, error) {
	dest := LocalSpecPath{Root: importRel, Feature: feature}
	return writeLocalSpec(repoRoot, dest, doc, opts)
}

func writeLocalSpec(repoRoot string, dest LocalSpecPath, doc SourceDocument, opts SyncOptions) (SyncResult, error) {
	feature := dest.Feature
	if feature == "" {
		feature = doc.Feature
	}
	if strings.TrimSpace(doc.Markdown) == "" {
		return SyncResult{}, fmt.Errorf("spec vide: refus d'importer")
	}
	importRoot := dest.Root
	if !filepath.IsAbs(importRoot) {
		importRoot = filepath.Join(repoRoot, filepath.Clean(importRoot))
	}
	dir := filepath.Join(importRoot, feature)
	specPath := filepath.Join(dir, "spec.md")
	if _, err := os.Stat(specPath); err == nil && !opts.Force {
		if !opts.Interactive {
			return SyncResult{Feature: feature, Path: dir, Conflict: true, NeedsConfirm: true}, fmt.Errorf("spec locale modifiée: utilisez --force ou confirmez")
		}
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return SyncResult{}, err
	}
	if err := os.WriteFile(specPath, []byte(doc.Markdown), 0o644); err != nil {
		return SyncResult{}, err
	}
	if doc.TasksYAML != "" {
		if err := os.WriteFile(filepath.Join(dir, "tasks.yaml"), []byte(doc.TasksYAML), 0o644); err != nil {
			return SyncResult{}, err
		}
	}
	now := time.Now().UTC()
	srcType := "local"
	if doc.Ref.ID != "" || strings.Contains(doc.Ref.URL, "notion") {
		srcType = "notion"
	}
	src := map[string]any{
		"type":              srcType,
		"page_id":           doc.Ref.ID,
		"url":               doc.Ref.URL,
		"last_synced_at":    now.Format(time.RFC3339),
		"remote_updated_at": doc.RemoteUpdatedAt.Format(time.RFC3339),
	}
	body, _ := json.MarshalIndent(src, "", "  ")
	if err := os.WriteFile(filepath.Join(dir, "source.json"), body, 0o644); err != nil {
		return SyncResult{}, err
	}
	status := doc.Status
	if status == "" {
		status = "ready"
	}
	meta := map[string]any{
		"feature":   feature,
		"source":    "notion",
		"status":    status,
		"synced_at": now.Format(time.RFC3339),
	}
	mb, _ := yaml.Marshal(meta)
	if err := os.WriteFile(filepath.Join(dir, "metadata.yaml"), mb, 0o644); err != nil {
		return SyncResult{}, err
	}
	return SyncResult{Feature: feature, Path: dir, Overwritten: true}, nil
}
