package cli

// Feature: audit-coherence-consolidation, Property 12
//
// Property 12 (P12) — Remédiation guidée sans panic
// (Validates: Requirements 6.3, 7.7)
//
// Pour toute commande exécutée alors qu'un prérequis requis (outil, dépôt,
// configuration ou dépendance) est absent, le système doit :
//   - s'arrêter SANS `panic` (erreur retournée comme valeur, 03-standards.md) ;
//   - retourner une erreur non nil → code de sortie non nul à la frontière CLI
//     (Cobra propage l'erreur de `RunE` en exit ≠ 0) ;
//   - émettre une Guided_Remediation NOMMANT l'élément manquant ET au moins une
//     action de résolution.
//
// Frontière testée : `loadContext` (application/internal/cli/app_context.go),
// par laquelle transitent toutes les Unitary_Command et le Guided_Path
// (next, continue, work, …). `loadContext` enchaîne deux prérequis :
//
//   1. `bootstrap.GitRoot` — exige l'outil/dépôt Git. En son absence (pas de
//      dépôt Git, ou binaire `git` introuvable), il retourne un message nommant
//      le dépôt Git et l'action de résolution (`git init` / cloner).
//   2. `config.Load` — exige la configuration `.asagiri/config.yaml`. En son
//      absence, `loadContext` enrichit l'erreur d'E/S brute en une
//      Guided_Remediation nommant le fichier de config et l'action `asa init`.
//
// Le générateur produit des cas variés couvrant ces deux familles de prérequis
// manquants (outil/dépôt vs configuration), avec un état de dépôt pré-existant
// varié. La propriété asserte l'absence de panic, la présence d'une erreur, et
// que le message nomme l'élément manquant ET une action de résolution.
//
// Convention : une seule propriété par test, ≥ 100 itérations, générateur
// déterministe (testing/quick). Préfixe `p12` unique pour éviter toute collision
// avec les autres tests du package cli (5.3, 5.5–5.8 partagent ce package).

import (
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// p12Kind sélectionne quel prérequis est manquant dans le cas généré.
type p12Kind int

const (
	// p12NoGitRepo : aucun dépôt Git → bootstrap.GitRoot échoue (prérequis
	// outil/dépôt manquant).
	p12NoGitRepo p12Kind = iota
	// p12NoConfig : dépôt Git présent mais aucune configuration Asagiri →
	// config.Load échoue (prérequis configuration manquant).
	p12NoConfig
)

// p12File est un fichier pré-existant déterministe injecté dans l'état initial.
type p12File struct {
	RelPath string
	Content []byte
}

// p12Case décrit un scénario de prérequis manquant varié.
type p12Case struct {
	Kind   p12Kind
	DryRun bool
	Files  []p12File
}

// p12SafeRelPaths : chemins relatifs « sûrs » pour la génération. Exclut
// volontairement `.git/...`, `.asagiri/config.yaml` et le legacy
// `.agentflow/config.yaml` afin que le prérequis ciblé reste réellement
// manquant (l'abandon est toujours déclenché).
var p12SafeRelPaths = []string{
	"README.md",
	"main.go",
	"src/app.go",
	"docs/notes.txt",
	"data.json",
	"a/b/c.txt",
	".asagiri/state-old.txt",
	"pkg/util/helper.go",
}

// Generate implémente testing/quick.Generator : produit un cas varié mais
// déterministe (le rand est fixé via quick.Config.Rand dans le test).
func (p12Case) Generate(r *rand.Rand, _ int) reflect.Value {
	c := p12Case{
		Kind:   p12Kind(r.Intn(2)),
		DryRun: r.Intn(2) == 1,
	}
	n := r.Intn(5) // 0..4 fichiers pré-existants
	used := make(map[string]bool, n)
	for i := 0; i < n; i++ {
		rel := p12SafeRelPaths[r.Intn(len(p12SafeRelPaths))]
		if used[rel] {
			continue
		}
		used[rel] = true
		content := make([]byte, r.Intn(32))
		for j := range content {
			content[j] = byte('a' + r.Intn(26))
		}
		c.Files = append(c.Files, p12File{RelPath: rel, Content: content})
	}
	return reflect.ValueOf(c)
}

// p12Materialize crée l'état initial dans dir selon le cas : fichiers
// pré-existants puis, pour p12NoConfig uniquement, un dépôt Git réel (pour que
// le seul prérequis manquant soit la configuration).
func p12Materialize(t *testing.T, dir string, c p12Case) {
	t.Helper()
	for _, f := range c.Files {
		full := filepath.Join(dir, filepath.FromSlash(f.RelPath))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("mkdir parent %s: %v", f.RelPath, err)
		}
		if err := os.WriteFile(full, f.Content, 0o644); err != nil {
			t.Fatalf("write %s: %v", f.RelPath, err)
		}
	}
	if c.Kind == p12NoConfig {
		p12Git(t, dir, "init")
		p12Git(t, dir, "config", "user.email", "p12@example.com")
		p12Git(t, dir, "config", "user.name", "P12")
	}
}

func p12Git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// p12ElementTokens retourne les jetons (insensibles à la casse) dont au moins un
// doit apparaître dans le message pour prouver qu'il NOMME l'élément manquant.
func p12ElementTokens(kind p12Kind) []string {
	switch kind {
	case p12NoGitRepo:
		return []string{"git", "dépôt"}
	default: // p12NoConfig
		return []string{".asagiri/config.yaml", "config"}
	}
}

