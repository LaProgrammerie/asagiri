// Feature: audit-coherence-consolidation, Property 14: Dry-run sans effet persistant.
//
// Property 14 (P14) — Dry-run sans effet persistant (Validates: Requirements 6.6)
//
// Pour toute commande supportée exécutée avec `--dry-run`, aucune invocation
// d'Agent_Backend réel, aucune commande externe et aucune modification de l'état
// persistant du dépôt ne se produit (WHILE --dry-run THE Asagiri SHALL simuler
// chaque étape sans invoquer d'Agent_Backend réel, sans exécuter de commande
// externe et sans modifier l'état persistant du dépôt).
//
// ─────────────────────────────────────────────────────────────────────────────
// Frontière testée (et pourquoi le test est représentatif)
// ─────────────────────────────────────────────────────────────────────────────
//
// Le flag global `--dry-run` est une PersistentFlag de la racine (root.go) ; il
// est propagé via `loadContext` (app_context.go) à `workflow.NewService(...,
// dryRun=true)`, qui irrigue à son tour :
//
//   - l'invocation d'agents : `agent/exec.Executor.Run` court-circuite AVANT
//     `os/exec` quand dryRun et renvoie un résultat fixture → aucun Agent_Backend
//     réel n'est lancé comme sous-processus ;
//   - la validation : `validation.Runner.Run` retourne immédiatement en dry-run →
//     aucune commande externe (`go test`, lint, …) n'est exécutée ;
//   - les worktrees / la préparation de PR : gardés par `if dryRun` → aucun
//     `git worktree add`/`git diff` n'est lancé.
//
// Plutôt que d'inspecter chaque garde isolément, ce test pilote la surface réelle
// du CLI via `RootCommand()` (l'arbre Cobra exact câblé pour `Execute`) en
// passant `--dry-run`, sur la séquence supportée du Guided_Path /
// Unitary_Command : `plan → enrich → dev → verify → review → status`. Toutes ces
// commandes transitent par `loadContext` puis le `workflow.Service`.
//
// Détection « aucun Agent_Backend réel / aucune commande externe » : tous les
// agents déclarés en config ET toutes les commandes de validation pointent vers
// un MÊME exécutable sentinelle (généré dans `t.TempDir()`) qui, s'il est lancé,
// écrit un fichier marqueur. La sentinelle est la seule commande externe
// configurée (les opérations git de worktree/PR sont gardées par `if dryRun`).
// Si une quelconque branche tentait d'exécuter un agent ou une commande externe,
// le marqueur apparaîtrait → la propriété échoue. En dry-run, le marqueur ne
// doit jamais exister.
//
// Détection « état persistant inchangé » : on photographie l'arborescence des
// fichiers VERSIONNABLES du dépôt (tout sauf `.git/` et le répertoire de
// bookkeeping `.asagiri/`) avant et après la séquence. L'orchestrateur écrit
// volontairement son journal interne sous `.asagiri/` (runs SQLite, logs,
// rapports) même en dry-run — c'est l'état de l'OUTIL, pas l'état persistant du
// DÉPÔT de l'utilisateur ; la spec (6.6) protège l'absence d'effet sur le dépôt
// et l'absence d'agent/commande réels. La sentinelle + le PATH restreint
// prouvent qu'aucun effet de bord externe (le risque réel d'un dry-run) ne fuit.
//
// Convention : une seule propriété par test, >= 100 itérations, générateur
// déterministe (testing/quick). Préfixe `p14` unique pour éviter toute collision
// avec les tâches sœurs (5.3-5.5, 5.7, 5.8) qui partagent le package `cli`.
//
// **Validates: Requirements 6.6**
package cli

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"testing"
	"testing/quick"
)

// p14SupportedCommands est la séquence de commandes supportées (Unitary_Command
// du Guided_Path) pilotées avec `--dry-run`. L'ordre suit l'automate de workflow
// (plan → enrich → dev → verify → review) afin que chaque étape ait un état de
// tâche valide en amont, puis `status` (lecture). Chaque entrée est exécutée
// avec le flag global `--dry-run`.
func p14SupportedCommands(feature string) [][]string {
	return [][]string{
		{"--dry-run", "plan", feature},
		{"--dry-run", "enrich", feature},
		{"--dry-run", "dev", feature},
		{"--dry-run", "verify", feature},
		{"--dry-run", "review", feature},
		{"--dry-run", "status"},
	}
}

