package policy

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// Feature: audit-coherence-consolidation, Property 11
//
// P11 — Refus des rôles interdits sans panic (Validates: Requirements 5.4).
//
// Pour tout rôle, CheckOllamaRole retourne une erreur nommant le rôle si et
// seulement si ce rôle appartient à OllamaForbiddenRoles, retourne nil sinon, et
// ne panique jamais.
//
// Les helpers sont préfixés `p11` afin de ne pas entrer en collision avec les
// autres tests du package (notamment la sous-tâche 3.2, Property 10).

// p11Role est un rôle généré déterministiquement par testing/quick. C'est un
// type string nommé pour que les contre-exemples s'impriment lisiblement.
type p11Role string

// p11RandomAlphabet borne l'espace des chaînes aléatoires générées à des
// caractères de rôle plausibles (lettres minuscules et underscore).
const p11RandomAlphabet = "abcdefghijklmnopqrstuvwxyz_"

// Generate produit un mélange déterministe de rôles interdits, de rôles
// autorisés et de chaînes aléatoires, pour couvrir les deux branches du « ssi »
// ainsi que les rôles inconnus.
func (p11Role) Generate(rng *rand.Rand, _ int) reflect.Value {
	switch rng.Intn(3) {
	case 0:
		if len(OllamaForbiddenRoles) > 0 {
			return reflect.ValueOf(p11Role(OllamaForbiddenRoles[rng.Intn(len(OllamaForbiddenRoles))]))
		}
	case 1:
		if len(OllamaAllowedRoles) > 0 {
			return reflect.ValueOf(p11Role(OllamaAllowedRoles[rng.Intn(len(OllamaAllowedRoles))]))
		}
	}
	return reflect.ValueOf(p11Role(p11RandomRole(rng)))
}

// p11RandomRole construit une chaîne aléatoire (éventuellement vide) à partir
// d'un alphabet de rôle plausible.
func p11RandomRole(rng *rand.Rand) string {
	n := rng.Intn(24)
	var b strings.Builder
	for i := 0; i < n; i++ {
		b.WriteByte(p11RandomAlphabet[rng.Intn(len(p11RandomAlphabet))])
	}
	return b.String()
}

// p11Contains indique si role figure dans roles.
func p11Contains(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}
	return false
}

// p11CheckNoPanic exécute CheckOllamaRole en capturant un éventuel panic, afin
// de prouver que la frontière de policy ne panique jamais (03-standards.md).
func p11CheckNoPanic(role string) (err error, panicked bool) {
	defer func() {
		if rec := recover(); rec != nil {
			panicked = true
		}
	}()
	return CheckOllamaRole(role), false
}

// TestProperty11ForbiddenRolesRefusedWithoutPanic vérifie la Property 11 :
// CheckOllamaRole refuse exactement les rôles interdits (erreur nommant le
// rôle), accepte tous les autres (nil), et ne panique jamais.
func TestProperty11ForbiddenRolesRefusedWithoutPanic(t *testing.T) {
	property := func(r p11Role) bool {
		role := string(r)
		err, panicked := p11CheckNoPanic(role)
		if panicked {
			t.Errorf("P11: CheckOllamaRole a paniqué pour le rôle %q", role)
			return false
		}

		forbidden := p11Contains(OllamaForbiddenRoles, role)
		if forbidden {
			// Branche « interdit » : erreur non nil ET message nommant le rôle.
			if err == nil {
				t.Errorf("P11: rôle interdit %q accepté (err == nil), attendu un refus", role)
				return false
			}
			if !strings.Contains(err.Error(), role) {
				t.Errorf("P11: l'erreur pour le rôle interdit %q ne nomme pas le rôle: %v", role, err)
				return false
			}
			return true
		}

		// Branche « non interdit » : nil attendu.
		if err != nil {
			t.Errorf("P11: rôle non interdit %q refusé: %v", role, err)
			return false
		}
		return true
	}

	// Générateur déterministe (seed fixe) et ≥ 100 itérations.
	cfg := &quick.Config{
		MaxCount: 1000,
		Rand:     rand.New(rand.NewSource(11)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("P11: propriété violée: %v", err)
	}
}
