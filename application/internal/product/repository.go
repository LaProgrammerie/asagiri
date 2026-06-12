package product

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type Repository struct {
	RepoRoot string
}

func NewRepository(repoRoot string) *Repository {
	return &Repository{RepoRoot: repoRoot}
}

func (r *Repository) productRoot(product string) string {
	return filepath.Join(r.RepoRoot, ".asagiri", "products", Slug(product))
}

func (r *Repository) specsRoot(product string) string {
	return filepath.Join(r.RepoRoot, ".asagiri", "specs", Slug(product))
}

func (r *Repository) tasksRoot(product string) string {
	return filepath.Join(r.RepoRoot, ".asagiri", "tasks", Slug(product))
}

func (r *Repository) EnsureRoots(product string) error {
	dirs := []string{
		r.productRoot(product),
		filepath.Join(r.productRoot(product), "prototype"),
		filepath.Join(r.productRoot(product), "flows"),
		filepath.Join(r.productRoot(product), "screens"),
		filepath.Join(r.productRoot(product), "contracts"),
		filepath.Join(r.productRoot(product), "extraction"),
		filepath.Join(r.productRoot(product), "generated-specs"),
		filepath.Join(r.productRoot(product), "reviews"),
		r.specsRoot(product),
		r.tasksRoot(product),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) WriteProductFile(product, rel string, body []byte) error {
	return r.writeSafe(r.productRoot(product), rel, body)
}

func (r *Repository) WriteSpecsFile(product, rel string, body []byte) error {
	return r.writeSafe(r.specsRoot(product), rel, body)
}

func (r *Repository) WriteTaskFile(product, rel string, body []byte) error {
	return r.writeSafe(r.tasksRoot(product), rel, body)
}

func (r *Repository) writeSafe(base, rel string, body []byte) error {
	cleanRel := filepath.Clean(rel)
	if strings.HasPrefix(cleanRel, "..") {
		return fmt.Errorf("refus path traversal: %s", rel)
	}
	target := filepath.Join(base, cleanRel)
	if !strings.HasPrefix(target, base+string(filepath.Separator)) && target != base {
		return fmt.Errorf("path hors scope: %s", rel)
	}
	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		return err
	}
	return os.WriteFile(target, body, 0o644)
}

func EncodeYAML(v any) ([]byte, error) {
	out, err := yaml.Marshal(v)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func Slug(value string) string {
	v := strings.TrimSpace(strings.ToLower(value))
	v = strings.ReplaceAll(v, "_", "-")
	v = strings.ReplaceAll(v, " ", "-")
	v = strings.Trim(v, "-")
	if v == "" {
		return "product"
	}
	return v
}