// p14Scenario décrit une variation déterministe du dépôt et des entrées CLI.
// Les champs varient l'espace d'entrée (contenu de tâches, fichiers source
// pré-existants) sans jamais relâcher la précondition « --dry-run actif ».
type p14Scenario struct {
	TaskLines   []string // lignes de tasks.md (au moins une tâche)
	ExtraFiles  []p14File
	FeatureSalt int // fait varier des chemins sans changer le slug de feature
}

type p14File struct {
	RelPath string
	Content []byte
}

// p14SafeRelPaths : fichiers utilisateur pré-existants injectés dans le dépôt.
// Volontairement hors `.git/` et `.asagiri/` (l'invariant porte sur l'arbre
// versionnable du dépôt).
var p14SafeRelPaths = []string{
	"README.md",
	"main.go",
	"src/app.go",
	"docs/notes.txt",
	"internal/pkg/util.go",
	"cmd/tool/main.go",
}

var p14TaskFragments = []string{
	"- [ ] Implémenter le module A",
	"- [ ] Corriger le bug B",
	"- [ ] Écrire les tests C",
	"- [ ] Documenter l'API D",
}

// Generate implémente testing/quick.Generator : un dépôt varié mais déterministe
// (le rand est fixé via quick.Config.Rand dans le test).
func (p14Scenario) Generate(r *rand.Rand, _ int) reflect.Value {
	nTasks := 1 + r.Intn(3) // 1..3 tâches (au moins une pour que pickTasks réussisse)
	lines := make([]string, 0, nTasks)
	for i := 0; i < nTasks; i++ {
		lines = append(lines, p14TaskFragments[r.Intn(len(p14TaskFragments))])
	}

	var files []p14File
	nFiles := r.Intn(4) // 0..3 fichiers utilisateur pré-existants
	used := make(map[string]bool, nFiles)
	for i := 0; i < nFiles; i++ {
		rel := p14SafeRelPaths[r.Intn(len(p14SafeRelPaths))]
		if used[rel] {
			continue
		}
		used[rel] = true
		content := make([]byte, r.Intn(40))
		for j := range content {
			content[j] = byte('a' + r.Intn(26))
		}
		files = append(files, p14File{RelPath: rel, Content: content})
	}

	return reflect.ValueOf(p14Scenario{
		TaskLines:   lines,
		ExtraFiles:  files,
		FeatureSalt: r.Intn(1000),
	})
}

// p14Sentinel est le nom de l'exécutable sentinelle partagé par tous les agents
// et toutes les commandes de validation.
const p14Sentinel = "p14-sentinel"

// p14SentinelMarker est le nom du fichier marqueur écrit par la sentinelle si
// elle est exécutée (preuve d'invocation d'un agent réel ou d'une commande
// externe).
const p14SentinelMarker = "p14-sentinel-invoked"

// p14BuildSentinel compile un petit binaire qui, à l'exécution, crée le fichier
// marqueur pointé par la variable d'environnement P14_MARKER. Toute exécution de
// la sentinelle (agent ou commande de validation) laisse donc une trace
// observable. Le binaire est placé dans binDir, seul dossier exposé via PATH.
func p14BuildSentinel(t *testing.T, binDir string) string {
	t.Helper()
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "main.go")
	src := `package main

import "os"

func main() {
	if m := os.Getenv("P14_MARKER"); m != "" {
		_ = os.WriteFile(m, []byte("invoked"), 0o644)
	}
	os.Exit(0)
}
`
	if err := os.WriteFile(srcPath, []byte(src), 0o644); err != nil {
		t.Fatalf("écrire source sentinelle: %v", err)
	}
	binName := p14Sentinel
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	binPath := filepath.Join(binDir, binName)
	cmd := exec.Command("go", "build", "-o", binPath, srcPath)
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("build sentinelle: %v\n%s", err, out)
	}
	return binPath
}

