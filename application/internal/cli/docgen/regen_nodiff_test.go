package docgen_test

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/cli"
	"github.com/LaProgrammerie/asagiri/application/internal/cli/docgen"
)

// Feature: audit-coherence-consolidation
//
// Example test proving Regeneration_Without_Diff for the committed CLI
// reference (R1 — AUD-001/AUD-002). Regenerating the live cobra command tree
// into a temp dir must yield byte-for-byte identical *.mdx pages to the
// committed corpus under docs-site/content/docs/en/cli/generated/.
//
// meta.json is excluded on both sides: it is hand-maintained and never emitted
// by docgen, so comparing only *.mdx files avoids a perpetual false positive
// (present := { e | ext(e) == ".mdx" }).
//
// Requirements: 1.3, 1.6, 8.1

// nodiffGeneratedRel is the repo-relative path to the committed CLI reference.
const nodiffGeneratedRel = "docs-site/content/docs/en/cli/generated"

// nodiffRunsSiblingLink is the committed fratrie-link form (AUD-002). The design
// quotes it as "> - [Runs](./runs.mdx)"; the leading "> " is Markdown blockquote
// decoration, the committed line itself is the list item below.
const nodiffRunsSiblingLink = "- [Runs](./runs.mdx)"

// TestNodiffRegenerationWithoutDiff regenerates the live command tree into a
// temporary directory and asserts there is no difference (missing, orphan, or
// byte-divergent *.mdx) against the committed reference. On the corrected repo
// this passes (exit 0); any drift lists the offending files (Requirements 1.3,
// 8.1).
func TestNodiffRegenerationWithoutDiff(t *testing.T) {
	committedDir := filepath.Join(repoRoot(t), filepath.FromSlash(nodiffGeneratedRel))

	tmpDir := t.TempDir()
	root := nodiffRootForDocs()
	if err := docgen.Generate(root, tmpDir); err != nil {
		t.Fatalf("generate into temp dir: %v", err)
	}

	regen := nodiffReadMDX(t, tmpDir)
	committed := nodiffReadMDX(t, committedDir)

	var missing, orphan, divergent []string
	for name, want := range regen {
		got, ok := committed[name]
		if !ok {
			// Regen produces it but the committed corpus lacks it: a command was
			// added without regenerating the docs.
			missing = append(missing, name)
			continue
		}
		if !bytes.Equal(want, got) {
			divergent = append(divergent, name)
		}
	}
	for name := range committed {
		if _, ok := regen[name]; !ok {
			// Committed corpus has a page the live tree no longer produces.
			orphan = append(orphan, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(orphan)
	sort.Strings(divergent)

	if len(missing) > 0 || len(orphan) > 0 || len(divergent) > 0 {
		t.Fatalf("Regeneration_Without_Diff violated (run: go run ./application/cmd/asa docs generate-cli --output %s)\n"+
			"  missing from committed: %v\n"+
			"  orphaned in committed:  %v\n"+
			"  byte-divergent:         %v",
			nodiffGeneratedRel, missing, orphan, divergent)
	}
}

// TestNodiffRunsSiblingLinkPresent asserts the AUD-002 fix: at least one sibling
// page of `runs` carries the fratrie link to runs.mdx (Requirement 1.6).
func TestNodiffRunsSiblingLinkPresent(t *testing.T) {
	committedDir := filepath.Join(repoRoot(t), filepath.FromSlash(nodiffGeneratedRel))

	pages := nodiffReadMDX(t, committedDir)
	if _, ok := pages["runs.mdx"]; !ok {
		t.Fatalf("expected committed runs.mdx to exist (AUD-001) under %s", nodiffGeneratedRel)
	}

	var withLink []string
	for name, body := range pages {
		if name == "runs.mdx" {
			continue // a page does not link to itself
		}
		if strings.Contains(string(body), nodiffRunsSiblingLink) {
			withLink = append(withLink, name)
		}
	}
	if len(withLink) == 0 {
		t.Fatalf("no sibling page of `runs` contains fratrie link %q (AUD-002) under %s",
			nodiffRunsSiblingLink, nodiffGeneratedRel)
	}
}

// nodiffRootForDocs builds the command tree exactly as the committed reference
// was produced: by the `asa docs generate-cli` binary path. Cobra lazily injects
// the default `help` and `completion` commands when a command is executed, so
// the shipped corpus carries fratrie links to them. We reproduce that here by
// initializing those default commands; otherwise a bare RootCommand() would
// report ~50 spurious "byte-divergent" pages (missing Completion/Help links).
// docgen never writes completion.mdx/help.mdx (they are skipped), so only their
// sibling links matter.
func nodiffRootForDocs() *cobra.Command {
	root := cli.RootCommand()
	root.InitDefaultHelpCmd()
	root.InitDefaultCompletionCmd()
	return root
}

// nodiffReadMDX reads every *.mdx file in dir into a name→bytes map, ignoring
// subdirectories and non-mdx artefacts such as the hand-maintained meta.json.
func nodiffReadMDX(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("read dir %s: %v", dir, err)
	}
	out := make(map[string][]byte, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".mdx" {
			continue
		}
		data, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			t.Fatalf("read %s: %v", e.Name(), err)
		}
		out[e.Name()] = data
	}
	return out
}
