Spec — Asagiri Execution Graph Planner

1. Vision

L’Execution Graph Planner ajoute à Asagiri une couche de planification d’exécution basée sur un graphe de dépendances, de risques, de coûts et de validations.

Le but est de remplacer une approche linéaire :

spec
  ↓
tasks
  ↓
run one task after another

par une approche orchestrée :

intent / flow / spec
  ↓
execution graph
  ↓
dependency-aware planning
  ↓
agent assignment
  ↓
parallelizable execution
  ↓
checkpoints
  ↓
validation gates
  ↓
rollback / recovery
  ↓
verified delivery

Asagiri ne doit plus seulement exécuter une liste de tâches. Il doit construire un plan d’exécution structuré, compréhensible, optimisé et vérifiable.

⸻

2. Objectifs

Cette évolution doit permettre de :

1. transformer specs, flows et tasks en graphe d’exécution ;
2. détecter les dépendances entre tâches ;
3. identifier les tâches parallélisables ;
4. estimer coût, durée, risque et blast radius par nœud ;
5. choisir les agents selon compétence, coût, risque et contexte ;
6. définir des checkpoints ;
7. ajouter des gates de validation ;
8. produire un plan de rollback ;
9. reprendre un plan interrompu ;
10. générer un rapport d’exécution graphé.

⸻

3. Non-objectifs

Cette spec ne doit pas chercher à :

* créer un orchestrateur distribué cloud ;
* remplacer GitHub Actions ;
* remplacer les agents existants ;
* ajouter une queue distante obligatoire ;
* exécuter plusieurs agents en parallèle sans isolation ;
* supprimer les primitives existantes spec / plan / enrich / dev / verify / review ;
* introduire de la magie opaque dans le planning.

Le planner doit rester local-first, inspectable et déterministe autant que possible.

⸻

4. Principe central

Le plan d’exécution devient un graphe dirigé.

Chaque nœud représente une unité exécutable :

* investigation ;
* enrichment ;
* implementation ;
* validation ;
* review ;
* architecture derivation ;
* contract validation ;
* documentation ;
* release check.

Chaque arête représente une dépendance :

* requires ;
* blocks ;
* validates ;
* produces_context_for ;
* must_run_after ;
* can_run_after ;
* rollback_depends_on.

⸻

5. Commandes CLI

5.1 asa plan graph

Générer un graphe d’exécution sans exécuter.

asa plan graph workspace-saas

Options :

asa plan graph workspace-saas \
  --flow onboarding \
  --from-product \
  --from-spec \
  --include-reviews \
  --include-docs \
  --estimate \
  --output markdown

⸻

5.2 asa plan explain

Expliquer pourquoi le plan est construit ainsi.

asa plan explain workspace-saas

Sortie attendue :

Execution Plan Explanation
──────────────────────────
Task task-003 depends on task-001 because it requires the Workspace aggregate.
Task task-004 can run in parallel with task-005 because they touch independent paths.
Security review is required because invite_member is a sensitive action.
Observability validation is required because onboarding has success metrics.

⸻

5.3 asa graph run

Exécuter le graphe.

asa graph run workspace-saas

Options :

asa graph run workspace-saas \
  --max-parallel 2 \
  --stop-on-risk high \
  --strict-trust \
  --budget 1.00 \
  --checkpoint-every node \
  --dry-run

⸻

5.4 asa graph status

Afficher l’état d’un graphe en cours ou interrompu.

asa graph status graph-2026-05-27-001

⸻

5.5 asa graph resume

Reprendre un graphe interrompu.

asa graph resume graph-2026-05-27-001

⸻

5.6 asa graph visualize

Exporter le graphe.

asa graph visualize graph-2026-05-27-001 --format mermaid

Formats :

mermaid
json
dot
markdown

⸻

6. Modèle d’exécution

Créer un artefact :

.asagiri/graphs/<graph-id>/execution-graph.yaml
.asagiri/graphs/<graph-id>/execution-graph.json
.asagiri/graphs/<graph-id>/plan.md
.asagiri/graphs/<graph-id>/events.jsonl
.asagiri/graphs/<graph-id>/report.md

⸻

7. Format execution-graph.yaml

Exemple :

