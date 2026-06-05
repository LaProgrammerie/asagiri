package cli

// Feature: audit-coherence-consolidation, Property 13
//
// Property 13 (P13) — Dépôt inchangé sur abandon (Validates: Requirements 6.4)
//
// Pour toute commande qui s'arrête à cause d'un prérequis manquant, l'état du
// dépôt après l'arrêt est identique à l'état avant l'appel : aucun fichier créé,
// modifié ou supprimé, aucun répertoire créé, aucun artefact partiel.
//
// Frontière testée : loadContext (application/internal/cli/app_context.go), par
// laquelle transitent toutes les Unitary_Command et le Guided_Path (next,
// continue, work, …). loadContext lit le dépôt (bootstrap.GitRoot puis
// config.Load) AVANT toute écriture (l'ouverture/migration SQLite n'a lieu
// qu'après un chargement de config réussi). Un prérequis manquant (pas de dépôt
// Git, ou dépôt Git sans config) provoque donc un abandon sans aucune écriture.
//
// Les helpers de ce fichier sont préfixés `p13` pour éviter toute collision avec
// les tâches sœurs (5.3, 5.4, 5.6-5.8) qui partagent le package `cli`.

import (
	"crypto/sha256"
	"encoding/hex"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"testing/quick"
)

// p13File est un fichier pré-existant déterministe injecté dans l'état initial.
type p13File struct {
	RelPath string
	Content []byte
}

// p13InitialState décrit un état initial varié du dépôt avant un abandon.
//
//   - WithGit == false : aucun dépôt Git → bootstrap.GitRoot échoue.
//   - WithGit == true  : dépôt Git mais aucune config → config.Load échoue.
//
// Dans les deux cas, le prérequis est manquant et loadContext doit abandonner
// sans rien écrire.
type p13InitialState struct {
	WithGit bool
	DryRun  bool
	Files   []p13File
}

// p13SafeRelPaths est le vivier de chemins relatifs « sûrs » pour la génération.
// Il exclut volontairement `.git/...` et `.asagiri/config.yaml` afin de
// garantir que le prérequis reste manquant (l'abandon est toujours déclenché).
var p13SafeRelPaths = []string{
	"README.md",
	"main.go",
	"src/app.go",
	"docs/notes.txt",
	"data.json",
	"a/b/c.txt",
	".asagiri/state-old.txt",
	"pkg/util/helper.go",
}

// Generate implémente quick.Generator pour produire des états initiaux variés
// mais déterministes (le rand est fixé via quick.Config.Rand dans le test).
func (p13InitialState) Generate(r *rand.Rand, _ int) reflect.Value {
	state := p13InitialState{
		WithGit: r.Intn(2) == 1,
		DryRun:  r.Intn(2) == 1,
	}
	n := r.Intn(5) // 0..4 fichiers pré-existants
	used := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		rel := p13SafeRelPaths[r.Intn(len(p13SafeRelPaths))]
		if used[rel] {
			continue
		}
		used[rel] = true
		content := make([]byte, r.Intn(48)) // <= 47 octets
		for j := range content {
			content[j] = byte('a' + r.Intn(26))
		}
		state.Files = append(state.Files, p13File{RelPath: rel, Content: content})
	}
	return reflect.ValueOf(state)
}

// p13Materialize crée l'état initial dans dir : fichiers pré-existants puis,
// si demandé, un dépôt Git réel (avant le snapshot, donc inclus dedans).
func p13Materialize(t *testing.T, dir string, state p13InitialState) {
	t.Helper()
	for _, f := range state.Files {
		full := filepath.Join(dir, filepath.FromSlash(f.RelPath))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir parent %s: %v", f.RelPath, err)
		}
		if err := os.WriteFile(full, f.Content, 0o644); err != nil {
			t.Fatalf("write %s: %v", f.RelPath, err)
		}
	}
	if state.WithGit {
		p13Git(t, dir, "init")
		p13Git(t, dir, "config", "user.email", "p13@example.com")
		p13Git(t, dir, "config", "user.name", "P13")
	}
}

func p13Git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// p13Snapshot capture l'état complet de l'arborescence sous dir : chaque
// répertoire (marqueur "dir:") et chaque fichier régulier (hash SHA-256 de son
// contenu). Deux snapshots égaux prouvent qu'aucun fichier/répertoire n'a été
// créé, modifié ou supprimé.
func p13Snapshot(t *testing.T, dir string) map[string]string {
	t.Helper()
	snap := make(map[string]string)
	err := filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}
		switch {
		case d.IsDir():
			snap["dir:"+filepath.ToSlash(rel)] = ""
		case d.Type().IsRegular():
			data, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			sum := sha256.Sum256(data)
			snap["file:"+filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		default:
			// liens symboliques / spéciaux : on enregistre leur présence.
			snap["other:"+filepath.ToSlash(rel)] = d.Type().String()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	return snap
}

// p13DiffKeys retourne les clés divergentes entre deux snapshots (pour un
// message d'échec lisible listant les artefacts inattendus).
func p13DiffKeys(before, after map[string]string) []string {
	var diff []string
	for k, v := range after {
		if bv, ok := before[k]; !ok {
			diff = append(diff, "créé: "+k)
		} else if bv != v {
			diff = append(diff, "modifié: "+k)
		}
	}
	for k := range before {
		if _, ok := after[k]; !ok {
			diff = append(diff, "supprimé: "+k)
		}
	}
	sort.Strings(diff)
	return diff
}

// TestProp13RepoUnchangedOnAbort vérifie la Property 13 : un abandon sur
// prérequis manquant laisse le dépôt strictement inchangé.
func TestProp13RepoUnchangedOnAbort(t *testing.T) {
	root := t.TempDir()
	// Empêche `git` de remonter au-dessus de `root` lors de la découverte du
	// dépôt : garantit que le scénario « pas de dépôt Git » échoue réellement
	// même si le TMPDIR est imbriqué dans un clone Git.
	t.Setenv("GIT_CEILING_DIRECTORIES", root)

	property := func(state p13InitialState) bool {
		dir, err := os.MkdirTemp(root, "repo-")
		if err != nil {
			t.Fatalf("mkdtemp: %v", err)
		}
		p13Materialize(t, dir, state)

		before := p13Snapshot(t, dir)

		ctx, loadErr := loadContext(dir, state.DryRun)
		if loadErr == nil {
			// Le prérequis manquant aurait dû provoquer un abandon : si
			// loadContext réussit, le générateur n'a pas produit un cas
			// d'abandon (invariant de test violé).
			ctx.Close()
			t.Errorf("loadContext a réussi alors qu'un prérequis manque (WithGit=%v)", state.WithGit)
			return false
		}

		after := p13Snapshot(t, dir)
		if !reflect.DeepEqual(before, after) {
			t.Errorf("dépôt modifié après abandon (WithGit=%v): %v",
				state.WithGit, p13DiffKeys(before, after))
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200, // >= 100 itérations
		Rand:     rand.New(rand.NewSource(13)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 13 (dépôt inchangé sur abandon) violée: %v", err)
	}
}
