package docgen_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPublicDocsNoPlaceholders(t *testing.T) {
	root := repoRoot(t)
	contentDir := filepath.Join(root, "docs-site", "content", "docs")
	bad := []string{
		"placeholder content",
		"Hyper Fast Builder",
		"Coming soon",
	}
	var hits []string
	err := filepath.Walk(contentDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Ext(path) != ".mdx" {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"cli"+string(filepath.Separator)+"generated"+string(filepath.Separator)) {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		body := string(data)
		for _, needle := range bad {
			if strings.Contains(body, needle) {
				hits = append(hits, path+": contains "+needle)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(hits) > 0 {
		t.Fatalf("placeholder or stale branding in public docs:\n%s", strings.Join(hits, "\n"))
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			if _, err := os.Stat(filepath.Join(dir, "docs-site")); err == nil {
				return dir
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}
