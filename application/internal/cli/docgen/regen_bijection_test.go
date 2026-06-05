package docgen_test

// Feature: audit-coherence-consolidation, Property 1
// Design property: P1 — Bijection commande ↔ page MDX.
// Validates: Requirements 1.1
//
// For any Command_Tree, regenerating through docgen.Generate yields exactly one
// MDX page (Slug(p).mdx) for every reachable command returned by
// CommandPathsWithoutRoot (including `runs`), and no orphan .mdx page without a
// corresponding command. meta.json is hand-maintained and never emitted, so the
// "present" set is restricted to .mdx files.

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"testing/quick"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/cli"
	"github.com/LaProgrammerie/asagiri/application/internal/cli/docgen"
)

func TestProperty1BijectionCommandMDX(t *testing.T) {
	// Anchor the property on the living CLI tree: every reachable command must
	// have exactly one page and there must be no orphan page. `runs` (ADR-029,
	// AUD-001) must be present with its own MDX page.
	t.Run("RootCommand", func(t *testing.T) {
		root := cli.RootCommand()
		dir := t.TempDir()
		if err := docgen.Generate(root, dir); err != nil {
			t.Fatalf("generate: %v", err)
		}

		paths := docgen.CommandPathsWithoutRoot(root)
		if len(paths) == 0 {
			t.Fatal("expected at least one command path")
		}

		missing, orphan := bijectionDiff(t, root, dir)
		if len(missing) > 0 || len(orphan) > 0 {
			t.Fatalf("bijection broken on live CLI tree: missing=%v orphan=%v", missing, orphan)
		}

		if !containsPath(paths, "runs") {
			t.Fatalf(`expected "runs" command in CommandPathsWithoutRoot, got %v`, paths)
		}
		runsPage := filepath.Join(dir, docgen.Slug("runs")+".mdx")
		if _, err := os.Stat(runsPage); err != nil {
			t.Fatalf("expected exactly one page for runs command at %s: %v", runsPage, err)
		}
	})

	// Property over many synthetic Cobra trees (>= 100 iterations, fixed seed for
	// determinism). Synthetic command names are globally unique, so every
	// reachable path maps to a unique, non-empty slug — the bijection must hold
	// for every generated tree shape.
	t.Run("synthetic", func(t *testing.T) {
		var lastMissing, lastOrphan []string
		property := func(tree commandTree) bool {
			dir, err := os.MkdirTemp("", "docgen-bijection-")
			if err != nil {
				lastMissing = []string{"mkdir-temp: " + err.Error()}
				return false
			}
			defer func() { _ = os.RemoveAll(dir) }()

			if err := docgen.Generate(tree.root, dir); err != nil {
				lastMissing = []string{"generate: " + err.Error()}
				return false
			}
			missing, orphan := bijectionDiff(t, tree.root, dir)
			lastMissing, lastOrphan = missing, orphan
			return len(missing) == 0 && len(orphan) == 0
		}

		cfg := &quick.Config{
			MaxCount: 200,
			Rand:     rand.New(rand.NewSource(1)), // deterministic generator
		}
		if err := quick.Check(property, cfg); err != nil {
			t.Fatalf("bijection property failed: %v (missing=%v orphan=%v)", err, lastMissing, lastOrphan)
		}
	})
}

// bijectionDiff compares the set of expected MDX filenames (one per reachable
// command path) against the set of .mdx files actually present in dir. It
// returns the missing files (expected but absent) and orphan files (present but
// unexpected). Non-.mdx entries (e.g. the hand-maintained meta.json) are ignored.
func bijectionDiff(t *testing.T, root *cobra.Command, dir string) (missing, orphan []string) {
	t.Helper()

	expected := map[string]struct{}{}
	for _, rel := range docgen.CommandPathsWithoutRoot(root) {
		expected[docgen.Slug(rel)+".mdx"] = struct{}{}
	}

	present := map[string]struct{}{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read output dir %s: %v", dir, err)
	}
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".mdx" {
			continue
		}
		present[e.Name()] = struct{}{}
	}

	for f := range expected {
		if _, ok := present[f]; !ok {
			missing = append(missing, f)
		}
	}
	for f := range present {
		if _, ok := expected[f]; !ok {
			orphan = append(orphan, f)
		}
	}
	sort.Strings(missing)
	sort.Strings(orphan)
	return missing, orphan
}

func containsPath(paths []string, want string) bool {
	for _, p := range paths {
		if p == want {
			return true
		}
	}
	return false
}

// commandTree is a testing/quick generator producing random cobra command trees.
// Each command name is globally unique ("n1", "n2", ...), guaranteeing that
// Slug() yields a unique, non-empty filename for every reachable command path,
// so the bijection property is structurally well-defined for any generated tree.
type commandTree struct {
	root *cobra.Command
}

// Generate implements quick.Generator.
func (commandTree) Generate(r *rand.Rand, _ int) reflect.Value {
	counter := 0
	nextName := func() string {
		counter++
		return fmt.Sprintf("n%d", counter)
	}

	root := &cobra.Command{Use: "root"}

	const maxDepth = 3
	var attach func(parent *cobra.Command, depth int)
	attach = func(parent *cobra.Command, depth int) {
		if depth <= 0 {
			return
		}
		breadth := r.Intn(4) + 1 // 1..4 children at this level
		for i := 0; i < breadth; i++ {
			child := &cobra.Command{Use: nextName(), Short: "synthetic node"}
			parent.AddCommand(child)
			if r.Intn(2) == 0 { // sometimes recurse to vary depth
				attach(child, depth-1)
			}
		}
	}
	attach(root, maxDepth)

	return reflect.ValueOf(commandTree{root: root})
}