id: graph-2026-05-27-001
product: workspace-saas
flow: onboarding
status: planned
created_at: "2026-05-27T00:00:00+02:00"
strategy:
  max_parallel: 2
  stop_on_risk: high
  strict_trust: true
  budget: 1.00
nodes:
  - id: investigate-onboarding
    type: investigation
    title: Investigate onboarding flow
    agent: local
    risk: low
    estimated_cost: 0.00
    estimated_duration: 30s
    outputs:
      - context_pack
  - id: implement-workspace-create
    type: implementation
    title: Implement workspace creation endpoint
    task: task-001
    agent: cursor
    risk: medium
    estimated_cost: 0.08
    estimated_duration: 4m
    required_checks:
      - tests
      - static-analysis
  - id: implement-invite-member
    type: implementation
    title: Implement member invitation
    task: task-002
    agent: cursor
    risk: high
    estimated_cost: 0.12
    estimated_duration: 6m
    required_checks:
      - tests
      - security
      - observability
  - id: verify-onboarding-flow
    type: verification
    title: Verify onboarding flow integrity
    agent: local
    risk: medium
    required_checks:
      - flows
      - contracts
      - trust
edges:
  - from: investigate-onboarding
    to: implement-workspace-create
    type: produces_context_for
  - from: implement-workspace-create
    to: implement-invite-member
    type: requires
    reason: invitation requires workspace ownership model
  - from: implement-invite-member
    to: verify-onboarding-flow
    type: validates
checkpoints:
  - after: investigate-onboarding
  - after: implement-workspace-create
  - after: implement-invite-member
rollback:
  strategy: worktree_reset
  preserve_reports: true

⸻

8. Node types

Types de nœuds supportés :

investigation
enrichment
architecture_derivation
contract_generation
implementation
validation
review
trust_verification
documentation
release_check
manual_approval
rollback

⸻

9. Edge types

Types d’arêtes :

requires
blocks
validates
produces_context_for
must_run_after
can_run_after
parallel_with
rollback_depends_on
requires_human_approval

⸻

10. Dependency inference

Le planner doit inférer les dépendances depuis :

* flows ;
* tasks ;
* contracts ;
* OpenAPI ;
* events ;
* permissions ;
* fichiers touchés ;
* architecture projection ;
* graphes de dépendances code ;
* tests associés ;
* historique mémoire.

Exemples :

task B uses API generated by task A → B requires A
both tasks modify same file → no parallel execution
flow action requires permission missing → permission task blocks implementation
security-sensitive action → security review required
public API contract change → backward compatibility check required

⸻

11. Parallelization strategy

Le planner doit identifier les nœuds parallélisables.

Critères :

* scopes fichiers disjoints ;
* flows indépendants ;
* pas de dépendance contractuelle ;
* pas de shared migration ;
* pas de shared config ;
* risque acceptable ;
* worktrees séparés ;
* budget disponible.

Règles :

* ne jamais paralléliser deux tâches qui modifient le même fichier ;
* ne jamais paralléliser deux migrations DB ;
* ne jamais paralléliser deux changements sur le même contrat public sans gate ;
* limiter parallélisme par défaut à 2 ;
* mode CI peut forcer --max-parallel 1.

⸻

12. Agent assignment

Le planner doit choisir les agents selon :

* type de nœud ;
* complexité ;
* risque ;
* budget ;
* historique réussite ;
* disponibilité ;
* taille contexte ;
* besoin local/cloud.

Exemple :

assignment:
  investigation: local
  enrichment: ollama
  implementation.medium: cursor
  implementation.high: claude
  review.security: codex
  trust_verification: local

⸻

13. Cost-aware planning

Le graphe doit être estimable avant exécution.

Sortie :

Execution Graph Estimate
────────────────────────
Nodes:              12
Parallel groups:    3
Estimated duration: 18m
Estimated cost:     €0.42
Highest risk:       high
Budget status:      OK

Le coût doit être agrégé :

* par nœud ;
* par agent ;
* par flow ;
* par phase ;
* par branche.

⸻

14. Risk-aware planning

Chaque nœud doit avoir :

* risk level ;
* blast radius ;
* required checks ;
* human approval requirement ;
* rollback strategy.

Exemple :

