// Package docgen generates MDX reference pages from a cobra command tree.
package docgen

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"
)

// Generate walks all non-root commands under root and writes one .mdx file per command to outDir.
func Generate(root *cobra.Command, outDir string) error {
	if root == nil {
		return fmt.Errorf("docgen: root command is nil")
	}
	if err := os.MkdirAll(outDir, 0o750); err != nil {
		return fmt.Errorf("docgen: mkdir output: %w", err)
	}
	if err := removeStaleMDX(outDir); err != nil {
		return err
	}

	var errs []error
	eachSortedCommand(root, func(cmd *cobra.Command) {
		rel := commandPathRelative(root, cmd)
		slug := Slug(rel)
		if slug == "" {
			errs = append(errs, fmt.Errorf("docgen: empty slug for relative path %q", rel))
			return
		}
		doc, err := buildMDX(rel, slug, cmd, root)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", rel, err))
			return
		}
		path := filepath.Join(outDir, slug+".mdx")
		if writeErr := os.WriteFile(path, doc, 0o640); writeErr != nil {
			errs = append(errs, fmt.Errorf("docgen: write %q: %w", path, writeErr))
		}
	})
	return errors.Join(errs...)
}

// CommandPathsWithoutRoot returns space-separated cobra-relative paths like "inspect symbol".
// Paths are sorted for stable output.
func CommandPathsWithoutRoot(root *cobra.Command) []string {
	if root == nil {
		return nil
	}
	var paths []string
	eachSortedCommand(root, func(cmd *cobra.Command) {
		paths = append(paths, commandPathRelative(root, cmd))
	})
	sort.Strings(paths)
	return paths
}

// Slug maps a relative path ("cost report") to a deterministic filename slug ("cost-report").
func Slug(relativeSpaceSeparated string) string {
	return slugify(strings.Join(strings.Fields(relativeSpaceSeparated), "-"))
}

func eachSortedCommand(root *cobra.Command, fn func(*cobra.Command)) {
	var walk func(*cobra.Command)
	walk = func(parent *cobra.Command) {
		subs := slicesClone(parent.Commands())
		sort.Slice(subs, func(i, j int) bool { return subs[i].Name() < subs[j].Name() })
		for _, sub := range subs {
			if sub.Hidden || shouldSkipCommand(root, sub) {
				continue
			}
			fn(sub)
			walk(sub)
		}
	}
	walk(root)
}

func slicesClone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	out := make([]T, len(s))
	copy(out, s)
	return out
}

func commandPathRelative(root *cobra.Command, cmd *cobra.Command) string {
	var parts []string
	for cur := cmd; cur != root && cur != nil; cur = cur.Parent() {
		parts = append(parts, cur.Name())
	}
	sortReverseStrings(parts)
	return strings.Join(parts, " ")
}

func sortReverseStrings(s []string) {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
}

func slugify(raw string) string {
	var b strings.Builder
	prevDash := true // suppress leading hyphen
	raw = strings.ToLower(raw)
	for _, r := range raw {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		case r == ' ', r == '_' || r == '-' || r == '/':
			if !prevDash && b.Len() > 0 {
				b.WriteRune('-')
				prevDash = true
			}
		default:
			// drop placeholders like <feature>; digits/letters handled above.
		}
	}
	out := strings.Trim(b.String(), "-")
	out = collapseDuplicateHyphens(out)
	return out
}

func collapseDuplicateHyphens(s string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		if r == '-' {
			if prevHyphen || b.Len() == 0 {
				continue
			}
			b.WriteRune('-')
			prevHyphen = true
			continue
		}
		prevHyphen = false
		b.WriteRune(r)
	}
	return strings.TrimSuffix(b.String(), "-")
}

