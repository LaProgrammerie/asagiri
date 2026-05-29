Spec — Asagiri Engineering Knowledge Graph

1. Vision

L’Engineering Knowledge Graph ajoute à Asagiri une couche de connaissance structurelle du système logiciel.

Le but est de ne plus faire raisonner les agents uniquement sur du texte brut, des fichiers isolés ou des prompts longs.

Asagiri doit construire un graphe reliant :

product flows
↔ API contracts
↔ code symbols
↔ events
↔ permissions
↔ tests
↔ infrastructure
↔ observability
↔ reviews
↔ incidents
↔ costs
↔ ADRs

Ce graphe devient une source de contexte fiable pour :

* investigation ;
* planning ;
* review ;
* impact analysis ;
* trust scoring ;
* agent routing ;
* documentation ;
* refactoring.

⸻

2. Objectifs

Cette évolution doit permettre de :

* indexer la structure réelle du projet ;
* relier produit, architecture, code et opérations ;
* détecter les impacts d’un changement ;
* réduire le contexte envoyé aux agents ;
* améliorer investigation et root cause analysis ;
* alimenter le Trust Engine ;
* alimenter l’Execution Graph Planner ;
* rendre les décisions plus explicables.

⸻

3. Non-objectifs

Cette spec ne doit pas chercher à :

* créer une base de connaissance cloud ;
* indexer tout Internet ;
* remplacer les tests ;
* remplacer les specs ;
* rendre le graphe parfait dès la V1 ;
* imposer un modèle unique à tous les langages.

Le graphe doit être local-first, incrémental et tolérant aux informations partielles.

⸻

4. Concepts principaux

4.1 Nodes

Types de nœuds :

product
flow
flow_step
action
screen
api_operation
event
permission
metric
trace
log
module
file
symbol
test
migration
infra_resource
config
secret_boundary
adr
review
incident
cost_center
agent

⸻

4.2 Edges

Types de relations :

implements
calls
emits
requires
validates
tests
observes
configures
depends_on
owns
impacts
breaks
produces
consumes
reviewed_by
failed_in
costs

⸻

5. Arborescence

.asagiri/knowledge/
  graph.sqlite
  graph.json
  indexes/
  snapshots/
  reports/

⸻

6. Commandes CLI

6.1 asa knowledge build

Construire ou mettre à jour le graphe.

asa knowledge build

Options :

asa knowledge build \
  --incremental \
  --scope product:workspace-saas \
  --include-code \
  --include-flows \
  --include-tests \
  --include-infra

⸻

6.2 asa knowledge query

Interroger le graphe.

asa knowledge query "what implements invite_member?"

Exemples :

asa knowledge query "which tests cover onboarding?"
asa knowledge query "what APIs are impacted by WorkspaceService?"
asa knowledge query "which flows use member.invite?"

⸻

6.3 asa impact analyze

Analyser l’impact d’un changement.

asa impact analyze --file src/Invitation/InvitationService.php

Ou :

asa impact analyze --flow onboarding --action invite_member

⸻

6.4 asa knowledge explain

Expliquer pourquoi un élément est lié à un autre.

asa knowledge explain onboarding invite_member InvitationService

⸻

6.5 asa knowledge snapshot

Créer un snapshot du graphe.

asa knowledge snapshot --name before-onboarding-refactor

⸻

7. Graph Storage

V1 : SQLite local.

Tables minimales :

nodes
edges
node_properties
edge_properties
snapshots
index_metadata

Le système doit pouvoir exporter :

JSON
DOT
Mermaid

⸻

8. Node model

type GraphNode struct {
    ID         string
    Type       NodeType
    Name       string
    Path       string
    Properties map[string]any
    Source     GraphSource
    Confidence float64
    CreatedAt  time.Time
    UpdatedAt  time.Time
}

⸻

9. Edge model

type GraphEdge struct {
    ID         string
    From       string
    To         string
    Type       EdgeType
    Properties map[string]any
    Source     GraphSource
    Confidence float64
    CreatedAt  time.Time
}

