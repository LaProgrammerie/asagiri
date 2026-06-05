// Feature: audit-coherence-consolidation, Property 21
//
// Property 21 — Alignement des ensembles de chemins de locales (R8 — AUD-004).
// Pour chaque locale non `en` (`fr`, `de`, `es`), l'ensemble des chemins de
// pages sous `docs-site/content/docs/<loc>/` est égal à celui de `en` PRIVÉ DE
// la référence CLI générée `en`-only (`en/cli/generated/`).
//
// La référence CLI générée vit uniquement sous `cli/generated/` et n'est jamais
// traduite ; elle est donc exclue de la comparaison sur chaque locale (y compris
// `en`). Toute divergence de chemins HORS `cli/generated/` constitue un vrai
// désalignement de locale et fait échouer la propriété en listant les pages
// concernées.
//
// **Validates: Requirements 8.3**
package cli

import (
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"testing/quick"
)

// Constantes structurelles de la vérification d'alignement des locales. Noms
// préfixés `loc84` (tâche 8.4) pour éviter toute collision avec les autres tests
// du package `cli` exécutés en parallèle.
const (
	loc84DocsRel         = "docs-site/content/docs"
	loc84BaseLocale      = "en"
	loc84GeneratedPrefix = "cli/generated/"
	loc84PageExt         = ".mdx"
)

// loc84Locales est l'ensemble des locales non `en` devant s'aligner sur `en`
// (privé de `cli/generated/`).
var loc84Locales = []string{"fr", "de", "es"}

// loc84RepoRoot remonte depuis le répertoire de test jusqu'au répertoire
// contenant `go.mod` (racine du dépôt), où vit `docs-site/`.
func loc84RepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("loc84: os.Getwd : %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("loc84: go.mod introuvable en remontant depuis %q", dir)
		}
		dir = parent
	}
}

// loc84PagePaths retourne l'ensemble (set) des chemins relatifs de pages `.mdx`
// sous `docs-site/content/docs/<locale>/`, séparateurs normalisés en `/`. Les
// fichiers non `.mdx` (ex. `meta.json`) ne sont pas des pages et sont ignorés.
func loc84PagePaths(t *testing.T, root, locale string) map[string]bool {
	t.Helper()
	base := filepath.Join(root, filepath.FromSlash(loc84DocsRel), locale)
	set := make(map[string]bool)
	err := filepath.WalkDir(base, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(path) != loc84PageExt {
			return nil
		}
		rel, relErr := filepath.Rel(base, path)
		if relErr != nil {
			return relErr
		}
		set[filepath.ToSlash(rel)] = true
		return nil
	})
	if err != nil {
		t.Fatalf("loc84: parcours des pages de la locale %q : %v", locale, err)
	}
	return set
}

// loc84WithoutGenerated retourne une copie du set privée des chemins sous
// `cli/generated/` (référence CLI générée `en`-only, non traduite).
func loc84WithoutGenerated(in map[string]bool) map[string]bool {
	out := make(map[string]bool, len(in))
	for p := range in {
		if strings.HasPrefix(p, loc84GeneratedPrefix) {
			continue
		}
		out[p] = true
	}
	return out
}

// loc84Difference retourne les clés présentes dans `a` mais absentes de `b`.
func loc84Difference(a, b map[string]bool) []string {
	var diff []string
	for p := range a {
		if !b[p] {
			diff = append(diff, p)
		}
	}
	sort.Strings(diff)
	return diff
}

// TestLocalesPathAlignmentProperty vérifie la Property 21 (P21) : pour chaque
// locale non `en`, l'ensemble des chemins de pages HORS `cli/generated/` est
// strictement égal à celui de `en` HORS `cli/generated/`.
//
// La propriété est exprimée comme une universelle sur l'espace des locales
// {fr, de, es} : un générateur déterministe (`testing/quick`) tire une locale
// au hasard à chaque itération (≥ 100 itérations) et la propriété doit tenir
// pour toutes. L'issue est déterministe (le système de fichiers ne varie pas),
// mais la formulation reste celle d'un property-based test sur l'espace d'entrée.
//
// **Validates: Requirements 8.3**
func TestLocalesPathAlignmentProperty(t *testing.T) {
	root := loc84RepoRoot(t)

	// Référence : pages de `en` privées de la CLI générée `en`-only.
	expected := loc84WithoutGenerated(loc84PagePaths(t, root, loc84BaseLocale))
	if len(expected) == 0 {
		t.Fatalf("loc84: aucune page trouvée pour la locale de référence %q (chemin docs incorrect ?)", loc84BaseLocale)
	}

	// Pré-calcul des ensembles de chemins (hors `cli/generated/`) par locale.
	localeSets := make(map[string]map[string]bool, len(loc84Locales))
	for _, loc := range loc84Locales {
		localeSets[loc] = loc84WithoutGenerated(loc84PagePaths(t, root, loc))
	}

	// Propriété P21 : ∀ loc ∈ {fr, de, es}, pages(loc)\generated == pages(en)\generated.
	property := func(pick uint16) bool {
		loc := loc84Locales[int(pick)%len(loc84Locales)]
		got := localeSets[loc]

		missing := loc84Difference(expected, got) // dans `en` mais absentes de `loc`
		extra := loc84Difference(got, expected)   // dans `loc` mais absentes de `en`
		if len(missing) == 0 && len(extra) == 0 {
			return true
		}

		// Vrai désalignement HORS `cli/generated/` : on le signale explicitement
		// (pages manquantes / en trop) plutôt que de le masquer.
		t.Errorf("Property 21 — locale %q désalignée (hors %s)\n  pages manquantes (présentes en %q) : %v\n  pages en trop (absentes en %q)   : %v",
			loc, loc84GeneratedPrefix, loc84BaseLocale, missing, loc84BaseLocale, extra)
		return false
	}

	if err := quick.Check(property, &quick.Config{MaxCount: 200}); err != nil {
		t.Fatalf("Property 21 (alignement des ensembles de chemins de locales) violée : %v", err)
	}
}