risk:
  level: high
  reasons:
    - touches permission model
    - impacts onboarding critical flow
    - emits public event
  required_approval: true

⸻

15. Checkpoints

Le graphe doit définir des checkpoints.

Un checkpoint doit contenir :

* état Git ;
* état worktree ;
* outputs produits ;
* validations passées ;
* coût consommé ;
* temps consommé ;
* contexte utilisé.

Objectifs :

* reprise ;
* rollback ;
* audit ;
* comparaison branches.

⸻

16. Rollback planning

Chaque nœud à risque doit avoir une stratégie de rollback.

Stratégies :

worktree_reset
patch_revert
migration_down
feature_flag_disable
manual

Le planner doit signaler les nœuds sans rollback clair.

⸻

17. Integration with Trust Engine

Le planner doit intégrer le Trust & Verification Engine.

Règles :

* si un nœud modifie un contract public → trust gate obligatoire ;
* si un flow critique est impacté → flow integrity check obligatoire ;
* si sécurité sensible → security check obligatoire ;
* si observability metrics absentes → gate warning/block selon politique.

⸻

18. Integration with Investigation Engine

Le planner doit pouvoir insérer automatiquement des nœuds d’investigation.

Cas :

* tâche ambiguë ;
* échec précédent ;
* faible confiance ;
* contexte trop large ;
* risque élevé ;
* blast radius inconnu.

Exemple :

Low confidence task detected.
Inserting investigation node before implementation.

⸻

19. Integration with Runtime

Le graphe doit être exécutable par le runtime persistant.

Le runtime doit :

* enregistrer les événements ;
* mettre à jour l’état des nœuds ;
* gérer workers ;
* appliquer limites parallèles ;
* suivre coûts ;
* permettre attach/resume ;
* publier métriques.

Événements :

graph.created
graph.started
graph.node.started
graph.node.completed
graph.node.failed
graph.checkpoint.created
graph.blocked
graph.completed

⸻

20. State machine

États du graphe :

planned
ready
running
blocked
paused
failed
completed
aborted
rolled_back

États de nœud :

pending
ready
running
succeeded
failed
skipped
blocked
rolled_back

⸻

21. Architecture Go

Créer :

internal/executiongraph/
  model.go
  planner.go
  dependency.go
  scheduler.go
  estimator.go
  risk.go
  checkpoints.go
  rollback.go
  renderer.go
  repository.go
  state_machine.go

Interfaces :

type ExecutionGraphPlanner interface {
    Build(ctx context.Context, req GraphPlanRequest) (ExecutionGraph, error)
}
type DependencyInferer interface {
    Infer(ctx context.Context, input DependencyInput) ([]GraphEdge, error)
}
type GraphScheduler interface {
    Schedule(ctx context.Context, graph ExecutionGraph) (ExecutionSchedule, error)
}
type GraphExecutor interface {
    Run(ctx context.Context, graph ExecutionGraph) (GraphRunResult, error)
    Resume(ctx context.Context, graphID string) (GraphRunResult, error)
}

⸻

22. UX terminal cible

Asagiri Execution Graph
═══════════════════════
Product: workspace-saas
Flow:    onboarding
Graph
─────
Nodes:             12
Dependencies:      16
Parallel groups:   3
Checkpoints:       5
Highest risk:      high
Estimated cost:    €0.42
Estimated duration: 18m
Parallel group 1
────────────────
✓ investigate-onboarding        local        0.00€
Parallel group 2
────────────────
○ implement-workspace-create    cursor       0.08€
○ implement-ui-onboarding       cursor       0.06€
Blocked
───────
implement-invite-member waits for implement-workspace-create
security-review waits for implement-invite-member

⸻

23. Reports

Créer :

.asagiri/graphs/<graph-id>/plan.md
.asagiri/graphs/<graph-id>/report.md
.asagiri/graphs/<graph-id>/timeline.jsonl
.asagiri/graphs/<graph-id>/metrics.json

Le rapport doit inclure :

* graphe ;
* décisions de planning ;
* dépendances ;
* agents choisis ;
* coûts estimés/réels ;
* durée estimée/réelle ;
* checkpoints ;
* validations ;
* risques ;
* rollback status.

⸻

24. Configuration

