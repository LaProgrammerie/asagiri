package source

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"gopkg.in/yaml.v3"
)

// LocalSource scans configured local paths (specv2 §7.2).
type LocalSource struct {
	RepoRoot string
	Config   config.LocalSourceConfig
}

func (l *LocalSource) Name() string { return "local" }

func (l *LocalSource) List(ctx context.Context) ([]SourceItem, error) {
	_ = ctx
	if !l.Config.Enabled {
		return nil, nil
	}
	var items []SourceItem
	for _, rel := range l.Config.Paths {
		root := filepath.Join(l.RepoRoot, filepath.Clean(rel))
		entries, err := os.ReadDir(root)
		if err != nil {
			if os.IsNotExist(err) {
				continue
			}
			return nil, err
		}
		for _, e := range entries {
			if !e.IsDir() {
				if strings.HasSuffix(e.Name(), ".md") && rel == "docs/ai/active" {
					items = append(items, SourceItem{
						Ref:       SourceRef{Name: strings.TrimSuffix(e.Name(), ".md")},
						Feature:   strings.TrimSuffix(e.Name(), ".md"),
						Status:    "ready",
						UpdatedAt: modTime(filepath.Join(root, e.Name())),
						PathHint:  filepath.Join(rel, e.Name()),
					})
				}
				continue
			}
			dir := filepath.Join(root, e.Name())
			status := readLocalStatus(dir)
			items = append(items, SourceItem{
				Ref:       SourceRef{Name: e.Name()},
				Feature:   e.Name(),
				Status:    status,
				UpdatedAt: modTime(dir),
				PathHint:  filepath.Join(rel, e.Name()),
			})
		}
	}
	return items, nil
}

func (l *LocalSource) Fetch(ctx context.Context, ref SourceRef) (SourceDocument, error) {
	_ = ctx
	for _, rel := range l.Config.Paths {
		dir := filepath.Join(l.RepoRoot, rel, ref.Name)
		specPath := filepath.Join(dir, "spec.md")
		data, err := os.ReadFile(specPath)
		if err == nil {
			return SourceDocument{
				Feature:  ref.Name,
				Title:    ref.Name,
				Markdown: string(data),
				Status:   readLocalStatus(dir),
				Ref:      ref,
			}, nil
		}
		req := filepath.Join(dir, "requirements.md")
		data, err = os.ReadFile(req)
		if err == nil {
			return SourceDocument{
				Feature:  ref.Name,
				Title:    ref.Name,
				Markdown: string(data),
				Status:   "ready",
				Ref:      ref,
			}, nil
		}
	}
	return SourceDocument{}, os.ErrNotExist
}

func (l *LocalSource) Sync(ctx context.Context, ref SourceRef, dest LocalSpecPath, opts SyncOptions) (SyncResult, error) {
	_ = ctx
	_ = opts
	doc, err := l.Fetch(ctx, ref)
	if err != nil {
		return SyncResult{}, err
	}
	return writeLocalSpec(l.RepoRoot, dest, doc, opts)
}

func readLocalStatus(dir string) string {
	meta := filepath.Join(dir, "metadata.yaml")
	data, err := os.ReadFile(meta)
	if err != nil {
		return "ready"
	}
	var m struct {
		Status string `yaml:"status"`
	}
	if err := yaml.Unmarshal(data, &m); err != nil || m.Status == "" {
		return "ready"
	}
	return m.Status
}

func modTime(path string) time.Time {
	fi, err := os.Stat(path)
	if err != nil {
		return time.Time{}
	}
	return fi.ModTime()
}
