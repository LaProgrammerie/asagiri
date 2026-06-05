// Feature: audit-coherence-consolidation
//
// Test exemple (non property-based) — tâche 8.2 : cohérence du
// Remediation_Register matérialisé dans problems.md à la racine du dépôt.
//
// Il parse la table Remediation_Register et vérifie, pour les constats d'audit
// AUD-* :
//   - exactement une entrée par constat de sévérité error/blocking
//     (AUD-001, AUD-002, AUD-003) ;
//   - toutes les colonnes non vides ;
//   - le statut appartient au domaine fermé {ouvert, en cours, clôturé} ;
//   - aucun constat blocking ne reste au statut ouvert (filet de sécurité de
//     clôture).
//
// _Requirements: 3.1, 3.5, 8.6_
package policy

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// reg82StatusDomain est l'ensemble fermé des statuts valides du
// Remediation_Register (design R3 / exigence 3.2) : exactement "ouvert",
// "en cours" ou "clôturé".
var reg82StatusDomain = map[string]bool{
	"ouvert":   true,
	"en cours": true,
	"clôturé":  true,
}

// reg82ErrorFindings liste les constats error/blocking qui DOIVENT chacun avoir
// exactement une entrée dans le registre (exigence 3.1).
var reg82ErrorFindings = []string{"AUD-001", "AUD-002", "AUD-003"}

// reg82Entry est une ligne parsée du Remediation_Register.
type reg82Entry struct {
	id       string
	zone     string
	probleme string
	severite string
	statut   string
}

// reg82RepoRoot remonte depuis le répertoire du package jusqu'au répertoire
// contenant go.mod (racine du dépôt), où vit problems.md.
func reg82RepoRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	for {
		if _, statErr := os.Stat(filepath.Join(dir, "go.mod")); statErr == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("go.mod introuvable en remontant depuis %q", dir)
		}
		dir = parent
	}
}

// reg82SplitRow découpe une ligne de table Markdown `| a | b | c |` en cellules
// trimées, sans les barres de bord.
func reg82SplitRow(line string) []string {
	trimmed := strings.Trim(strings.TrimSpace(line), "|")
	parts := strings.Split(trimmed, "|")
	cells := make([]string, len(parts))
	for i, p := range parts {
		cells[i] = strings.TrimSpace(p)
	}
	return cells
}

// reg82ParseRegister extrait les entrées AUD-* de la table Remediation_Register.
// Il se limite à la section "## Remediation_Register" (jusqu'au prochain titre
// de niveau 2) pour ne jamais confondre le registre actif avec la table
// d'archive située plus bas dans problems.md.
func reg82ParseRegister(t *testing.T, content string) []reg82Entry {
	t.Helper()
	lines := strings.Split(content, "\n")

	start := -1
	for i, ln := range lines {
		if strings.HasPrefix(strings.TrimSpace(ln), "## Remediation_Register") {
			start = i
			break
		}
	}
	if start < 0 {
		t.Fatalf("section '## Remediation_Register' introuvable dans problems.md")
	}

	// Borne de fin : prochain titre de niveau 2 (exclut la table d'archive).
	end := len(lines)
	for i := start + 1; i < len(lines); i++ {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "## ") {
			end = i
			break
		}
	}

	headerSeen := false
	var entries []reg82Entry
	for i := start; i < end; i++ {
		raw := strings.TrimSpace(lines[i])
		if !strings.HasPrefix(raw, "|") {
			continue
		}
		cells := reg82SplitRow(raw)
		if len(cells) != 5 {
			continue
		}
		// Ligne d'en-tête : on confirme les colonnes attendues.
		if cells[0] == "ID" {
			headerSeen = true
			want := []string{"ID", "Zone", "Problème", "Sévérité", "Statut"}
			for j := range want {
				if cells[j] != want[j] {
					t.Fatalf("en-tête de table inattendu: %v (attendu %v)", cells, want)
				}
			}
			continue
		}
		// Ligne de séparation `| --- | --- | ...`.
		if strings.HasPrefix(cells[0], "---") {
			continue
		}
		// On ne retient que les constats d'audit AUD-*.
		if !strings.HasPrefix(cells[0], "AUD-") {
			continue
		}
		entries = append(entries, reg82Entry{
			id:       cells[0],
			zone:     cells[1],
			probleme: cells[2],
			severite: cells[3],
			statut:   cells[4],
		})
	}

	if !headerSeen {
		t.Fatalf("table Remediation_Register `| ID | Zone | Problème | Sévérité | Statut |` introuvable")
	}
	return entries
}

func TestRemediationRegisterCoherence(t *testing.T) {
	root := reg82RepoRoot(t)
	path := filepath.Join(root, "problems.md")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("lecture de problems.md (%s): %v", path, err)
	}

	entries := reg82ParseRegister(t, string(data))
	if len(entries) == 0 {
		t.Fatalf("aucune entrée AUD-* parsée dans le Remediation_Register")
	}

	counts := make(map[string]int, len(entries))
	for _, e := range entries {
		counts[e.id]++

		// Colonnes non vides.
		if e.id == "" || e.zone == "" || e.probleme == "" || e.severite == "" || e.statut == "" {
			t.Errorf("entrée %q: colonne(s) vide(s): %+v", e.id, e)
		}

		// Statut unique dans le domaine {ouvert, en cours, clôturé}.
		if !reg82StatusDomain[e.statut] {
			t.Errorf("entrée %q: statut %q hors du domaine {ouvert, en cours, clôturé}", e.id, e.statut)
		}

		// Aucun constat blocking au statut ouvert (filet de sécurité de clôture).
		if e.severite == "blocking" && e.statut == "ouvert" {
			t.Errorf("entrée %q: constat 'blocking' au statut 'ouvert' (clôture non garantie)", e.id)
		}
	}

	// Exactement une entrée par constat error/blocking présent dans le registre.
	seen := make(map[string]bool)
	for _, e := range entries {
		if e.severite != "error" && e.severite != "blocking" {
			continue
		}
		if seen[e.id] {
			continue
		}
		seen[e.id] = true
		if counts[e.id] != 1 {
			t.Errorf("constat %q (%s): %d entrées, attendu exactement 1", e.id, e.severite, counts[e.id])
		}
	}

	// Les constats error connus (AUD-001/002/003) sont chacun présents une fois.
	for _, id := range reg82ErrorFindings {
		if counts[id] != 1 {
			t.Errorf("constat error %q: %d entrée(s), attendu exactement 1", id, counts[id])
		}
	}
}