// p12ActionTokens : jetons d'action de résolution (insensibles à la casse).
// Au moins un doit apparaître dans le message pour prouver qu'une ACTION de
// résolution est proposée.
var p12ActionTokens = []string{
	"git init",
	"asa init",
	"clone",
	"copiez",
	"exécutez",
	"lancez",
}

func p12ContainsAny(haystackLower string, tokens []string) bool {
	for _, tok := range tokens {
		if strings.Contains(haystackLower, strings.ToLower(tok)) {
			return true
		}
	}
	return false
}

// p12RunLoadContext exécute loadContext en capturant tout panic éventuel, afin
// de prouver concrètement l'absence de `panic` à la frontière CLI.
func p12RunLoadContext(dir string, dryRun bool) (ctx *appContext, panicked any, err error) {
	defer func() { panicked = recover() }()
	ctx, err = loadContext(dir, dryRun)
	return ctx, nil, err
}

// TestProperty12GuidedRemediationNoPanic vérifie la Property 12 : un prérequis
// manquant provoque un arrêt sans panic, une erreur (exit ≠ 0 à la frontière) et
// un message nommant l'élément manquant ET au moins une action de résolution.
// _Requirements: 6.3, 7.7_
func TestProperty12GuidedRemediationNoPanic(t *testing.T) {
	root := t.TempDir()
	// Empêche `git` de remonter au-dessus de `root` : garantit que le scénario
	// « pas de dépôt Git » échoue réellement même si TMPDIR est imbriqué dans un
	// clone Git.
	t.Setenv("GIT_CEILING_DIRECTORIES", root)

	property := func(c p12Case) bool {
		dir, err := os.MkdirTemp(root, "repo-")
		if err != nil {
			t.Fatalf("mkdtemp: %v", err)
		}
		p12Materialize(t, dir, c)

		ctx, panicked, loadErr := p12RunLoadContext(dir, c.DryRun)

		// (a) Aucune panic à la frontière CLI.
		if panicked != nil {
			t.Errorf("loadContext a paniqué (kind=%d): %v", c.Kind, panicked)
			return false
		}
		// (b) Le prérequis manquant DOIT provoquer un abandon (erreur non nil).
		//     Une réussite signifie que le cas généré n'a pas réellement un
		//     prérequis manquant (invariant de test violé).
		if loadErr == nil {
			if ctx != nil {
				ctx.Close()
			}
			t.Errorf("loadContext a réussi alors qu'un prérequis manque (kind=%d)", c.Kind)
			return false
		}

		msgLower := strings.ToLower(loadErr.Error())

		// (c) Le message NOMME l'élément manquant.
		if !p12ContainsAny(msgLower, p12ElementTokens(c.Kind)) {
			t.Errorf("le message ne nomme pas l'élément manquant (kind=%d): %q", c.Kind, loadErr.Error())
			return false
		}
		// (d) Le message propose au moins une ACTION de résolution.
		if !p12ContainsAny(msgLower, p12ActionTokens) {
			t.Errorf("le message ne propose aucune action de résolution (kind=%d): %q", c.Kind, loadErr.Error())
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200, // ≥ 100 itérations
		Rand:     rand.New(rand.NewSource(12)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 12 (remédiation guidée sans panic) violée: %v", err)
	}
}

// TestGuidedRemediationExamples documente la frontière sur deux cas concrets et
// contrastés, en complément du test de propriété : (1) dépôt Git absent →
// message nommant le dépôt + action ; (2) dépôt Git présent mais configuration
// absente → message nommant la config + action `asa init`. _Requirements: 6.3, 7.7_
func TestGuidedRemediationExamples(t *testing.T) {
	root := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", root)

	// (1) Aucun dépôt Git : prérequis outil/dépôt manquant.
	noGit, err := os.MkdirTemp(root, "no-git-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	if _, err := loadContext(noGit, false); err == nil {
		t.Fatal("attendu une erreur en l'absence de dépôt Git")
	} else {
		m := strings.ToLower(err.Error())
		if !p12ContainsAny(m, p12ElementTokens(p12NoGitRepo)) {
			t.Fatalf("message %q ne nomme pas le dépôt Git", err.Error())
		}
		if !p12ContainsAny(m, p12ActionTokens) {
			t.Fatalf("message %q ne propose pas d'action de résolution", err.Error())
		}
	}

	// (2) Dépôt Git présent mais configuration absente : prérequis config manquant.
	withGit, err := os.MkdirTemp(root, "no-config-")
	if err != nil {
		t.Fatalf("mkdtemp: %v", err)
	}
	p12Git(t, withGit, "init")
	p12Git(t, withGit, "config", "user.email", "p12@example.com")
	p12Git(t, withGit, "config", "user.name", "P12")
	if _, err := loadContext(withGit, false); err == nil {
		t.Fatal("attendu une erreur en l'absence de configuration")
	} else {
		m := strings.ToLower(err.Error())
		if !p12ContainsAny(m, p12ElementTokens(p12NoConfig)) {
			t.Fatalf("message %q ne nomme pas la configuration manquante", err.Error())
		}
		if !strings.Contains(m, "asa init") {
			t.Fatalf("message %q ne propose pas l'action `asa init`", err.Error())
		}
	}
}
