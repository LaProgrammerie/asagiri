package docgen_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/cli"
	"github.com/LaProgrammerie/asagiri/application/internal/cli/docgen"
)

func TestCLICommandsDocumented(t *testing.T) {
	root := cli.RootCommand()
	dir := t.TempDir()

	if err := docgen.Generate(root, dir); err != nil {
		t.Fatalf("generate: %v", err)
	}

	paths := docgen.CommandPathsWithoutRoot(root)
	if len(paths) == 0 {
		t.Fatal("expected at least one command path")
	}

	seenFiles := map[string]struct{}{}
	for _, rel := range paths {
		slug := docgen.Slug(rel)

		fp := filepath.Join(dir, slug+".mdx")

		stat, err := os.Stat(fp)
		if err != nil {
			t.Fatalf("missing file for command path %q (slug=%q): %v", rel, slug, err)
		}
		if stat.Size() == 0 {
			t.Fatalf("empty file for command path %q", rel)
		}
		seenFiles[fp] = struct{}{}
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read output dir: %v", err)
	}
	if len(entries) != len(seenFiles) {
		t.Fatalf("unexpected artefacts: entries=%d want=%d", len(entries), len(seenFiles))
	}
	for _, entry := range entries {
		if entry.IsDir() {
			t.Fatalf("unexpected subdirectory %s", entry.Name())
		}
		if filepath.Ext(entry.Name()) != ".mdx" {
			t.Fatalf("unexpected extension for %s", entry.Name())
		}
		if _, ok := seenFiles[filepath.Join(dir, entry.Name())]; !ok {
			t.Fatalf("unexpected file %s", entry.Name())
		}
	}
}
