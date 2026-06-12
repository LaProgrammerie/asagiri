Spec — Asagiri Replay & Deterministic Execution

1. Vision

Le Replay & Deterministic Execution System ajoute à Asagiri une capacité de reproductibilité des workflows d’ingénierie.

Le but est de pouvoir :

* rejouer ;
* auditer ;
* comparer ;
* debugger ;
* valider ;
* expliquer.

Asagiri doit être capable de rejouer un workflow avec :

* le même contexte ;
* les mêmes inputs ;
* les mêmes artefacts ;
* les mêmes versions ;
* les mêmes agents ;
* les mêmes policies.

Pattern cible :

intent
  ↓
execution
  ↓
artifacts
  ↓
replay package
  ↓
re-execution
  ↓
comparison
  ↓
auditability

⸻

2. Objectifs

Cette évolution doit permettre :

* replay workflows ;
* replay investigations ;
* replay trust verification ;
* replay execution graphs ;
* replay multi-agent pipelines ;
* reproduire failures ;
* comparer deux exécutions ;
* debugger comportements agents ;
* auditer outputs ;
* améliorer confiance.

⸻

3. Non-objectifs

Cette spec ne doit pas chercher à :

* garantir un déterminisme parfait des LLMs ;
* reproduire exactement des APIs externes ;
* remplacer Git ;
* snapshotter toute la machine ;
* créer un orchestrateur cloud distribué.

Le système doit viser un :

practical deterministic engineering replay

⸻

4. Concepts principaux

Replay Package

Bundle permettant de rejouer un workflow.

Contient :

* contexte ;
* inputs ;
* prompts ;
* outputs ;
* artefacts ;
* événements ;
* versions ;
* policies ;
* graph state.

⸻

Replay Session

Exécution d’un replay.

⸻

Replay Comparison

Comparaison entre deux runs.

⸻

Replay Provenance

Traçabilité complète d’un artefact.

⸻

5. Arborescence

.asagiri/replays/
  replay-2026-05-27-001/
    replay.yaml
    context/
    prompts/
    outputs/
    graph/
    trust/
    investigations/
    runtime/
    reports/

⸻

6. Commandes CLI

6.1 asa replay create

Créer un replay package.

asa replay create --from-run run-2026-05-27-001

Options :

asa replay create \
  --from-graph graph-001 \
  --from-investigation inv-001 \
  --include-runtime \
  --include-prompts \
  --include-events

⸻

6.2 asa replay run

Rejouer un workflow.

asa replay run replay-2026-05-27-001

Options :

asa replay run replay-2026-05-27-001 \
  --dry-run \
  --compare \
  --strict \
  --offline

⸻

6.3 asa replay compare

Comparer deux runs.

asa replay compare replay-a replay-b

⸻

6.4 asa replay explain

Expliquer les divergences.

asa replay explain replay-a replay-b

⸻

7. Replay Package Format

id: replay-2026-05-27-001
created_at: "2026-05-27T00:00:00+02:00"
source:
  run: run-2026-05-27-001
  graph: graph-001
repo:
  commit: abc123
  branch: feature/onboarding
runtime:
  asagiri_version: 0.8.0
  runtime_mode: guided
agents:
  implementer: cursor
  reviewer: claude
  validator: local
policies:
  strict_trust: true
  max_parallel: 2
artifacts:
  - context-pack.md
  - trust-report.md
  - investigation-report.md

⸻

8. Captured Artifacts

Le replay doit capturer :

execution graph
runtime events
handoffs
trust reports
investigation reports
context packs
agent outputs
prompts
validation outputs
metrics
costs

⸻

9. Runtime Event Replay

Le replay doit pouvoir rejouer :

graph.node.started
agent.completed
verification.completed
trust.low_confidence
review.rejected

Objectif :

* comprendre le déroulé ;
* auditer ;
* debugger.

⸻

10. Deterministic Context

Le replay doit tenter de figer :

* commit Git ;
* policies ;
* configuration ;
* versions runtime ;
* modèles utilisés ;
* prompts ;
* graph state.

⸻

11. Replay Modes

Modes supportés :

full
simulation
offline
audit
compare

⸻

12. Simulation Mode

Le mode simulation ne réexécute pas les agents.

Il rejoue uniquement :

* graph ;
* runtime events ;
* outputs existants ;
* trust reports.

Usage :

asa replay run replay-001 --simulation

⸻

13. Offline Mode

Le mode offline interdit :

* appels cloud ;
* APIs externes ;
* nouveaux prompts distants.

Objectif :

* audit ;
* reproductibilité locale ;
* debugging.

⸻

14. Replay Comparison

Comparer :

* coûts ;
* durée ;
* outputs ;
* trust scores ;
* graph changes ;
* agent decisions ;
* context packs ;
* validations.

Exemple :

Replay Comparison
─────────────────
Replay A cost:        0.18€
Replay B cost:        0.42€
Trust score diff:
- Architecture: +0.11
- Security:     -0.07
Differences:
- replay B inserted investigation node
- replay B added security review
- replay A skipped observability validation

⸻

15. Divergence Detection

Le système doit détecter :

* outputs différents ;
* trust changes ;
* missing validations ;
* graph divergence ;
* runtime divergence ;
* changed dependencies ;
* stale knowledge graph.

⸻

16. Replay Provenance

