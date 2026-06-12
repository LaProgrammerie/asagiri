package policy

// Feature: audit-coherence-consolidation, Property 20
//
// Property 20: Automate de statut du Remediation_Register.
// For any séquence de transitions appliquée à une entrée du registre, seules les
// transitions `ouvert → en cours`, `en cours → clôturé` et `clôturé → ouvert`
// sont acceptées, et le statut reste toujours dans `{ouvert, en cours, clôturé}`.
//
// Validates: Requirements 3.2, 3.4
//
// La fonction de transition est PURE et définie dans ce test : aucune logique de
// production n'est requise (le Remediation_Register est l'artefact Markdown
// `problems.md`, maintenu à la main). Les helpers sont préfixés `fsm83` pour ne
// pas entrer en collision avec la tâche sœur 8.2 du même package.

import (
	"math/rand"
	"testing"
	"testing/quick"
)

// fsm83Status est une valeur de statut du Remediation_Register.
type fsm83Status string

// Domaine fermé des statuts (exigence 3.2) — valeurs exactes du design.
const (
	fsm83Ouvert  fsm83Status = "ouvert"
	fsm83EnCours fsm83Status = "en cours"
	fsm83Cloture fsm83Status = "clôturé"
)

// fsm83Domain est l'ensemble des statuts valides du registre.
var fsm83Domain = map[fsm83Status]bool{
	fsm83Ouvert:  true,
	fsm83EnCours: true,
	fsm83Cloture: true,
}

// fsm83AllowedTransitions encode les seules transitions autorisées de l'automate
// (exigences 3.2, 3.4) : `ouvert → en cours`, `en cours → clôturé` et la
// réouverture `clôturé → ouvert`. Toute autre paire (source, cible) est rejetée.
var fsm83AllowedTransitions = map[fsm83Status]fsm83Status{
	fsm83Ouvert:  fsm83EnCours,
	fsm83EnCours: fsm83Cloture,
	fsm83Cloture: fsm83Ouvert,
}

// fsm83InDomain indique si un statut appartient au domaine fermé du registre.
func fsm83InDomain(s fsm83Status) bool {
	return fsm83Domain[s]
}

// fsm83Transition est la fonction de transition PURE de l'automate de statut.
//
// Elle ne mute aucun état partagé : pour une (current, target) donnée, elle
// retourne le statut résultant et un booléen d'acceptation. Une transition est
// acceptée si et seulement si `current` est dans le domaine et que la paire
// (current → target) figure dans `fsm83AllowedTransitions`. Lorsqu'une
// transition est rejetée, le statut reste inchangé : l'automate ne quitte jamais
// le domaine `{ouvert, en cours, clôturé}`.
func fsm83Transition(current, target fsm83Status) (next fsm83Status, accepted bool) {
	if !fsm83InDomain(current) {
		// Garde défensive : un statut courant hors domaine ne transitionne pas.
		return current, false
	}
	if want, ok := fsm83AllowedTransitions[current]; ok && want == target {
		return target, true
	}
	return current, false
}

// fsm83candidates est l'alphabet des cibles de transition générées. Il inclut
// les trois statuts valides ET une valeur hors domaine, afin de prouver que les
// requêtes invalides sont rejetées et que l'automate reste dans le domaine.
var fsm83candidates = []fsm83Status{
	fsm83Ouvert,
	fsm83EnCours,
	fsm83Cloture,
	fsm83Status("statut-invalide"),
}

// fsm83CandidateFromByte mappe un octet généré vers une cible de transition.
func fsm83CandidateFromByte(b byte) fsm83Status {
	return fsm83candidates[int(b)%len(fsm83candidates)]
}

// TestProperty20_RegisterStatusAutomaton vérifie l'automate de statut du
// Remediation_Register sur des séquences de transitions générées.
//
// Property 20 (Validates: Requirements 3.2, 3.4) : pour toute séquence, seules
// les transitions autorisées sont acceptées et le statut demeure dans le
// domaine fermé `{ouvert, en cours, clôturé}`.
func TestProperty20_RegisterStatusAutomaton(t *testing.T) {
	property := func(raw []byte) bool {
		// L'état initial du registre est `ouvert` ([*] --> ouvert dans le design).
		current := fsm83Ouvert

		for _, b := range raw {
			target := fsm83CandidateFromByte(b)
			next, accepted := fsm83Transition(current, target)

			// Invariant 1 : le statut résultant reste toujours dans le domaine.
			if !fsm83InDomain(next) {
				return false
			}

			// L'acceptation attendue, dérivée indépendamment de la table des
			// transitions autorisées (exigences 3.2, 3.4).
			want, hasSuccessor := fsm83AllowedTransitions[current]
			expectedAccept := hasSuccessor && want == target

			// Invariant 2 : une transition est acceptée ssi elle est autorisée.
			if accepted != expectedAccept {
				return false
			}

			// Invariant 3 : transition acceptée → le statut devient la cible ;
			// transition rejetée → le statut reste inchangé.
			if accepted {
				if next != target {
					return false
				}
			} else if next != current {
				return false
			}

			current = next
		}

		return true
	}

	// Générateurs déterministes (graine fixe) et ≥ 100 itérations.
	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(0x5052_3230)), // "PR20"
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 20 violée : %v", err)
	}
}

// TestProperty20_AllowedTransitionsExplicit complète la propriété par des
// exemples explicites : les trois transitions autorisées avancent l'état, et un
// échantillon de transitions interdites laisse le statut inchangé.
func TestProperty20_AllowedTransitionsExplicit(t *testing.T) {
	accepted := []struct {
		from, to fsm83Status
	}{
		{fsm83Ouvert, fsm83EnCours},
		{fsm83EnCours, fsm83Cloture},
		{fsm83Cloture, fsm83Ouvert},
	}
	for _, tc := range accepted {
		next, ok := fsm83Transition(tc.from, tc.to)
		if !ok {
			t.Errorf("transition %q → %q devrait être acceptée", tc.from, tc.to)
		}
		if next != tc.to {
			t.Errorf("transition %q → %q : statut résultant = %q, attendu %q", tc.from, tc.to, next, tc.to)
		}
	}

	rejected := []struct {
		from, to fsm83Status
	}{
		{fsm83Ouvert, fsm83Cloture},                  // saut interdit
		{fsm83Ouvert, fsm83Ouvert},                   // auto-transition interdite
		{fsm83EnCours, fsm83Ouvert},                  // retour arrière interdit
		{fsm83Cloture, fsm83Cloture},                 // auto-transition interdite
		{fsm83Cloture, fsm83EnCours},                 // saut interdit
		{fsm83Ouvert, fsm83Status("statut-inconnu")}, // cible hors domaine
	}
	for _, tc := range rejected {
		next, ok := fsm83Transition(tc.from, tc.to)
		if ok {
			t.Errorf("transition %q → %q devrait être rejetée", tc.from, tc.to)
		}
		if next != tc.from {
			t.Errorf("transition rejetée %q → %q : statut = %q, attendu inchangé %q", tc.from, tc.to, next, tc.from)
		}
		if !fsm83InDomain(next) {
			t.Errorf("statut résultant %q hors domaine après rejet", next)
		}
	}
}