Ajouter :

execution_graph:
  enabled: true
  max_parallel: 2
  default_strategy: risk_aware
  require_checkpoints: true
  stop_on_risk: high
  allow_parallel_agents: true
  require_isolated_worktrees: true
  gates:
    trust_required_for_high_risk: true
    human_approval_for:
      - migration
      - security_sensitive
      - public_contract_change
  rollback:
    require_strategy_for_high_risk: true
    preserve_failed_worktrees: true

⸻

25. CI mode

Le graph planner doit supporter CI.

Usage :

asa plan graph workspace-saas --ci --json
asa graph run workspace-saas --ci --max-parallel 1

En CI :

* pas d’interaction ;
* sortie JSON ;
* erreurs structurées ;
* parallélisme conservateur ;
* gates bloquants.

⸻

26. Tests

Unit tests

* parsing graph ;
* dependency inference ;
* cycle detection ;
* scheduling ;
* risk scoring ;
* rollback strategy validation.

Golden tests

Fixtures :

testdata/execution-graph/
  simple-linear/
  parallel-independent/
  blocked-by-contract/
  high-risk-security/
  rollback-required/

Integration tests

asa plan graph workspace-saas --flow onboarding
asa graph visualize <id> --format mermaid
asa graph run <id> --dry-run
asa graph resume <id>

⸻

27. Critères d’acceptation

Cette évolution est acceptable si :

* asa plan graph génère un graphe valide ;
* les cycles sont détectés ;
* les dépendances principales sont inférées ;
* les tâches parallélisables sont identifiées ;
* les tâches conflictuelles ne sont pas parallélisées ;
* le coût est estimé par nœud et globalement ;
* les risques sont visibles ;
* les checkpoints sont générés ;
* les gates trust peuvent bloquer ;
* le graphe peut être exporté en Mermaid/JSON ;
* asa graph run --dry-run fonctionne ;
* asa graph resume reprend un état interrompu ;
* les rapports sont générés ;
* les tests unitaires/golden/integration passent.

⸻

28. Découpage d’implémentation recommandé

Phase 1 — Graph model

* types Go ;
* YAML/JSON ;
* repository ;
* renderer ;
* tests parsing.

Phase 2 — Dependency inference V1

* dependencies depuis tasks ;
* dependencies depuis flow actions ;
* conflits fichiers ;
* cycle detection.

Phase 3 — Scheduling V1

* topological sort ;
* parallel groups ;
* max parallel ;
* blocked nodes.

Phase 4 — Cost/risk estimation

* coût par nœud ;
* durée ;
* risk ;
* blast radius minimal.

Phase 5 — CLI commands

* asa plan graph ;
* asa plan explain ;
* asa graph visualize ;
* asa graph run --dry-run.

Phase 6 — Runtime integration

* graph state machine ;
* events ;
* checkpoints ;
* resume.

Phase 7 — Trust/Investigation integration

* insertion nœuds investigation ;
* gates trust ;
* rollback planning.

⸻

29. Risques

Plan trop complexe

Mitigation :

* expliquer le plan ;
* Mermaid export ;
* dry-run ;
* stratégies simples par défaut.

Mauvaise parallélisation

Mitigation :

* worktrees isolés ;
* conflits fichiers ;
* max_parallel conservateur ;
* désactivation possible.

Surcoût d’orchestration

Mitigation :

* mode linéaire fallback ;
* graph only pour features complexes ;
* cache ;
* plan rapide.

Faux sentiment de contrôle

Mitigation :

* gates ;
* evidence ;
* reports ;
* human approval sur risques forts.

⸻

30. Résumé

Cette évolution transforme Asagiri d’un exécuteur de tâches en orchestrateur de graphes d’exécution.

Avant :

task list
  ↓
sequential execution

Après :

execution graph
  ↓
dependency-aware scheduling
  ↓
risk/cost-aware planning
  ↓
parallel execution where safe
  ↓
checkpoints
  ↓
trust gates
  ↓
verified delivery

Principe clé :

Asagiri ne doit pas seulement savoir quoi faire. Il doit savoir dans quel ordre, avec quel agent, avec quel risque, avec quel coût, sous quelles validations, et avec quelle stratégie de reprise.