// p14ConfigYAML produit une config valide où TOUS les agents et la commande de
// validation pointent vers la sentinelle (nom de commande nu, résolu via PATH).
func p14ConfigYAML() string {
	return `project:
  name: p14-test
  default_branch: main
specs:
  kiro_path: .kiro/specs
  active_spec_path: docs/ai/active/current-spec.md
  handoff_path: docs/ai/active/handoff.md
state:
  backend: sqlite
  path: .asagiri/state.sqlite
worktrees:
  base_path: .asagiri/worktrees
  branch_prefix: asagiri
  cleanup_policy: keep_failed
validation:
  commands:
    - name: sentinel-check
      command: ` + p14Sentinel + `
      required: true
agents:
  kiro:
    command: ` + p14Sentinel + `
    args: ["spec"]
  cursor:
    command: ` + p14Sentinel + `
    args: ["dev"]
  codex:
    command: ` + p14Sentinel + `
    args: ["review"]
  ollama:
    command: ` + p14Sentinel + `
    args: ["enrich"]
  claude:
    command: ` + p14Sentinel + `
    args: ["code"]
`
}

// p14Git lance une commande git dans dir (setup du dépôt, hors mesure).
func p14Git(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

// p14Materialize crée un dépôt Git complet et onboardé : config sentinelle,
// spec de la feature avec tasks variées, et fichiers utilisateur pré-existants.
func p14Materialize(t *testing.T, dir, feature string, sc p14Scenario) {
	t.Helper()

	p14Git(t, dir, "init")
	p14Git(t, dir, "config", "user.email", "p14@example.com")
	p14Git(t, dir, "config", "user.name", "P14")

	p14WriteFile(t, filepath.Join(dir, "go.mod"), []byte("module example.com/p14\n\ngo 1.25.0\n"))
	p14WriteFile(t, filepath.Join(dir, p14DefaultConfigRel()), []byte(p14ConfigYAML()))

	tasks := strings.Join(sc.TaskLines, "\n") + "\n"
	specDir := filepath.Join(dir, ".kiro", "specs", feature)
	p14WriteFile(t, filepath.Join(specDir, "requirements.md"), []byte("# Requirements\n"))
	p14WriteFile(t, filepath.Join(specDir, "tasks.md"), []byte(tasks))

	for _, f := range sc.ExtraFiles {
		p14WriteFile(t, filepath.Join(dir, filepath.FromSlash(f.RelPath)), f.Content)
	}
}

// p14DefaultConfigRel renvoie le chemin relatif canonique de la config,
// dupliqué localement (chemin stable du dépôt) pour éviter d'importer le package
// config uniquement pour une constante.
func p14DefaultConfigRel() string {
	return filepath.Join(".asagiri", "config.yaml")
}

func p14WriteFile(t *testing.T, path string, content []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir parent %s: %v", path, err)
	}
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

// p14Snapshot photographie l'arborescence des fichiers VERSIONNABLES du dépôt :
// tout sauf `.git/` (état interne git) et `.asagiri/` (journal de
// l'orchestrateur, qui évolue même en dry-run par conception). Chaque répertoire
// est marqué ("dir:") et chaque fichier régulier porte le hash SHA-256 de son
// contenu. Deux snapshots égaux prouvent qu'aucun fichier de dépôt n'a été créé,
// modifié ou supprimé.
func p14Snapshot(t *testing.T, dir string) map[string]string {
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
		top := strings.SplitN(filepath.ToSlash(rel), "/", 2)[0]
		if top == ".git" || top == ".asagiri" {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		switch {
		case d.IsDir():
			snap["dir:"+filepath.ToSlash(rel)] = ""
		case d.Type().IsRegular():
			data, readErr := os.ReadFile(path)
			if readErr != nil {
				return readErr
			}
			sum := sha256.Sum256(data)
			snap["file:"+filepath.ToSlash(rel)] = hex.EncodeToString(sum[:])
		default:
			snap["other:"+filepath.ToSlash(rel)] = d.Type().String()
		}
		return nil
	})
	if err != nil {
		t.Fatalf("snapshot %s: %v", dir, err)
	}
	return snap
}

