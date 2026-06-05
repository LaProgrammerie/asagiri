package onboarding_test

// Feature: audit-coherence-consolidation, Property 17
//
// P17 — Non-dégradation de la Readiness.
//
// Pour tout état initial de dépôt sur lequel l'Onboarding_Wizard applique sa
// configuration, le score de Readiness après application est supérieur ou égal
// au score précédant l'application.
//
// Frontière d'application du wizard retenue : onboarding.ApplyReadinessAutofixes
// — le chemin d'application automatique sûr du wizard (déclenché par
// `asa ready --autofix` / `asa onboard --autofix`). Il ne requiert ni dépôt git
// ni chdir, contient ses effets de bord dans le dépôt temporaire et renvoie un
// Report ré-évalué, ce qui en fait une frontière déterministe adaptée à un test
// de propriété (≥ 100 itérations, générateurs testing/quick).
//
// Validates: Requirements 7.1, 7.2

import (
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

// p17State décrit un état initial de dépôt généré de façon déterministe.
type p17State struct {
	gitignore int  // 0 absent · 1 vide · 2 state seul · 3 worktrees seul · 4 les deux
	docs      int  // 0 aucun · 1 placeholder · 2 réel/substantiel
	kiroSpec  bool // présence d'au moins une feature sous .kiro/specs/
}

// Generate satisfait quick.Generator pour produire des états variés mais bornés.
func (p17State) Generate(r *rand.Rand, _ int) reflect.Value {
	return reflect.ValueOf(p17State{
		gitignore: r.Intn(5),
		docs:      r.Intn(3),
		kiroSpec:  r.Intn(2) == 0,
	})
}

var _ quick.Generator = p17State{}

// p17RealDoc est un contenu produit (non placeholder) suffisamment substantiel
// pour que isPlaceholderContent retourne false (aucun marqueur, ≥ 15 lignes).
const p17RealDoc = `# Produit

Asagiri orchestre des workflows de développement agentique en local.

## Utilisateurs

Les mainteneurs pilotent specs, tasks et revues depuis un seul binaire.

## Valeur

Décrire un besoin, produire des artefacts, valider aux jalons clés.

## Périmètre

Le moteur reste local-first et déterministe pour des entrées identiques.

## Contraintes

Les sorties plain et json conservent la même information.

## Jalons

La confirmation humaine garde la main sur les actions sensibles.
`

func p17Write(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0o644))
}

// p17Materialize écrit l'état initial dans repo.
func p17Materialize(t *testing.T, repo string, s p17State) {
	t.Helper()

	// .gitignore : seule dimension réparable par les autofixes.
	switch s.gitignore {
	case 1:
		p17Write(t, filepath.Join(repo, ".gitignore"), "")
	case 2:
		p17Write(t, filepath.Join(repo, ".gitignore"), ".asagiri/state.sqlite\n")
	case 3:
		p17Write(t, filepath.Join(repo, ".gitignore"), "worktrees/\n")
	case 4:
		p17Write(t, filepath.Join(repo, ".gitignore"), ".asagiri/state.sqlite\nworktrees/\n")
	default:
		// cas 0 : aucun fichier .gitignore
	}

	// docs/ai/01-product.md : fait varier la base de score (non réparable).
	switch s.docs {
	case 1:
		p17Write(t, filepath.Join(repo, "docs", "ai", "01-product.md"),
			"# Template\n\nplaceholder — remplace ce paragraphe.\n")
	case 2:
		p17Write(t, filepath.Join(repo, "docs", "ai", "01-product.md"), p17RealDoc)
	}

	// .kiro/specs : présence d'une feature (non réparable par autofix).
	if s.kiroSpec {
		require.NoError(t, os.MkdirAll(filepath.Join(repo, ".kiro", "specs", "demo"), 0o755))
	}
}

func TestP17ReadinessNonDegradation(t *testing.T) {
	var failDetail string

	prop := func(s p17State) bool {
		repo := t.TempDir()
		p17Materialize(t, repo, s)

		// Score avant : même chargement de config que ApplyReadinessAutofixes
		// (aucune config.yaml écrite → cfg nil), assessment non strict.
		before, err := onboarding.AssessReadiness(repo, nil, false)
		if err != nil {
			failDetail = fmt.Sprintf("AssessReadiness(avant) erreur: %v (state=%+v)", err, s)
			return false
		}

		// Application du wizard (autofixes sûrs) puis ré-évaluation.
		_, after, err := onboarding.ApplyReadinessAutofixes(repo)
		if err != nil {
			failDetail = fmt.Sprintf("ApplyReadinessAutofixes erreur: %v (state=%+v)", err, s)
			return false
		}

		if after.Score < before.Score {
			failDetail = fmt.Sprintf("readiness dégradée: avant=%d après=%d (state=%+v)",
				before.Score, after.Score, s)
			return false
		}
		return true
	}

	cfg := &quick.Config{MaxCount: 200, Rand: rand.New(rand.NewSource(17))}
	if err := quick.Check(prop, cfg); err != nil {
		t.Fatalf("Property 17 (non-dégradation de la Readiness) violée: %v\n%s", err, failDetail)
	}
}