⸻

10. Sources d’extraction

Le graphe doit être construit depuis :

* .asagiri/products/** ;
* flows YAML ;
* contracts OpenAPI ;
* permissions ;
* events ;
* analytics ;
* observability ;
* specs ;
* tasks ;
* code source ;
* tests ;
* config ;
* infrastructure ;
* ADRs ;
* trust reports ;
* investigation reports ;
* runtime events.

⸻

11. Extractors

Créer :

internal/knowledge/extractors/
  flows.go
  contracts.go
  code.go
  tests.go
  events.go
  permissions.go
  observability.go
  infra.go
  adr.go
  runtime.go

Chaque extractor doit produire :

* nodes ;
* edges ;
* confidence ;
* warnings.

⸻

12. Code graph V1

Le code graph doit commencer simple.

V1 :

* fichiers ;
* symboles principaux ;
* imports ;
* appels évidents ;
* tests associés ;
* routes/endpoints.

Langages prioritaires :

* Go ;
* PHP ;
* TypeScript.

Pas besoin d’un AST parfait en V1.

⸻

13. Flow-to-code linking

Le système doit relier les flows au code.

Exemple :

flow:onboarding
  action:invite_member
    requires api:POST /invitations
      implemented_by symbol:InvitationController::create
      calls symbol:InvitationService::invite
      emits event:member.invited
      tested_by test:InvitationServiceTest

⸻

14. Impact Analysis

asa impact analyze doit produire :

Impact Analysis
───────────────
Input: src/Invitation/InvitationService.php
Impacted flows:
- onboarding / invite_team / invite_member
Impacted APIs:
- POST /invitations
Impacted events:
- member.invited
Impacted tests:
- InvitationServiceTest
- OnboardingFlowTest
Risk:
medium
Recommended checks:
- composer test -- tests/Invitation
- asa verify trust --flow onboarding

⸻

15. Context Retrieval

Le graphe doit alimenter les agents.

Exemple :

asa context build --from-graph --flow onboarding --action invite_member

Sortie :

* fichiers pertinents ;
* symboles ;
* tests ;
* contracts ;
* flows ;
* events ;
* métriques ;
* risques.

⸻

16. Integration avec Investigation Engine

L’investigation doit utiliser le graphe pour :

* résoudre scope ;
* trouver fichiers ;
* trouver tests ;
* relier logs à flows ;
* proposer hypothèses ;
* créer root cause graph.

⸻

17. Integration avec Trust Engine

Le Trust Engine doit utiliser le graphe pour :

* blast radius ;
* regression confidence ;
* flow integrity ;
* contract validation ;
* observability coverage ;
* security impact.

⸻

18. Integration avec Execution Graph Planner

Le planner doit utiliser le graphe pour :

* inférer dépendances ;
* détecter conflits ;
* déterminer parallélisation sûre ;
* assigner validations ;
* prévoir rollback.

⸻

19. Integration avec Multi-Agent Coordination

Le coordinateur doit utiliser le graphe pour :

* choisir agent ;
* réduire contexte ;
* produire handoff ;
* éviter conflits ;
* comparer outputs agents.

⸻

20. Confidence & Provenance

Chaque nœud/relation doit avoir :

* source ;
* confidence ;
* timestamp ;
* extractor ;
* evidence.

Exemple :

edge:
  from: action:invite_member
  to: api:POST /invitations
  type: requires
  confidence: 0.91
  source: flows/onboarding.flow.yaml

⸻

21. Staleness detection

Le graphe doit détecter les données périmées.

Triggers :

* fichier modifié ;
* contract changé ;
* flow changé ;
* tests ajoutés/supprimés ;
* runtime event ;
* ADR modifié.

Sortie :

Knowledge graph stale
─────────────────────
3 files changed since last build
2 edges may be outdated
Run: asa knowledge build --incremental

⸻

22. Architecture Go

Créer :

internal/knowledge/
  graph.go
  node.go
  edge.go
  store.go
  query.go
  builder.go
  snapshot.go
  impact.go
  provenance.go
  staleness.go
  extractors/
  renderers/

Interfaces :

type KnowledgeGraphBuilder interface {
    Build(ctx context.Context, req BuildRequest) (BuildResult, error)
}
type GraphStore interface {
    UpsertNode(ctx context.Context, node GraphNode) error
    UpsertEdge(ctx context.Context, edge GraphEdge) error
    Query(ctx context.Context, query GraphQuery) (GraphQueryResult, error)
}
type ImpactAnalyzer interface {
    Analyze(ctx context.Context, req ImpactRequest) (ImpactResult, error)
}

⸻

23. UX terminal cible

Asagiri Knowledge Graph
═══════════════════════
Build complete
Nodes:        1,284
Edges:        3,912
Sources:      flows, contracts, code, tests, runtime
Confidence:   0.82 avg
Stale:        0
Top warnings
────────────
- 4 flow actions have no linked test
- 2 API operations have no linked permission
- 1 critical metric has no dashboard contract

⸻

24. Tests

Unit tests

* node model ;
* edge model ;
* graph store ;
* query engine ;
* staleness detection.

Golden tests

Fixtures :

testdata/knowledge-graph/
  onboarding-flow/
  api-events/
  missing-tests/
  stale-graph/

Integration tests

asa knowledge build
asa knowledge query "which tests cover onboarding?"
asa impact analyze --flow onboarding
asa knowledge snapshot --name test

⸻

25. Critères d’acceptation

Cette évolution est acceptable si :

* asa knowledge build construit un graphe local ;
* flows, contracts, code et tests produisent des nœuds ;
* les relations principales sont créées ;
* asa knowledge query retourne des résultats utiles ;
* asa impact analyze identifie flows/tests/contracts impactés ;
* les nœuds/edges ont provenance et confidence ;
* le graphe détecte le staleness ;
* le graphe alimente investigation/trust/planning ;
* le mode JSON est disponible ;
* les tests unitaires/golden/intégration passent.

⸻

26. Découpage d’implémentation recommandé

Phase 1 — Graph core

* modèles node/edge ;
* SQLite store ;
* query basique ;
* export JSON.

Phase 2 — Flow & contract extractors

* flows ;
* actions ;
* OpenAPI ;
* permissions ;
* events.

Phase 3 — Code/test extractors V1

* fichiers ;
* symboles simples ;
* imports ;
* tests associés.

Phase 4 — Impact analysis

* input file/symbol/flow/action ;
* impacted nodes ;
* recommended checks.

Phase 5 — Integrations

* Investigation Engine ;
* Trust Engine ;
* Execution Graph Planner.

Phase 6 — Staleness & snapshots

* incremental build ;
* snapshots ;
* stale warnings.

⸻

27. Risques

Graphe trop ambitieux

Mitigation :

* V1 simple ;
* confidence ;
* extraction partielle assumée ;
* amélioration incrémentale.

Faux liens

Mitigation :

* provenance ;
* confidence ;
* warnings ;
* review manuelle possible.

Performance

Mitigation :

* incremental build ;
* cache ;
* extractors bornés ;
* SQLite indexes.

Complexité cognitive

Mitigation :

* commandes simples ;
* reports lisibles ;
* graph invisible sauf besoin ;
* query orientée usage.

⸻

28. Résumé

Cette évolution ajoute à Asagiri une couche de connaissance structurelle locale.

Avant :

files + prompts + specs

Après :

engineering knowledge graph
  ↓
impact analysis
  ↓
context retrieval
  ↓
trust scoring
  ↓
execution planning
  ↓
agent coordination

Principe clé :

Les agents doivent raisonner sur les relations du système, pas seulement sur du texte brut. Asagiri doit construire et maintenir un graphe reliant produit, architecture, code, tests et opérations.