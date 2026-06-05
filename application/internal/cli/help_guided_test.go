// Feature: audit-coherence-consolidation
//
// Test exemple (R6 — AUD-007) : le help racine met en avant le Guided_Path
// (« Pour commencer ») et aucune Unitary_Command de référence n'est retirée de
// l'arbre Cobra. Garde anti-régression pour les exigences 6.1 et 6.2.
package cli

import (
	"strings"
	"testing"
)

// guidedReferenceUnitaryCommands est l'ensemble des Unitary_Command de référence
// qui doivent rester exécutables de façon autonome (exigence 6.2). Nom unique au
// fichier pour éviter toute collision avec les autres tests du package cli.
var guidedReferenceUnitaryCommands = []string{
	"spec", "plan", "enrich", "dev", "verify", "review",
}

// TestGuidedPathBlockPresentInRootLong vérifie que le texte long de la racine
// contient le bloc « Pour commencer » décrivant le Guided_Path unique et
// découvrable (onboarding → work → jalons de validation). _Requirements: 6.1_
func TestGuidedPathBlockPresentInRootLong(t *testing.T) {
	for _, snippet := range []string{
		"Pour commencer (chemin guidé) :",
		"asa onboard",
		"asa work",
		"validation humaine",
		"Commandes unitaires (toujours disponibles)",
	} {
		if !strings.Contains(rootLong, snippet) {
			t.Fatalf("rootLong ne contient pas le fragment Guided_Path %q\n---\n%s", snippet, rootLong)
		}
	}
}

// TestUnitaryCommandsPreservedInRootCommand garantit que chaque Unitary_Command
// de référence reste enregistrée dans RootCommand(), c'est-à-dire que la mise en
// avant du Guided_Path n'a retiré aucune commande unitaire. _Requirements: 6.2_
func TestUnitaryCommandsPreservedInRootCommand(t *testing.T) {
	root := RootCommand()

	registered := make(map[string]bool)
	for _, cmd := range root.Commands() {
		registered[cmd.Name()] = true
	}

	for _, name := range guidedReferenceUnitaryCommands {
		if !registered[name] {
			t.Errorf("Unitary_Command %q absente de RootCommand() : régression du Guided_Path", name)
		}
	}
}
