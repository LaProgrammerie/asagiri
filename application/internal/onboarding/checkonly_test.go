package onboarding_test

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

// Feature: audit-coherence-consolidation, Property 18: `--check-only` n'écrit rien.
//
// P18 (design.md) : For any dépôt, l'exécution de l'Onboarding_Wizard en mode
// `--check-only` n'entraîne la création, la modification ni la suppression
// d'aucun fichier. On prouve la propriété en prenant un snapshot octet-à-octet
// (chemins + contenu) de l'arborescence avant et après l'exécution du chemin
// check-only sur des états de dépôt initiaux variés et déterministes.
//
// Validates: Requirements 7.3

// p18RepoState décrit un état initial de dépôt généré de façon déterministe.
type p18RepoState struct {
	HasGoMod          bool
	HasCastor         bool
	HasGitignore      bool
	GitignoreFull     bool
	HasConfigExample  bool
	HasConfig         bool
	HasProductDoc     bool
	HasKiroFeature    bool
	HasExistingReport bool
	Strict            bool
}

// Generate satisfait testing/quick.Generator : produit un état de dépôt varié.
func (p18RepoState) Generate(r *rand.Rand, _ int) reflect.Value {
	bit := func() bool { return r.Intn(2) == 0 }
	return reflect.ValueOf(p18RepoState{
		HasGoMod:          bit(),
		HasCastor:         bit(),
		HasGitignore:      bit(),
		GitignoreFull:     bit(),
		HasConfigExample:  bit(),
		HasConfig:         bit(),
		HasProductDoc:     bit(),
		HasKiroFeature:    bit(),
		HasExistingReport: bit(),
		Strict:            bit(),
	})
}

func TestProperty18CheckOnlyWritesNothing(t *testing.T) {
	prop := func(state p18RepoState) bool {
		repo := t.TempDir()
		// Résoudre les symlinks (sur macOS t.TempDir() est sous /var → /private/var)
		// afin que la racine snapshotée corresponde au toplevel résolu par git,
		// c'est-à-dire l'endroit où l'onboarding écrirait s'il écrivait.
		root, err := filepath.EvalSymlinks(repo)
		if err != nil {
			t.Fatalf("eval symlinks: %v", err)
		}
		p18InitRepo(t, root)
		p18ApplyState(t, root, state)

		before := p18Snapshot(t, root)

		opts := onboarding.Options{
			CheckOnly:      true,
			NonInteractive: true,
			Plain:          true,
			Strict:         state.Strict,
			// Autofix volontairement à false : le mode check-only pur ne doit
			// jamais muter le dépôt.
		}
		var out bytes.Buffer
		if _, err := onboarding.Onboard(root, opts, nil, &out); err != nil {
			// Un check en lecture seule peut légitimement retourner une erreur
			// (ex. config absente) ; il doit néanmoins n'avoir rien écrit.
			t.Logf("onboard --check-only a retourné une erreur (toléré): %v", err)
		}

		after := p18Snapshot(t, root)
		if diff := p18Diff(before, after); diff != "" {
			t.Logf("check-only a muté le système de fichiers pour l'état %+v:\n%s", state, diff)
			return false
		}
		return true
	}

	// Générateur déterministe (seed fixe) + ≥ 100 itérations (conventions de test).
	cfg := &quick.Config{
		MaxCount: 100,
		Rand:     rand.New(rand.NewSource(20240518)),
	}
	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("Property 18 (`--check-only` n'écrit rien) violée: %v", err)
	}
}

// p18InitRepo initialise un dépôt git minimal dans dir.
func p18InitRepo(t *testing.T, dir string) {
	t.Helper()
	p18RunGit(t, dir, "init")
	p18RunGit(t, dir, "config", "user.email", "p18@example.com")
	p18RunGit(t, dir, "config", "user.name", "P18")
}

// p18ApplyState matérialise l'état initial décrit par s.
func p18ApplyState(t *testing.T, root string, s p18RepoState) {
	t.Helper()
	if s.HasGoMod {
		p18WriteFile(t, filepath.Join(root, "go.mod"), "module example.com/p18\n\ngo 1.25.0\n")
	}
	if s.HasCastor {
		p18WriteFile(t, filepath.Join(root, "castor.php"), "<?php\n")
	}
	if s.HasGitignore {
		content := "node_modules/\n"
		if s.GitignoreFull {
			content = ".asagiri/state.sqlite\nworktrees/\n"
		}
		p18WriteFile(t, filepath.Join(root, ".gitignore"), content)
	}
	if s.HasConfigExample {
		p18WriteFile(t, filepath.Join(root, ".asagiri", "config.yaml.example"), minimalExampleConfig())
	}
	if s.HasConfig {
		p18WriteFile(t, filepath.Join(root, ".asagiri", "config.yaml"), minimalExampleConfig())
	}
	if s.HasProductDoc {
		p18WriteFile(t, filepath.Join(root, "docs", "ai", "01-product.md"),
			"# Produit réel\n\nDescription détaillée du produit avec suffisamment de contenu pour ne pas être un placeholder, couvrant le périmètre, les utilisateurs et les invariants.\n")
	}
	if s.HasKiroFeature {
		p18WriteFile(t, filepath.Join(root, ".kiro", "specs", "demo", "requirements.md"), "# Demo feature\n")
	}
	if s.HasExistingReport {
		// Un rapport préexistant doit rester inchangé après check-only.
		p18WriteFile(t, filepath.Join(root, ".asagiri", "onboarding", "report.json"),
			"{\n  \"ready\": false,\n  \"score\": 0\n}\n")
	}
}

// p18Snapshot retourne une empreinte chemin→sha256 de tous les fichiers sous root.
func p18Snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	snap := map[string]string{}
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return relErr
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		sum := sha256.Sum256(data)
		snap[rel] = fmt.Sprintf("%x", sum)
		return nil
	})
	require.NoError(t, err)
	return snap
}

// p18Diff décrit les créations/modifications/suppressions entre deux snapshots.
// Retourne une chaîne vide si les deux snapshots sont identiques.
func p18Diff(before, after map[string]string) string {
	var b strings.Builder

	afterKeys := make([]string, 0, len(after))
	for k := range after {
		afterKeys = append(afterKeys, k)
	}
	sort.Strings(afterKeys)
	for _, k := range afterKeys {
		old, ok := before[k]
		if !ok {
			_, _ = fmt.Fprintf(&b, "  créé: %s\n", k)
			continue
		}
		if old != after[k] {
			_, _ = fmt.Fprintf(&b, "  modifié: %s\n", k)
		}
	}

	beforeKeys := make([]string, 0, len(before))
	for k := range before {
		beforeKeys = append(beforeKeys, k)
	}
	sort.Strings(beforeKeys)
	for _, k := range beforeKeys {
		if _, ok := after[k]; !ok {
			_, _ = fmt.Fprintf(&b, "  supprimé: %s\n", k)
		}
	}

	return b.String()
}

func p18RunGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	require.NoError(t, err, string(out))
}

func p18WriteFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}
