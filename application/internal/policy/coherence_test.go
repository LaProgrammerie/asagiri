package policy

// Feature: audit-coherence-consolidation, Property 10 (P10): Cohérence policy ↔ canon.
//
// Pour tout rôle figurant dans les listes Ollama autorisées ou interdites, le
// Policy_Coherence_Check réussit si et seulement si ce rôle correspond à une
// entrée du canon courant ; sinon il échoue en NOMMANT le rôle divergent.
//
// Le canon courant est la source canonique unique du package
// (ollamaRoleCanon, dont dérivent OllamaAllowedRoles / OllamaForbiddenRoles).
// Le Policy_Coherence_Check est un contrôle de niveau test (R5) : il ne vit pas
// dans le binaire runtime. On l'implémente donc ici, sous forme d'une petite
// fonction pure qui vérifie l'alignement et nomme tout rôle divergent.
//
// Validates: Requirements 5.2, 5.3

import (
	"fmt"
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"
)

// p10CanonSet construit l'ensemble des rôles reconnus par le canon courant à
// partir de la source canonique unique du package (ollamaRoleCanon). C'est la
// référence contre laquelle le Policy_Coherence_Check valide les listes.
func p10CanonSet() map[string]struct{} {
	canon := make(map[string]struct{}, len(ollamaRoleCanon.Allowed)+len(ollamaRoleCanon.Forbidden))
	for _, role := range ollamaRoleCanon.Allowed {
		canon[role] = struct{}{}
	}
	for _, role := range ollamaRoleCanon.Forbidden {
		canon[role] = struct{}{}
	}
	return canon
}

// p10CanonRoles retourne les rôles du canon courant sous forme de tranche, pour
// alimenter le générateur déterministe.
func p10CanonRoles() []string {
	roles := make([]string, 0, len(ollamaRoleCanon.Allowed)+len(ollamaRoleCanon.Forbidden))
	roles = append(roles, ollamaRoleCanon.Allowed...)
	roles = append(roles, ollamaRoleCanon.Forbidden...)
	return roles
}

// p10CheckCoherence est le Policy_Coherence_Check : il vérifie que chaque rôle
// des listes allowed/forbidden correspond à une entrée du canon courant. Au
// premier rôle divergent rencontré, il retourne une erreur NOMMANT ce rôle ;
// sinon il retourne nil. Il ne panique jamais.
func p10CheckCoherence(allowed, forbidden []string, canon map[string]struct{}) error {
	for _, role := range allowed {
		if _, ok := canon[role]; !ok {
			return fmt.Errorf("policy: rôle autorisé %q absent du canon courant (docs/ai/)", role)
		}
	}
	for _, role := range forbidden {
		if _, ok := canon[role]; !ok {
			return fmt.Errorf("policy: rôle interdit %q absent du canon courant (docs/ai/)", role)
		}
	}
	return nil
}

// p10RoleSets est l'entrée générée par testing/quick : un couple de listes
// (allowed, forbidden) composé d'un sous-ensemble du canon, éventuellement
// pollué de rôles divergents injectés (hors canon).
type p10RoleSets struct {
	Allowed   []string
	Forbidden []string
}

// Generate produit des listes déterministes : sous-ensemble aléatoire du canon
// courant, plus 0..2 rôles divergents injectés dont le préfixe garantit
// l'absence du canon. Implémente quick.Generator.
func (p10RoleSets) Generate(rnd *rand.Rand, size int) reflect.Value {
	canon := p10CanonRoles()
	canonSet := p10CanonSet()

	var allowed, forbidden []string
	for _, role := range canon {
		switch rnd.Intn(3) {
		case 0:
			allowed = append(allowed, role)
		case 1:
			forbidden = append(forbidden, role)
		default:
			// rôle omis de cette itération
		}
	}

	nDivergent := rnd.Intn(3)
	for i := 0; i < nDivergent; i++ {
		divergent := fmt.Sprintf("p10-divergent-role-%d-%d", size, rnd.Intn(1<<20))
		// Le préfixe rend la collision impossible ; on reste défensif.
		if _, clash := canonSet[divergent]; clash {
			continue
		}
		if rnd.Intn(2) == 0 {
			allowed = append(allowed, divergent)
		} else {
			forbidden = append(forbidden, divergent)
		}
	}

	return reflect.ValueOf(p10RoleSets{Allowed: allowed, Forbidden: forbidden})
}

// p10DivergentRoles retourne les rôles des listes absents du canon.
func p10DivergentRoles(rs p10RoleSets, canon map[string]struct{}) []string {
	var divergent []string
	for _, role := range append(append([]string{}, rs.Allowed...), rs.Forbidden...) {
		if _, ok := canon[role]; !ok {
			divergent = append(divergent, role)
		}
	}
	return divergent
}

// TestProperty10PolicyCanonCoherence vérifie la Property 10 : le
// Policy_Coherence_Check réussit ssi tous les rôles sont dans le canon ; sinon
// il échoue en nommant un rôle divergent.
//
// Feature: audit-coherence-consolidation, Property 10 (P10).
// Validates: Requirements 5.2, 5.3
func TestProperty10PolicyCanonCoherence(t *testing.T) {
	canon := p10CanonSet()

	property := func(rs p10RoleSets) bool {
		err := p10CheckCoherence(rs.Allowed, rs.Forbidden, canon)
		divergent := p10DivergentRoles(rs, canon)
		hasDivergent := len(divergent) > 0

		// Biconditionnel (5.2/5.3) : le check réussit ssi aucun rôle divergent.
		if (err == nil) != !hasDivergent {
			return false
		}
		// En cas d'échec, l'erreur doit NOMMER un rôle effectivement divergent.
		if err != nil {
			for _, role := range divergent {
				if strings.Contains(err.Error(), role) {
					return true
				}
			}
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200,
		Rand:     rand.New(rand.NewSource(0x10)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 10 (cohérence policy ↔ canon) violée : %v", err)
	}
}

// TestPolicyCoherenceRealListsAligned est un test exemple : sur le dépôt
// corrigé, les listes exportées réelles sont alignées sur le canon courant, donc
// le Policy_Coherence_Check passe sans nommer aucune divergence (5.1/5.2).
func TestPolicyCoherenceRealListsAligned(t *testing.T) {
	canon := p10CanonSet()
	if err := p10CheckCoherence(OllamaAllowedRoles, OllamaForbiddenRoles, canon); err != nil {
		t.Fatalf("listes Ollama réelles divergentes du canon courant : %v", err)
	}
}

// TestPolicyCoherenceNamesInjectedDivergentRole est un test exemple : un rôle
// divergent injecté fait échouer le check, et l'erreur nomme précisément ce
// rôle (5.3).
func TestPolicyCoherenceNamesInjectedDivergentRole(t *testing.T) {
	canon := p10CanonSet()
	const divergent = "p10-stale-role-from-historical-spec"

	err := p10CheckCoherence(append([]string{}, append(OllamaAllowedRoles, divergent)...), OllamaForbiddenRoles, canon)
	if err == nil {
		t.Fatalf("attendu un échec pour le rôle divergent %q, obtenu nil", divergent)
	}
	if !strings.Contains(err.Error(), divergent) {
		t.Fatalf("l'erreur doit nommer le rôle divergent %q, obtenu : %v", divergent, err)
	}
}
