package cli

// Feature: audit-coherence-consolidation
//
// Test exemple (tâche 2.10) — `asa explain routing` nomme backend + raison
// (Requirements: 4.8).
//
// Pour quelques `--step-class` (avec et sans flags), on vérifie que la sortie
// PLAIN et la sortie JSON de `asa explain routing` :
//   - nomment l'Agent_Backend retenu et la raison (Requirement 4.8) ;
//   - sont cohérentes avec la `routing.Decision` calculée par `routing.Route`
//     pour les mêmes entrées (mêmes config, classe d'étape et flags) ;
//   - sont en parité plain/json (mêmes agent, modèle, local, raison).
//
// Ce n'est PAS un test property-based : aucun tag `Property`, un seul fichier
// d'exemples déterministes. Les helpers sont préfixés `explain210` pour éviter
// toute collision avec les nombreux autres fichiers de test du package `cli`.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/routing"
)

// explain210ConfigYAML déclare deux Agent_Backend — un local (`ollama`, endpoint
// localhost) et un cloud (`claude`) — plus une stratégie de routage couvrant les
// classes utilisées par les cas de test. Aucun nom d'agent n'est codé en dur dans
// le chemin de décision : `routing.Route` ne sélectionne que des backends
// déclarés ici (Requirement 4.2).
func explain210ConfigYAML() string {
	return `project:
  name: explain-routing-test
  default_branch: main
agents:
  ollama:
    endpoint: http://localhost:11434
    model: llama3.1
  claude:
    command: claude
    model: claude-3-5-sonnet
work:
  default_agent: claude
  default_enricher: ollama
routing:
  default_strategy: cost_aware
  strategies:
    cost_aware:
      prefer_local_for:
        - enrich
      use_cloud_heavy_for:
        - implement
      use_cloud_fast_for:
        - review
`
}

// explain210SetupRepo crée un dépôt Git temporaire avec une config Asagiri
// valide et s'y positionne (chdir) pour la durée du test, comme le font les
// autres tests d'intégration CLI du package.
func explain210SetupRepo(t *testing.T) string {
	t.Helper()
	repo := t.TempDir()
	runGitCommand(t, repo, "init")
	runGitCommand(t, repo, "config", "user.email", "explain210@example.com")
	runGitCommand(t, repo, "config", "user.name", "Explain210")
	writeFile(t, filepath.Join(repo, "go.mod"), "module example.com/explain210\n\ngo 1.25.0\n")
	writeFile(t, filepath.Join(repo, ".asagiri", "config.yaml"), explain210ConfigYAML())

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(repo))
	t.Cleanup(func() { _ = os.Chdir(oldWd) })
	return repo
}

// explain210Case décrit un exemple : une classe d'étape et la combinaison de
// flags non interactifs passés à `asa explain routing`.
type explain210Case struct {
	name        string
	stepClass   string
	preferLocal bool
	noCloud     bool
	allowCloud  bool
}

// explain210Args construit les arguments CLI de base (sortie plain) pour un cas.
func (c explain210Case) explain210Args() []string {
	args := []string{"explain", "routing", "--step-class", c.stepClass}
	if c.preferLocal {
		args = append(args, "--prefer-local")
	}
	if c.noCloud {
		args = append(args, "--no-cloud")
	}
	if c.allowCloud {
		args = append(args, "--allow-cloud")
	}
	return args
}

// explain210Run exécute `asa <args>` sur un arbre de commandes neuf (pour éviter
// toute fuite d'état de flags entre invocations) et renvoie la sortie capturée.
func explain210Run(t *testing.T, args []string) string {
	t.Helper()
	root := newRootCmd()
	out := new(bytes.Buffer)
	root.SetOut(out)
	root.SetErr(out)
	root.SetArgs(args)
	require.NoError(t, root.Execute(), out.String())
	return out.String()
}

func TestExplainRoutingNamesBackendAndReason(t *testing.T) {
	// Feature: audit-coherence-consolidation
	// Validates (exemple) : Requirements 4.8
	repo := explain210SetupRepo(t)

	// Config chargée comme le fait la commande, pour calculer la Decision de
	// référence via routing.Route (cohérence sortie CLI ↔ Route).
	cfg, err := config.Load(config.ConfigPath(repo), repo)
	require.NoError(t, err)

	cases := []explain210Case{
		{name: "prefer_local par classe", stepClass: "enrich"},
		{name: "cloud_heavy par classe", stepClass: "implement"},
		{name: "cloud_fast par classe", stepClass: "review"},
		{name: "prefer_local par flag", stepClass: "review", preferLocal: true},
		{name: "no_cloud prevaut", stepClass: "implement", noCloud: true, allowCloud: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Decision de référence pour les mêmes entrées.
			decision, routeErr := routing.Route(cfg, tc.stepClass, tc.preferLocal, tc.noCloud, tc.allowCloud)
			require.NoError(t, routeErr, "routing.Route doit réussir pour le cas %q", tc.name)
			require.NotEmpty(t, decision.Agent, "la Decision doit nommer un backend")
			require.NotEmpty(t, decision.Reason, "la Decision doit porter une raison")
			// Le backend retenu est bien déclaré dans la config (Requirement 4.2).
			_, declared := cfg.Agents[decision.Agent]
			require.True(t, declared, "le backend %q doit être déclaré dans config.agents", decision.Agent)

			// --- Sortie PLAIN ---
			plain := explain210Run(t, tc.explain210Args())
			require.Contains(t, plain, "agent: "+decision.Agent,
				"la sortie plain doit nommer le backend retenu")
			require.Contains(t, plain, "reason: "+decision.Reason,
				"la sortie plain doit exposer la raison")
			// Le nom du backend et la raison sont présents en clair.
			require.Contains(t, plain, decision.Agent)
			require.Contains(t, plain, decision.Reason)

			// --- Sortie JSON ---
			jsonArgs := append(tc.explain210Args(), "--json")
			rawJSON := explain210Run(t, jsonArgs)
			var got RoutingExplanation
			require.NoError(t, json.Unmarshal([]byte(rawJSON), &got),
				"la sortie JSON doit être décodable: %s", rawJSON)

			// La sortie JSON nomme le backend et la raison, cohérents avec Route.
			require.Equal(t, decision.Agent, got.Agent, "agent JSON ↔ Decision")
			require.Equal(t, decision.Reason, got.Reason, "raison JSON ↔ Decision")
			require.Equal(t, decision.Model, got.Model, "modèle JSON ↔ Decision")
			require.Equal(t, decision.Local, got.Local, "local JSON ↔ Decision")
			require.Equal(t, decision.StepClass, got.StepClass, "step_class JSON ↔ Decision")

			// --- Parité plain ↔ json : le nom du backend et la raison annoncés
			// en JSON apparaissent aussi dans la sortie plain. ---
			require.Contains(t, plain, "agent: "+got.Agent)
			require.Contains(t, plain, "reason: "+got.Reason)
			require.Contains(t, plain, fmt.Sprintf("local: %t", got.Local))
		})
	}
}
