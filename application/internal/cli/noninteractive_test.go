// Feature: audit-coherence-consolidation, Property 15: Non-interactif sans `--yes`.
//
// Pour toute commande requérant une confirmation exécutée en mode non interactif
// (CI ou `--non-interactive`) sans `--yes`, le système s'arrête sans attendre de
// saisie, retourne une erreur (mappée en code de sortie non nul à la frontière
// CLI) et nomme le flag à fournir (`--yes`).
//
// Frontière testée : `requireConfirm` (application/internal/cli/intent_helpers.go),
// le gate de confirmation réutilisé par le Guided_Path (work/continue), la
// synchronisation et les jalons de validation humaine. Lorsque
// `opts.Interactive == false` (non-TTY / `--non-interactive`) et
// `opts.Yes == false`, la fonction renvoie un `*intent.ConfirmationRequiredError`
// AVANT tout appel à `confirmPrompt`, donc sans lire `os.Stdin`. Le test impose
// une entrée standard bloquante : si une régression tentait de lire la saisie,
// le scanner se bloquerait et le délai d'attente ferait échouer la propriété
// (preuve concrète de « s'arrête sans attendre de saisie »).
//
// Convention : une seule propriété par test, >= 100 itérations, générateur
// déterministe (testing/quick). Préfixe `p15` unique pour éviter toute collision
// avec les autres tests du package cli.
//
// **Validates: Requirements 6.7**
package cli

import (
	"errors"
	"math/rand"
	"os"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/intent"
)

// p15Case est un scénario de confirmation non interactif généré. Par
// construction il encode toujours la précondition de la Property 15 :
// Interactive=false (non-TTY ou --non-interactive) ET Yes=false (--yes absent).
// Seuls des champs variés et par ailleurs sans incidence sur la branche testée
// sont portés, afin que quick.Check imprime des contre-exemples lisibles ; les
// WorkOptions sont reconstruites de façon déterministe à partir d'eux.
type p15Case struct {
	msg      string
	dryRun   bool
	planOnly bool
	maxTasks int
}

// Generate implémente testing/quick.Generator : il produit un scénario varié
// mais satisfaisant toujours la précondition (non interactif, sans --yes).
func (p15Case) Generate(rnd *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(p15Case{
		msg:      p15RandomMessage(rnd),
		dryRun:   rnd.Intn(2) == 0,
		planOnly: rnd.Intn(2) == 0,
		maxTasks: rnd.Intn(8),
	})
}

// options reconstruit les WorkOptions de la précondition : non interactif et
// sans --yes. Les autres champs varient sans influencer la branche testée.
func (c p15Case) options() intent.WorkOptions {
	return intent.WorkOptions{
		Interactive: false, // non-TTY / --non-interactive
		Yes:         false, // --yes absent
		DryRun:      c.dryRun,
		PlanOnly:    c.planOnly,
		MaxTasks:    c.maxTasks,
	}
}

// p15MessageFragments couvre des messages de confirmation représentatifs des
// jalons (plan, budget, action sensible) ainsi que des cas limites (vide).
var p15MessageFragments = []string{
	"Proceed with execution plan?",
	"Écraser la spec locale modifiée?",
	"Dépassement de budget, continuer?",
	"Confirmer l'action sensible (suppression)?",
	"déployer en production ?",
	"",
}

func p15RandomMessage(rnd *rand.Rand) string {
	return p15MessageFragments[rnd.Intn(len(p15MessageFragments))]
}

// p15RunRequireConfirm exécute requireConfirm dans une goroutine et signale si
// l'appel est revenu sans lire l'entrée standard. Si requireConfirm tentait de
// lire `os.Stdin` (redirigé vers un tube jamais alimenté par l'appelant), il se
// bloquerait ; le délai d'attente surfacerait alors un échec plutôt que de
// suspendre la suite. Tout panic éventuel est capturé pour prouver l'absence de
// `panic` aux frontières CLI.
func p15RunRequireConfirm(opts intent.WorkOptions, msg string) (timedOut bool, panicked any, err error) {
	type result struct {
		err error
		rec any
	}
	done := make(chan result, 1)
	go func() {
		var rec any
		var e error
		func() {
			defer func() { rec = recover() }()
			e = requireConfirm(opts, msg)
		}()
		done <- result{err: e, rec: rec}
	}()
	select {
	case r := <-done:
		return false, r.rec, r.err
	case <-time.After(2 * time.Second):
		return true, nil, nil
	}
}