Chaque artefact doit être traçable.

Exemple :

artifact:
  id: trust-report-001
  produced_by: trust-engine
  graph_node: verify-onboarding
  source_commit: abc123
  replay: replay-001

⸻

17. Integration avec Investigation Engine

Le replay doit pouvoir rejouer :

* investigations ;
* hypothèses ;
* context packs ;
* root cause graphs.

Objectif :

* comparer diagnostics ;
* valider corrections ;
* debugger faux positifs.

⸻

18. Integration avec Trust Engine

Le replay doit permettre :

* comparer trust reports ;
* détecter trust regressions ;
* auditer gates ;
* reproduire validations.

⸻

19. Integration avec Execution Graph Planner

Le replay doit pouvoir restaurer :

* graph state ;
* checkpoints ;
* blocked nodes ;
* rollback decisions ;
* scheduling decisions.

⸻

20. Integration avec Multi-Agent Coordination

Le replay doit restaurer :

* handoffs ;
* agent assignments ;
* retries ;
* escalations ;
* reviews ;
* budgets.

⸻

21. Replay Snapshots

Le système doit pouvoir créer des snapshots.

Exemple :

asa replay snapshot replay-001 --name before-review

⸻

22. Replay Policies

Créer :

replay:
  capture_prompts: true
  capture_runtime_events: true
  capture_agent_outputs: true
  redact_secrets: true
  offline_mode_default: false

⸻

23. Secret Handling

Le replay ne doit jamais stocker :

* secrets bruts ;
* tokens ;
* credentials ;
* .env complets.

Le système doit :

* redacter ;
* anonymiser ;
* signaler données manquantes.

⸻

24. Replay Compression

Le système doit compresser :

* prompts ;
* logs ;
* runtime events ;
* outputs répétitifs.

Objectif :

* éviter replay packages énormes.

⸻

25. Architecture Go

Créer :

internal/replay/
  package.go
  capture.go
  replay.go
  compare.go
  divergence.go
  provenance.go
  compression.go
  snapshots.go
  policies.go

Interfaces :

type ReplayManager interface {
    Create(ctx context.Context, req ReplayCreateRequest) (ReplayPackage, error)
    Run(ctx context.Context, req ReplayRunRequest) (ReplayResult, error)
}
type ReplayComparator interface {
    Compare(ctx context.Context, a string, b string) (ReplayComparison, error)
}

⸻

26. UX terminal cible

Asagiri Replay Engine
═════════════════════
Replay: replay-2026-05-27-001
Artifacts
─────────
✓ execution graph
✓ trust report
✓ investigation report
✓ runtime events
✓ handoffs
Replay mode
───────────
Mode: compare
Offline: true
Comparison
──────────
Cost delta: +0.24€
Trust delta: -0.07 security
Runtime divergence: 2 nodes
Warnings
────────
- replay package missing dashboard metrics
- runtime version mismatch

⸻

27. Tests

Unit tests

* replay package ;
* provenance ;
* divergence detection ;
* comparison ;
* snapshot restore.

Golden tests

testdata/replay/
  basic-run/
  graph-run/
  trust-validation/
  investigation/
  divergence/

Integration tests

asa replay create --from-run run-001
asa replay run replay-001 --offline
asa replay compare replay-a replay-b

⸻

28. Critères d’acceptation

Cette évolution est acceptable si :

* asa replay create génère un replay package ;
* les artefacts principaux sont capturés ;
* asa replay run fonctionne ;
* le mode offline fonctionne ;
* les comparaisons détectent les divergences ;
* les trust reports sont rejouables ;
* les execution graphs sont restaurables ;
* les handoffs agents sont rejouables ;
* les secrets sont redacted ;
* les tests unitaires/golden/intégration passent.

⸻

29. Découpage d’implémentation recommandé

Phase 1 — Replay package

* format YAML ;
* capture artefacts ;
* restore minimal.

Phase 2 — Runtime capture

* runtime events ;
* graph state ;
* checkpoints.

Phase 3 — Comparison engine

* compare reports ;
* compare trust ;
* divergence detection.

Phase 4 — Offline/simulation modes

* offline ;
* simulation ;
* replay-only.

Phase 5 — Integrations

* investigation ;
* trust ;
* execution graph ;
* coordination.

Phase 6 — Provenance & snapshots

* provenance ;
* snapshots ;
* restore.

⸻

30. Risques

Replay packages trop lourds

Mitigation :

* compression ;
* retention ;
* partial replay ;
* snapshots.

Faux déterminisme

Mitigation :

* divergence reporting ;
* confidence ;
* provenance ;
* warnings.

Fuite de secrets

Mitigation :

* redaction ;
* secret scanning ;
* exclusion policies.

Complexité runtime

Mitigation :

* replay optional ;
* incremental rollout ;
* capture policies.

⸻

31. Résumé

Cette évolution ajoute à Asagiri une capacité de replay et d’audit des workflows d’ingénierie.

Avant :

execution happened once

Après :

execution
  ↓
replay package
  ↓
re-execution
  ↓
comparison
  ↓
auditability
  ↓
trust improvement

Principe clé :

Un workflow d’ingénierie piloté par agents doit être rejouable, comparable et explicable. Asagiri doit pouvoir capturer suffisamment de contexte et d’événements pour auditer et reproduire les décisions prises.