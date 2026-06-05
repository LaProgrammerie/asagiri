package docgen_test

// Feature: audit-coherence-consolidation, Property 3
// Design property: P3 — Regeneration_Check — exit ≠ 0 ssi différence, hors `meta.json`.
//
// For any pair (committed pages, regenerated pages) restricted to `.mdx` files,
// the Regeneration_Check reports a divergence (exit ≠ 0) if and only if there is
// at least one missing, orphan or byte-divergent file — and lists them. The
// presence of the hand-maintained `meta.json` file never triggers a divergence
// because the comparison only considers entries whose extension is `.mdx`
// (present = { e | ext(e) == ".mdx" }).
//
// Validates: Requirements 1.4, 1.5, 1.6

import (
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"testing/quick"
)

// regenCheckDiff is the result of comparing a regenerated (expected) MDX corpus
// against a committed directory, restricted to `.mdx` files. It mirrors the
// guardrail described in design R1: missing ∪ orphan ∪ diverged.
type regenCheckDiff struct {
	Missing  []string // expected `.mdx` absent from committed
	Orphan   []string // committed `.mdx` with no matching expected page
	Diverged []string // present in both but byte-different
}

// hasDivergence reports whether the Regeneration_Check should exit ≠ 0.
func (d regenCheckDiff) hasDivergence() bool {
	return len(d.Missing) > 0 || len(d.Orphan) > 0 || len(d.Diverged) > 0
}

// regenCheckReadMDX reads a directory and returns only its `.mdx` entries keyed
// by filename. Any non-`.mdx` entry (notably the hand-maintained `meta.json`)
// is excluded, so it can never be flagged as an orphan.
func regenCheckReadMDX(dir string) (map[string][]byte, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	out := make(map[string][]byte, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".mdx" {
			continue
		}
		b, readErr := os.ReadFile(filepath.Join(dir, e.Name()))
		if readErr != nil {
			return nil, readErr
		}
		out[e.Name()] = b
	}
	return out, nil
}

// regenCheckCompare implements the Regeneration_Check over two directories,
// considering only `.mdx` files (meta.json is ignored by construction).
func regenCheckCompare(expectedDir, committedDir string) (regenCheckDiff, error) {
	expected, err := regenCheckReadMDX(expectedDir)
	if err != nil {
		return regenCheckDiff{}, err
	}
	committed, err := regenCheckReadMDX(committedDir)
	if err != nil {
		return regenCheckDiff{}, err
	}

	var diff regenCheckDiff
	for name, want := range expected {
		got, ok := committed[name]
		if !ok {
			diff.Missing = append(diff.Missing, name)
			continue
		}
		if !bytes.Equal(want, got) {
			diff.Diverged = append(diff.Diverged, name)
		}
	}
	for name := range committed {
		if _, ok := expected[name]; !ok {
			diff.Orphan = append(diff.Orphan, name)
		}
	}
	sort.Strings(diff.Missing)
	sort.Strings(diff.Orphan)
	sort.Strings(diff.Diverged)
	return diff, nil
}

// regenCheckState describes how a base (regenerated) page is reflected in the
// synthetic committed directory.
type regenCheckState int

const (
	regenCheckIdentical regenCheckState = iota // committed copy matches byte-for-byte
	regenCheckDiverged                         // committed copy has different bytes
	regenCheckMissing                          // page absent from committed
)

type regenCheckFile struct {
	slug    string
	content []byte
	state   regenCheckState
}

type regenCheckOrphan struct {
	name    string
	content []byte
}

// regenCheckScenario is a synthetic divergence injection: a base set of
// regenerated `.mdx` pages, per-page committed states, extra orphan `.mdx`
// files, and an optional hand-maintained meta.json in the committed dir.
type regenCheckScenario struct {
	files       []regenCheckFile
	orphans     []regenCheckOrphan
	includeMeta bool
	metaContent []byte
}

func regenCheckRandBytes(r *rand.Rand) []byte {
	n := 1 + r.Intn(20)
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(33 + r.Intn(90)) // printable ASCII range, deterministic
	}
	return b
}

// Generate implements quick.Generator with a deterministic, constrained
// generator so the property explores the divergence space intelligently.
func (regenCheckScenario) Generate(r *rand.Rand, _ int) reflect.Value {
	sc := regenCheckScenario{}

	nFiles := r.Intn(6) // 0..5 base pages
	for i := 0; i < nFiles; i++ {
		sc.files = append(sc.files, regenCheckFile{
			slug:    fmt.Sprintf("cmd-%02d", i),
			content: regenCheckRandBytes(r),
			state:   regenCheckState(r.Intn(3)),
		})
	}

	nOrphans := r.Intn(3) // 0..2 committed-only pages
	for i := 0; i < nOrphans; i++ {
		sc.orphans = append(sc.orphans, regenCheckOrphan{
			name:    fmt.Sprintf("orphan-%02d.mdx", i),
			content: regenCheckRandBytes(r),
		})
	}

	sc.includeMeta = r.Intn(2) == 1
	sc.metaContent = regenCheckRandBytes(r)
	return reflect.ValueOf(sc)
}