// TestProperty15NonInteractiveWithoutYes vérifie la Property 15 : en mode non
// interactif sans --yes, une confirmation requise s'arrête sans attendre de
// saisie, retourne une erreur (exit non nul à la frontière CLI) et nomme le flag
// `--yes`. _Requirements: 6.7_
func TestProperty15NonInteractiveWithoutYes(t *testing.T) {
	// Entrée standard bloquante : un tube dont l'extrémité d'écriture n'est
	// jamais alimentée. requireConfirm NE DOIT PAS la lire dans la branche non
	// interactif / sans --yes ; s'il le faisait, le scanner de confirmPrompt se
	// bloquerait et le délai d'attente par itération ferait échouer la propriété.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}
	origStdin := os.Stdin
	os.Stdin = r
	t.Cleanup(func() {
		os.Stdin = origStdin
		_ = w.Close() // débloque une éventuelle goroutine lectrice fuitée
		_ = r.Close()
	})

	property := func(c p15Case) bool {
		opts := c.options()
		timedOut, panicked, errResult := p15RunRequireConfirm(opts, c.msg)

		if panicked != nil {
			t.Errorf("requireConfirm a paniqué (msg=%q): %v", c.msg, panicked)
			return false
		}
		// s'arrête sans attendre de saisie.
		if timedOut {
			t.Errorf("requireConfirm a attendu une saisie au lieu de s'arrêter (msg=%q)", c.msg)
			return false
		}
		// s'arrête avec une erreur -> code de sortie non nul à la frontière CLI.
		if errResult == nil {
			t.Errorf("aucune erreur renvoyée en mode non interactif sans --yes (msg=%q)", c.msg)
			return false
		}
		// l'erreur est bien le gate de confirmation, pas une erreur de lecture stdin.
		var confErr *intent.ConfirmationRequiredError
		if !errors.As(errResult, &confErr) {
			t.Errorf("type d'erreur %T, attendu *intent.ConfirmationRequiredError (msg=%q)", errResult, c.msg)
			return false
		}
		// nomme le flag à fournir.
		if !strings.Contains(errResult.Error(), "--yes") {
			t.Errorf("le message ne nomme pas le flag --yes: %q", errResult.Error())
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200, // >= 100 iterations
		Rand:     rand.New(rand.NewSource(15)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 15 (non-interactif sans --yes) violée: %v", err)
	}
}

// TestRequireConfirmNonInteractiveExamples documente la frontière sur deux cas
// concrets et contrastés, en complément du test de propriété : (1) non interactif
// sans --yes -> erreur nommant --yes ; (2) --yes présent -> aucune confirmation
// requise (nil), sans saisie. _Requirements: 6.7_
func TestRequireConfirmNonInteractiveExamples(t *testing.T) {
	// (1) non interactif, sans --yes : arrêt + message nommant --yes.
	err := requireConfirm(intent.WorkOptions{Interactive: false, Yes: false}, "Proceed?")
	if err == nil {
		t.Fatal("attendu une erreur en mode non interactif sans --yes")
	}
	var confErr *intent.ConfirmationRequiredError
	if !errors.As(err, &confErr) {
		t.Fatalf("type d'erreur %T, attendu *intent.ConfirmationRequiredError", err)
	}
	if !strings.Contains(err.Error(), "--yes") {
		t.Fatalf("message %q ne nomme pas le flag --yes", err.Error())
	}

	// (2) --yes présent : la confirmation est accordée sans saisie ni erreur.
	if err := requireConfirm(intent.WorkOptions{Interactive: false, Yes: true}, "Proceed?"); err != nil {
		t.Fatalf("avec --yes, aucune confirmation requise attendue, got %v", err)
	}
}
