package docgen_test

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/quick"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/cli"
	"github.com/LaProgrammerie/asagiri/application/internal/cli/docgen"
)

// Feature: audit-coherence-consolidation, Property 2
// Design property P2 — Déterminisme octet-à-octet de Docgen.
//
// For any Command_Tree, two executions of docgen.Generate towards distinct
// directories produce the same set of .mdx files and byte-for-byte identical
// contents.
//
// Validates: Requirements 1.2
func TestDocgenByteForByteDeterminism(t *testing.T) {
	// Anchor the property on the real production command tree first: the
	// generated synthetic trees broaden coverage, but the shipping CLI must
	// itself be deterministic.
	if err := determinismGenerateTwiceAndDiff(t, cli.RootCommand()); err != nil {
		t.Fatalf("real RootCommand() docgen output is not byte-for-byte deterministic: %v", err)
	}

	var failure string
	property := func(spec determinismTreeSpec) bool {
		if err := determinismGenerateTwiceAndDiff(t, spec.root); err != nil {
			failure = err.Error()
			return false
		}
		return true
	}

	// >= 100 iterations over generated command-subtree inputs, with a fixed
	// seed so the generator stays deterministic across runs.
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(20260531)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("docgen byte-for-byte determinism violated: %v\ndetail: %s", err, failure)
	}
}

// determinismGenerateTwiceAndDiff runs docgen.Generate twice towards two
// distinct t.TempDir() directories and asserts the produced .mdx corpora are
// identical (same paths, identical bytes).
func determinismGenerateTwiceAndDiff(t *testing.T, root *cobra.Command) error {
	t.Helper()
	dirA := t.TempDir()
	dirB := t.TempDir()
	if err := docgen.Generate(root, dirA); err != nil {
		return fmt.Errorf("generate run A: %w", err)
	}
	if err := docgen.Generate(root, dirB); err != nil {
		return fmt.Errorf("generate run B: %w", err)
	}
	return determinismDiffMDX(dirA, dirB)
}

// determinismDiffMDX compares the .mdx files of two directories: first the set
// of filenames, then the raw bytes of each file.
func determinismDiffMDX(dirA, dirB string) error {
	a, err := determinismReadMDX(dirA)
	if err != nil {
		return err
	}
	b, err := determinismReadMDX(dirB)
	if err != nil {
		return err
	}

	namesA := determinismSortedKeys(a)
	namesB := determinismSortedKeys(b)
	if !reflect.DeepEqual(namesA, namesB) {
		return fmt.Errorf("mdx file sets differ: runA=%v runB=%v", namesA, namesB)
	}
	for _, name := range namesA {
		if !bytes.Equal(a[name], b[name]) {
			return fmt.Errorf("byte mismatch for %q (runA=%d bytes, runB=%d bytes)",
				name, len(a[name]), len(b[name]))
		}
	}
	return nil
}

// determinismReadMDX reads every .mdx file in dir into a name -> bytes map.
// Non-.mdx entries (e.g. the hand-maintained meta.json) are ignored.
func determinismReadMDX(dir string) (map[string][]byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("read dir %q: %w", dir, err)
	}
	out := make(map[string][]byte)
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".mdx" {
			continue
		}
		data, readErr := os.ReadFile(filepath.Join(dir, e.Name()))
		if readErr != nil {
			return nil, fmt.Errorf("read %q: %w", e.Name(), readErr)
		}
		out[e.Name()] = data
	}
	return out, nil
}

func determinismSortedKeys(m map[string][]byte) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// determinismTreeSpec wraps a synthetic cobra command tree and implements
// testing/quick.Generator so quick.Check can drive the property across many
// generated Command_Tree inputs.
type determinismTreeSpec struct {
	root *cobra.Command
}

// Generate satisfies testing/quick.Generator.
func (determinismTreeSpec) Generate(rnd *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(determinismTreeSpec{root: determinismBuildTree(rnd)})
}

// determinismBuildTree builds a random but valid cobra command tree (1..3 deep,
// 1..3 children per node) whose command names always slugify to a non-empty
// value so docgen.Generate never short-circuits on an empty slug.
func determinismBuildTree(rnd *rand.Rand) *cobra.Command {
	root := &cobra.Command{Use: "root", Short: "synthetic root"}
	depth := 1 + rnd.Intn(3)       // 1..3 levels
	maxChildren := 1 + rnd.Intn(3) // 1..3 children per node

	var add func(parent *cobra.Command, d int)
	add = func(parent *cobra.Command, d int) {
		if d <= 0 {
			return
		}
		n := 1 + rnd.Intn(maxChildren)
		for i := 0; i < n; i++ {
			child := determinismBuildCommand(rnd)
			parent.AddCommand(child)
			add(child, d-1)
		}
	}
	add(root, depth)
	return root
}

func determinismBuildCommand(rnd *rand.Rand) *cobra.Command {
	cmd := &cobra.Command{
		Use:   determinismRandomToken(rnd, 1, 8),
		Short: determinismRandomShort(rnd),
	}
	if rnd.Intn(2) == 0 {
		cmd.Long = determinismRandomShort(rnd) + " " + determinismRandomShort(rnd)
	}
	if rnd.Intn(2) == 0 {
		cmd.Example = cmd.Use + " " + determinismRandomToken(rnd, 1, 6)
	}
	determinismAddFlags(rnd, cmd)
	return cmd
}

func determinismAddFlags(rnd *rand.Rand, cmd *cobra.Command) {
	count := rnd.Intn(4) // 0..3 flags
	seen := map[string]struct{}{}
	for i := 0; i < count; i++ {
		name := determinismRandomToken(rnd, 2, 8)
		if _, dup := seen[name]; dup {
			continue
		}
		seen[name] = struct{}{}
		if rnd.Intn(2) == 0 {
			cmd.Flags().String(name, determinismRandomToken(rnd, 0, 5), determinismRandomShort(rnd))
		} else {
			cmd.Flags().Bool(name, rnd.Intn(2) == 0, determinismRandomShort(rnd))
		}
	}
}

// determinismRandomToken returns a lowercase-letter token of length within
// [lo, hi]. Letters guarantee a non-empty slug for command names.
func determinismRandomToken(rnd *rand.Rand, lo, hi int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	n := lo
	if hi > lo {
		n = lo + rnd.Intn(hi-lo+1)
	}
	if n <= 0 {
		return ""
	}
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rnd.Intn(len(letters))]
	}
	return string(b)
}

func determinismRandomShort(rnd *rand.Rand) string {
	words := 1 + rnd.Intn(4)
	parts := make([]string, words)
	for i := range parts {
		parts[i] = determinismRandomToken(rnd, 2, 7)
	}
	return strings.TrimSpace(strings.Join(parts, " "))
}