func buildMDX(rel, slug string, cmd *cobra.Command, root *cobra.Command) ([]byte, error) {
	title := humanizeCliPath(rel)
	desc := englishDescription(rel, slug, cmd)

	fm := struct {
		Title       string `yaml:"title"`
		Description string `yaml:"description"`
	}{
		Title:       title + " · Asagiri CLI",
		Description: desc,
	}
	var fmBuf bytes.Buffer
	if _, err := fmBuf.WriteString("---\n"); err != nil {
		return nil, err
	}
	enc := yaml.NewEncoder(&fmBuf)
	enc.SetIndent(2)
	if err := enc.Encode(fm); err != nil {
		return nil, err
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	if _, err := fmBuf.WriteString("---"); err != nil {
		return nil, err
	}

	var body strings.Builder
	body.Write(fmBuf.Bytes())
	body.WriteString("\n\n")

	_, _ = fmt.Fprintf(&body, "CLI path: **`%s %s`**\n\n", root.Name(), rel)

	body.WriteString("## When to use\n\n")
	body.WriteString(strings.TrimSpace(englishWhen(rel, cmd)))
	body.WriteString("\n\n")

	body.WriteString("## Usage\n\n")
	body.WriteString("```bash\n")
	body.WriteString(strings.TrimSpace(strings.TrimSuffix(usageSnippet(cmd), "\n")))
	body.WriteString("\n```\n\n")

	body.WriteString("## Options\n\n")
	body.WriteString(optionsSection(cmd))

	body.WriteString("## Examples\n\n")
	body.WriteString(examplesSection(cmd, root))

	body.WriteString("## What happens internally\n\n")
	body.WriteString(strings.TrimSpace(englishInternals(rel)))
	body.WriteString("\n\n")

	body.WriteString("## Failure modes\n\n")
	body.WriteString(strings.TrimSpace(englishFailureModes(rel)))
	body.WriteString("\n\n")

	body.WriteString("## Related commands\n\n")
	body.WriteString(strings.TrimSpace(relatedCommandsSection(cmd, root)))

	return []byte(strings.TrimSuffix(body.String(), "\n") + "\n"), nil
}

func englishDescription(rel, slug string, cmd *cobra.Command) string {
	if s := strings.TrimSpace(cmd.Long); s != "" {
		return sanitizeDescription(singleLine(removeAnglePlaceholders(headParagraph(s))), slug)
	}
	if s := strings.TrimSpace(cmd.Short); s != "" {
		return sanitizeDescription(singleLine(removeAnglePlaceholders(s)), slug)
	}
	return fmt.Sprintf("Asagiri CLI reference for %s.", englishPathPhrase(rel))
}

func sanitizeDescription(s, slug string) string {
	out := strings.TrimSpace(s)
	if out == "" {
		return fmt.Sprintf("Asagiri CLI reference for `%s`.", slug)
	}
	if len([]rune(out)) > 220 {
		rs := []rune(out)
		out = strings.TrimSpace(string(rs[:217])) + "…"
	}
	return out
}

func headParagraph(s string) string {
	s = strings.TrimSpace(s)
	if idx := strings.IndexByte(s, '\n'); idx >= 0 {
		return strings.TrimSpace(s[:idx])
	}
	return s
}

func singleLine(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func removeAnglePlaceholders(s string) string {
	var b strings.Builder
	in := false
	for _, r := range s {
		if r == '<' {
			in = true
			continue
		}
		if in {
			if r == '>' {
				in = false
			}
			continue
		}
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func humanizeCliPath(rel string) string {
	parts := strings.Fields(rel)
	for i := range parts {
		parts[i] = capitalizeASCII(parts[i])
	}
	return strings.Join(parts, " ")
}

func capitalizeASCII(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func archetype(rel string) string {
	top := firstToken(rel)
	switch top {
	case "docs":
		return "docs"
	case "version":
		return "version"
	case "inspect":
		return "inspect"
	case "cost":
		return "cost"
	case "mcp":
		return "mcp"
	case "doctor", "init":
		return "bootstrap"
	case "spec", "plan", "enrich", "dev", "verify", "review", "report", "clean", "pr":
		return "workflow"
	case "resume", "status":
		return "runops"
	case "index":
		return "repoindex"
	default:
		return "orchestration"
	}
}

func firstToken(rel string) string {
	rel = strings.TrimSpace(rel)
	if idx := strings.IndexByte(rel, ' '); idx >= 0 {
		return rel[:idx]
	}
	return rel
}

func englishPathPhrase(rel string) string {
	if rel == "" {
		return "this command group"
	}
	segs := strings.Fields(rel)
	quoted := make([]string, 0, len(segs))
	for _, s := range segs {
		quoted = append(quoted, fmt.Sprintf("`%s`", s))
	}
	return "the " + strings.Join(quoted, " ")
}

func englishWhen(rel string, cmd *cobra.Command) string {
	var block string
	switch archetype(rel) {
	case "docs":
		block = strings.TrimSpace(`
Use **docs** commands when refreshing machine-generated reference material—for example, syncing this MDX corpus with shipping CLI behaviour ahead of docs-site releases.

**generate-cli** is safe to run repeatedly; outputs are regenerated wholesale under the **--output** directory.`)
	case "inspect":
		block = strings.TrimSpace(`
Use **inspect** when you need fast, offline answers about symbols, neighbouring tests, or a compact snapshot of unstaged Git changes without spawning remote agents.
`)
	case "cost":
		block = strings.TrimSpace(`
Use **cost** when reconciling SQLite-backed telemetry with pricing tables declared in **.asagiri/config.yaml** or when slicing spend over a bounded recency window.
`)
	case "mcp":
		block = strings.TrimSpace(`
Use **mcp serve** whenever an IDE expects a bundled Model Context Protocol stdio shim that mirrors **.asagiri** configuration.
`)
	case "workflow":
		block = strings.TrimSpace(`
Workflow commands advance tracked features across spec → plan → implementation → verification → reporting guardrails before opening a merge request.
`)
	case "bootstrap":
		block = strings.TrimSpace(`
Bootstrap commands scaffold or reconcile the Asagiri working tree so subsequent workflow nodes can reliably resolve configuration and repository metadata.
`)
	case "runops":
		block = strings.TrimSpace(`
Operational commands summarise historical runs or unblock automation by resuming deterministic workflow bookkeeping without rewriting task definitions manually.
`)
	case "repoindex":
		block = strings.TrimSpace(`
Indexing keeps embedded search artefacts aligned whenever significant repository reorganisation threatens stale lookup tables.
`)
	default:
		block = strings.TrimSpace(`
Orchestration commands coordinate higher-order Asagiri workloads such as ingestion, syncing, backlog triage, investigative passes, hydrated context payloads, coarse estimates, and workspace-specific automation.
`)
	}
	if short := strings.TrimSpace(cmd.Short); short != "" {
		block += "\n\n> Original CLI synopsis: " + short + "\n"
	}
	return strings.TrimSpace(block)
}

func usageSnippet(cmd *cobra.Command) string {
	return strings.TrimSpace(cmd.CommandPath() + " [flags]")
}

func examplesSection(cmd *cobra.Command, root *cobra.Command) string {
	if ex := strings.TrimSpace(cmd.Example); ex != "" {
		return "```bash\n" + strings.TrimSpace(ex) + "\n```\n\n"
	}
	tail := strings.TrimPrefix(strings.TrimSpace(cmd.CommandPath()), root.Name()+" ")
	snippet := "```bash\n" + strings.TrimSpace(root.Name()+" "+tail) + "\n```\n\n"
	note := "> Wire richer samples through **cobra.Command.Example** when the happy path warrants copy/paste-friendly automation beyond this scaffold.\n\n"
	return snippet + note
}

func optionsSection(cmd *cobra.Command) string {
	rows := collectFlagRows(cmd)
	if len(rows) == 0 {
		return "_No CLI flags are registered on this command node (cobra defaults notwithstanding)._\n\n"
	}

	var buf strings.Builder
	buf.WriteString("| Flag | Scope | Default | Description |\n| --- | --- | --- | --- |\n")

	for _, fr := range rows {
		_, _ = fmt.Fprintf(&buf, "| `%s` | %s | %s | %s |\n",
			fr.display, fr.scope, mdCell(fr.defValueQuoted), mdCell(fr.usageRaw))
	}
	buf.WriteByte('\n')
	return buf.String()
}

type assembledFlagRow struct {
	name           string // dedupe key
	display        string // pretty command-line token
	scope          string // Local / Inherited
	defValueQuoted string // includes inline backticks for markdown
	usageRaw       string
}

func collectFlagRows(cmd *cobra.Command) []assembledFlagRow {
	local := map[string]*pflag.Flag{}
	cmd.NonInheritedFlags().VisitAll(func(f *pflag.Flag) {
		if shouldSkipHelpFlag(f) {
			return
		}
		local[f.Name] = f
	})

	var merged []assembledFlagRow
	for _, name := range sortedFlagNames(local) {
		f := local[name]
		merged = append(merged, newFlagPresentation(f, "Local"))
	}

	cmd.InheritedFlags().VisitAll(func(f *pflag.Flag) {
		if shouldSkipHelpFlag(f) {
			return
		}
		if _, dup := local[f.Name]; dup {
			return
		}
		merged = append(merged, newFlagPresentation(f, "Inherited"))
	})

	sort.Slice(merged, func(i, j int) bool {
		return strings.ToLower(merged[i].name) < strings.ToLower(merged[j].name)
	})
	return merged
}

func sortedFlagNames(flags map[string]*pflag.Flag) []string {
	names := make([]string, 0, len(flags))
	for n := range flags {
		names = append(names, n)
	}
	sort.Strings(names)
	return names
}

func newFlagPresentation(f *pflag.Flag, scope string) assembledFlagRow {
	display := "--" + strings.TrimPrefix(f.Name, "--")
	sh := strings.TrimSpace(f.Shorthand)
	if sh != "" {
		display += " / -" + sh
	}
	defVal := strings.TrimSpace(f.DefValue)
	defQuoted := "`(empty)`"
	if defVal != "" {
		defQuoted = "`" + strings.ReplaceAll(defVal, "`", "'") + "`"
	}
	return assembledFlagRow{
		name:           f.Name,
		display:        display,
		scope:          scope,
		defValueQuoted: defQuoted,
		usageRaw:       strings.TrimSpace(f.Usage),
	}
}

func removeStaleMDX(outDir string) error {
	entries, err := os.ReadDir(outDir)
	if err != nil {
		return fmt.Errorf("docgen: read output dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".mdx") {
			continue
		}
		if err := os.Remove(filepath.Join(outDir, e.Name())); err != nil {
			return fmt.Errorf("docgen: remove stale %q: %w", e.Name(), err)
		}
	}
	return nil
}

func shouldSkipCommand(root, cmd *cobra.Command) bool {
	rel := commandPathRelative(root, cmd)
	if rel == "help" {
		return true
	}
	return strings.HasPrefix(rel, "completion")
}

func shouldSkipHelpFlag(f *pflag.Flag) bool {
	if f.Hidden {
		return true
	}
	name := strings.TrimPrefix(strings.ToLower(strings.TrimSpace(f.Name)), "--")
	switch name {
	case "help":
		return true
	default:
		return false
	}
}

func mdCell(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	s = strings.ReplaceAll(s, "|", "\\|")
	return s
}

func englishInternals(rel string) string {
	switch archetype(rel) {
	case "docs":
		if strings.TrimSpace(restAfter(rel, "docs")) == "generate-cli" {
			return strings.TrimSpace(`
- Traverses the live **cobra.Command** tree returned by **cli.RootCommand()** depth-first while sorting immediate children alphabetically by name for reproducible filenames.
- Renders deterministic MDX with YAML frontmatter and tables sourced from cobra metadata plus conservative narration heuristics.
- Writes only underneath the caller-provided **--output** destination; callers should ensure the folder is writable and version-controlled intentionally.
`)
		}
		return strings.TrimSpace(`
- Holds documentation-adjacent automation that should remain orthogonal to primary workflow commands (does not mutate SQLite artefacts).
`)

	case "version":
		return strings.TrimSpace(`
- Emits compile-time versioning metadata from **internal/version.Version** without mutating workspace files.
`)

	case "inspect":
		switch firstToken(restAfter(rel, "inspect")) {
		case "symbol":
			return strings.TrimSpace(`
- Loads contextual services via **loadContext**, runs the investigative runner, then prints filenames where the queried Go identifier appears.
`)
		case "tests":
			return strings.TrimSpace(`
- Executes the investigative scaffolding so the supplied filesystem path maps to heuristic related test suites.
`)
		case "diff":
			return strings.TrimSpace(`
- Spawns **git diff --stat** scoped with **git -C** at the Asagiri-managed repository root, truncating very large summaries locally.
`)
		default:
			return strings.TrimSpace(`
- Hosts read-only inspectors that intentionally avoid spawning remote delegation agents so local workflows stay deterministic.
`)
		}

	case "cost":
		switch firstToken(restAfter(rel, "cost")) {
		case "report":
			return strings.TrimSpace(`
- Parses the **--since** shorthand window, opens the SQLite telemetry store associated with **$PWD**, and aggregates totals through **telemetry.SummarizeSince**.
`)
		case "models":
			return strings.TrimSpace(`
- Streams configured pricing slabs and YAML-defined model mappings to stdout strictly for auditing; no datastore mutation occurs during display.
`)
		default:
			return strings.TrimSpace(`
- Focuses on interpreting persisted telemetry and declared pricing artefacts for finance debugging.
`)
		}

	case "mcp":
		return strings.TrimSpace(`
- Validates configuration guards before multiplexing MCP stdio traffic through **internal/mcp** adapters rooted at the resolved repository path.
`)

	case "bootstrap":
		switch firstToken(rel) {
		case "init":
			return strings.TrimSpace(`
- Writes Asagiri scaffolding (directories plus manifests) that align repositories with onboarding expectations documented for contributors.
`)
		case "doctor":
			return strings.TrimSpace(`
- Probes requisite binaries and configuration coherence—including MCP switches and caches—to surface actionable diagnostics before workflows run hot.
`)
		default:
			return strings.TrimSpace(`
- Validates checkout readiness unless otherwise performing explicit scaffold mutations guarded by narrower subcommands.
`)
		}

	case "workflow":
		return strings.TrimSpace(`
- Validates CLI arity, honours the global **--dry-run** switch, resolves repository metadata, hydrates SQLite-backed bookkeeping, then calls façade orchestration wrappers for the requested lifecycle stage.
- Persists deterministic run identifiers so **status** and **resume** can continue automation deterministically afterward.
`)

	case "runops":
		switch firstToken(rel) {
		case "status":
			return strings.TrimSpace(`
- Queries recent SQLite rows through the workflow façade and caps emitted records to preserve readable CLI ergonomics.
`)
		case "resume":
			return strings.TrimSpace(`
- Reloads workflow state for the run identifier; **--execute** chains remaining steps (dry-run rehearsal when global **--dry-run** is set, real agents otherwise); **--force** may re-run a succeeded step.
`)
		default:
			return strings.TrimSpace(`
- Operates solely on persisted metadata without authoring pull requests implicitly.
`)
		}

	case "repoindex":
		return strings.TrimSpace(`
- Crawls repositories to refresh Asagiri ingestion indexes referenced by tooling that depends on deterministic filesystem snapshots.
`)

	default:
		return strings.TrimSpace(`
- Hydrates contextual services resolved from **$PWD**, coordinating git-bound automation, ingestion pipelines, syncing, backlog triage, investigations, MCP hooks, contextual hydration, coarse estimates, and similar orchestrations for this subtree.
`)
	}
}

func restAfter(rel, prefixSpace string) string {
	rTok := strings.Fields(rel)
	ps := strings.Fields(prefixSpace)
	if len(ps) > len(rTok) {
		return ""
	}
	for i := range ps {
		if rTok[i] != ps[i] {
			return strings.Join(rTok[i:], " ")
		}
	}
	if len(rTok) <= len(ps) {
		return ""
	}
	return strings.TrimSpace(strings.Join(rTok[len(ps):], " "))
}

func englishFailureModes(rel string) string {
	write := func(b *strings.Builder, line string) {
		_, _ = fmt.Fprintf(b, "- %s\n", line)
	}
	var b strings.Builder
	write(&b, "Cobra rejects unknown flags or arity mismatches before any repository side effects occur.")
	write(&b, "Asagiri cannot resolve scaffolding when the command is invoked outside an initialised checkout.")

	switch archetype(rel) {
	case "docs":
		write(&b, "Filesystem ACLs can deny writes under **--output**, leaving partial artefacts if interrupted mid-write.")
	case "inspect":
		write(&b, "Git probing fails when binaries are missing or repositories are shallow or incomplete.")
		write(&b, "Symbol lookups can legitimately return empty output whenever indexes lag behind edits.")
	case "cost":
		write(&b, "SQLite contention or corrupted telemetry prevent summarisation queries from completing.")
		write(&b, "Malformed **--since** payloads fail fast before aggregates are queried.")
	case "mcp":
		write(&b, "Disabled MCP integrations stop early with guidance drawn from **.asagiri** configuration diagnostics.")
	case "version":
		write(&b, "Version strings may read as placeholders when distribution pipelines omit linkage metadata—a rare tooling misconfiguration.")
	case "workflow", "runops", "repoindex":
		write(&b, "Delegated subprocesses (**git**, MCP hosts, editor agents, etc.) can exit non-zero and bubble errors verbatim.")
	default:
		write(&b, "Downstream automation may fail whenever credentials, network endpoints, or cached prerequisites drift from expectations.")
	}

	return strings.TrimSpace(b.String())
}

func relatedCommandsSection(cmd *cobra.Command, root *cobra.Command) string {
	parent := cmd.Parent()
	if parent == nil {
		return "_No sibling relationships because the inspected command lacks a parent node._"
	}

	var lines []string
	for _, sibling := range parent.Commands() {
		if sibling.Hidden || sibling == cmd {
			continue
		}
		sibRel := commandPathRelative(root, sibling)
		link := fmt.Sprintf("- [%s](./%s.mdx)", humanizeCliPath(sibRel), Slug(sibRel))
		lines = append(lines, link)
	}
	sort.Strings(lines)
	if len(lines) == 0 {
		return "_No sibling commands at this depth (" + humanizeCliPath(commandPathRelative(root, cmd)) + ")._"
	}
	return strings.Join(lines, "\n")
}