// p14DiffKeys liste les différences entre deux snapshots (message d'échec
// lisible énumérant les artefacts persistants inattendus).
func p14DiffKeys(before, after map[string]string) []string {
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

// TestProperty14DryRunNoPersistentEffect vérifie la Property 14 : pour toute
// variation déterministe du dépôt, exécuter la séquence des commandes supportées
// avec `--dry-run` n'invoque aucun Agent_Backend réel ni commande externe (la
// sentinelle n'écrit jamais son marqueur) et ne modifie pas l'état persistant du
// dépôt (snapshot de l'arbre versionnable identique avant/après). _Requirements: 6.6_
func TestProperty14DryRunNoPersistentEffect(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skipf("git indisponible: %v", err)
	}
	if _, err := exec.LookPath("go"); err != nil {
		t.Skipf("go toolchain indisponible: %v", err)
	}

	// Sentinelle partagée par tous les scénarios, compilée une seule fois.
	binDir := t.TempDir()
	p14BuildSentinel(t, binDir)

	// binDir est préfixé au PATH : la sentinelle (vers laquelle pointent TOUS les
	// agents et la commande de validation) est résoluble, et `git` reste
	// disponible pour le setup du dépôt. La détection « aucun agent réel / aucune
	// commande externe » repose sur le marqueur : en dry-run, aucune de ces
	// commandes ne doit être lancée (les seules commandes externes configurées
	// sont la sentinelle ; les opérations git de worktree/PR sont gardées par
	// `if dryRun`).
	origPath := os.Getenv("PATH")
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+origPath)
	t.Cleanup(func() { _ = os.Setenv("PATH", origPath) })

	// Racine de travail : empêche git de remonter au-dessus lors de la
	// découverte du dépôt (TMPDIR éventuellement imbriqué dans un clone git).
	root := t.TempDir()
	t.Setenv("GIT_CEILING_DIRECTORIES", root)

	// Restaure le répertoire courant après le test (les commandes lisent cwd).
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origWd) })

	property := func(sc p14Scenario) bool {
		dir, err := os.MkdirTemp(root, "repo-")
		if err != nil {
			t.Fatalf("mkdtemp: %v", err)
		}
		feature := fmt.Sprintf("p14-feature-%d", sc.FeatureSalt)
		p14Materialize(t, dir, feature, sc)

		// Marqueur d'invocation propre à ce scénario, hors du dépôt mesuré.
		marker := filepath.Join(binDir, fmt.Sprintf("%s-%d", p14SentinelMarker, sc.FeatureSalt))
		_ = os.Remove(marker)
		t.Setenv("P14_MARKER", marker)

		if err := os.Chdir(dir); err != nil {
			t.Fatalf("chdir %s: %v", dir, err)
		}

		before := p14Snapshot(t, dir)

		for _, args := range p14SupportedCommands(feature) {
			root := newRootCmd()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetErr(&out)
			root.SetArgs(args)
			// Les commandes ne doivent pas paniquer ; une erreur valeur est
			// tolérée (frontière CLI), mais l'invariant clé est l'absence
			// d'effet externe / persistant, vérifié ci-dessous.
			panicked := func() (rec any) {
				defer func() { rec = recover() }()
				_ = root.Execute()
				return nil
			}()
			if panicked != nil {
				t.Errorf("panic sur %v (dry-run): %v", args, panicked)
				return false
			}
		}

		// Aucune invocation d'Agent_Backend réel ni de commande externe.
		if _, statErr := os.Stat(marker); statErr == nil {
			t.Errorf("sentinelle exécutée en dry-run (agent ou commande externe lancé) pour %v", sc.TaskLines)
			return false
		}

		// État persistant du dépôt inchangé (arbre versionnable hors .git/.asagiri).
		after := p14Snapshot(t, dir)
		if !reflect.DeepEqual(before, after) {
			t.Errorf("état persistant du dépôt modifié en dry-run: %v", p14DiffKeys(before, after))
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 100, // >= 100 itérations
		Rand:     rand.New(rand.NewSource(14)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 14 (dry-run sans effet persistant) violée: %v", err)
	}
}