// materialize writes the scenario to two directories and returns the oracle:
// the sets the Regeneration_Check is expected to report.
func (sc regenCheckScenario) materialize(expectedDir, committedDir string) (regenCheckDiff, error) {
	if err := os.MkdirAll(expectedDir, 0o750); err != nil {
		return regenCheckDiff{}, err
	}
	if err := os.MkdirAll(committedDir, 0o750); err != nil {
		return regenCheckDiff{}, err
	}

	var oracle regenCheckDiff
	for _, f := range sc.files {
		name := f.slug + ".mdx"
		// Every base page exists in the regenerated (expected) corpus.
		if err := os.WriteFile(filepath.Join(expectedDir, name), f.content, 0o640); err != nil {
			return regenCheckDiff{}, err
		}
		switch f.state {
		case regenCheckIdentical:
			if err := os.WriteFile(filepath.Join(committedDir, name), f.content, 0o640); err != nil {
				return regenCheckDiff{}, err
			}
		case regenCheckDiverged:
			// Appending a byte guarantees a byte-level difference.
			divergent := append(append([]byte{}, f.content...), '\n', 'X')
			if err := os.WriteFile(filepath.Join(committedDir, name), divergent, 0o640); err != nil {
				return regenCheckDiff{}, err
			}
			oracle.Diverged = append(oracle.Diverged, name)
		case regenCheckMissing:
			oracle.Missing = append(oracle.Missing, name)
		}
	}

	for _, o := range sc.orphans {
		if err := os.WriteFile(filepath.Join(committedDir, o.name), o.content, 0o640); err != nil {
			return regenCheckDiff{}, err
		}
		oracle.Orphan = append(oracle.Orphan, o.name)
	}

	if sc.includeMeta {
		// Hand-maintained file: present in committed, never emitted by docgen,
		// and must never be reported as a divergence.
		if err := os.WriteFile(filepath.Join(committedDir, "meta.json"), sc.metaContent, 0o640); err != nil {
			return regenCheckDiff{}, err
		}
	}

	sort.Strings(oracle.Missing)
	sort.Strings(oracle.Orphan)
	sort.Strings(oracle.Diverged)
	return oracle, nil
}

func regenCheckContains(list []string, target string) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}

func TestRegenCheckExitsNonZeroIffDivergenceExcludingMeta(t *testing.T) {
	base := t.TempDir()
	iter := 0

	property := func(sc regenCheckScenario) bool {
		iter++
		expectedDir := filepath.Join(base, fmt.Sprintf("exp-%05d", iter))
		committedDir := filepath.Join(base, fmt.Sprintf("com-%05d", iter))

		oracle, err := sc.materialize(expectedDir, committedDir)
		if err != nil {
			t.Errorf("materialize scenario: %v", err)
			return false
		}

		got, err := regenCheckCompare(expectedDir, committedDir)
		if err != nil {
			t.Errorf("regenCheckCompare: %v", err)
			return false
		}

		// exit ≠ 0 ssi différence : the computed divergence must match the oracle.
		if got.hasDivergence() != oracle.hasDivergence() {
			t.Errorf("divergence mismatch: got=%v oracle=%v (missing=%v orphan=%v diverged=%v)",
				got.hasDivergence(), oracle.hasDivergence(), got.Missing, got.Orphan, got.Diverged)
			return false
		}

		// The check must LIST the offending files exactly (1.5).
		if !reflect.DeepEqual(got.Missing, oracle.Missing) ||
			!reflect.DeepEqual(got.Orphan, oracle.Orphan) ||
			!reflect.DeepEqual(got.Diverged, oracle.Diverged) {
			t.Errorf("reported sets differ from oracle:\n got=%+v\n want=%+v", got, oracle)
			return false
		}

		// meta.json must never appear in any reported set, even when present.
		if regenCheckContains(got.Missing, "meta.json") ||
			regenCheckContains(got.Orphan, "meta.json") ||
			regenCheckContains(got.Diverged, "meta.json") {
			t.Errorf("meta.json wrongly reported as a divergence: %+v", got)
			return false
		}

		return true
	}

	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(3)), // deterministic generators
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 3 (Regeneration_Check) failed: %v", err)
	}
}
