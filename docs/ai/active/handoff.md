# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche :** `spec-my-D` (Asagiri Multi-Agent Coordination System) — **en cours**.

## Objectif

Livrer intégralement [`spec-my-D.md`](../../../spec-my-D.md) (§1–22, critères d'acceptation §20) : orchestration explicite des agents spécialisés, handoffs structurés, validation croisée, budgeting et coordination runtime au-dessus de l'execution graph (spec-my-C).

---

## Prérequis

- [`spec-my-A.md`](../../../spec-my-A.md) — **livrée** (`2026-05-27`) — product layer, runtime, analysis, investigation
- [`spec-my-B.md`](../../../spec-my-B.md) — **livrée** (`2026-05-29`) — trust & verification engine, gates, replay
- [`spec-my-C.md`](../../../spec-my-C.md) — **livrée** (`2026-05-29`) — execution graph planner, scheduler, runner, CLI `graph.*` / `plan graph`

**Code prérequis :** `application/internal/executiongraph/`, `internal/trust/`, `internal/runtime/`, `internal/config/` (agents, worktrees), `internal/agent/` selon intégration.

---

## Périmètre autorisé

- `application/internal/coordination/**` (nouveau §18)
- `application/internal/config/**` (bloc `coordination:`)
- `application/internal/runtime/**` (événements `agent.*` §10)
- intégration `executiongraph/`, `trust/`, `worktree/`, `agent/` selon besoin (sans élargir hors spec)
- `.asagiri/handoffs/`, `.asagiri/config.yaml.example`
- `docs/ai/**`, `docs-site/content/docs/{en,fr,de,es}/**`
- `application/internal/cli/**` uniquement si commandes ou flags requis par spec-my-D (sinon hors scope)

---

## Hors scope

- [`spec-phase-finale.md`](../../../spec-phase-finale.md) — **PF-*** (durcissements, stubs, reliquats transverses post A/B/C/D)
- Specs ultérieures (post my-D)
- Commit / push par l'agent
- Refonte du modèle execution graph (spec-my-C livré — extension coordination uniquement)

---

## Lots — Definition of Done

### Lot 1 — Foundation (§3–5, §18)

Package `internal/coordination/` : coordinator, assignment squelette, policies, interfaces `AgentCoordinator` / `AgentAssigner` / `HandoffBuilder` ; types rôles agents (§3–4) ; modes isolation (`shared`, `isolated_worktree`, `readonly`, `sandbox`) ; bloc `coordination:` squelette dans config + example.

**DoD testable :**

- [ ] `go test ./internal/coordination/...` — types rôles, capacités, restrictions, isolation modes
- [ ] Interfaces §18 compilées et mockables en test
- [ ] `coordination:` présent dans `config` + `.asagiri/config.yaml.example` (champs minimaux : `max_parallel_agents`, placeholders assignment)
- [ ] Aucune régression `go test ./internal/config/...`

---

### Lot 2 — Assignment & pipeline (§6–7, §12)

Moteur d'assignation multi-critères (type tâche, risque, coût, contexte, historique, spécialisation, confiance) ; pipelines rejouables par étape ; context reduction (scope, fichiers, flows, contrats, trust, investigation).

**DoD testable :**

- [ ] Tests assignation : mapping config `assignment:` → agent choisi pour scénarios représentatifs
- [ ] Pipeline ordonné (ex. investigator → architect → implementer → reviewer → validator) avec sortie structurée par étape
- [ ] Context reduction : pack minimal généré ; tests sur filtrage fichiers / injection flows-contrats-trust
- [ ] `go test ./internal/coordination/...` vert (assignment + pipeline + context)

---

### Lot 3 — Handoffs & validation (§8–9, §11)

Persistance `.asagiri/handoffs/` ; `HandoffBuilder` ; règles reviewer ≠ implementer, validator ≠ implementer ; security review flows sensibles ; coordination policies (`require_independent_review`, `allow_self_review`, `require_security_review_for`).

**DoD testable :**

- [ ] Handoff YAML/JSON écrit sous `.asagiri/handoffs/<id>/` avec champs spec (from, to, summary, files, constraints, confidence)
- [ ] Tests : reviewer et implementer ne peuvent pas être le même agent quand `allow_self_review: false`
- [ ] Tests : flow `auth` / `payments` déclenche security review si configuré
- [ ] Policies chargées depuis `coordination:` et appliquées par le coordinateur

---

### Lot 4 — Budget, retry, conflict, merge (§13–16)

Suivi coût / tokens / durée / retries / total pipeline ; retry & escalation configurables ; détection conflits (fichiers, contrats, flows, trust downgrade) ; merge policies (`require`, `block_if`).

**DoD testable :**

- [ ] Budget agrégé par agent et total pipeline (tests unitaires sur cumul)
- [ ] Retry : respect `max_attempts` ; escalation vers investigator / architecture_review selon config
- [ ] Conflict detection : au moins un test par catégorie (fichier concurrent, divergence contrat/flow, trust downgrade)
- [ ] Merge bloqué si `unresolved_conflicts` ou `low_security_confidence` ; autorisé si trust + review + validation passés

---

### Lot 5 — Runtime, coordination graph, worktrees (§5, §10, §17)

Événements `agent.*` ; graphe runtime coordination (agent ↔ task ↔ flow ↔ branch ↔ review ↔ trust ↔ investigation) ; intégration execution graph runner + trust gates ; isolation `isolated_worktree` opérationnelle.

**DoD testable :**

- [ ] Émission runtime : `agent.started`, `agent.completed`, `agent.failed`, `agent.blocked`, `agent.review_requested`, `agent.review_rejected`, `agent.handoff.created`, `agent.context_reduced`
- [ ] Graphe coordination persisté ou dérivable pour inspection (liens agent/task/flow/trust)
- [ ] Runner execution graph invoque coordinateur ; trust gates peuvent bloquer avant merge
- [ ] Mode `isolated_worktree` : test d'intégration ou fixture prouvant exécution hors workspace principal
- [ ] `go test ./internal/runtime/... ./internal/coordination/...` vert sur chemins touchés

---

### Lot 6 — UX terminal, acceptance, documentation (§19–20)

Affichage terminal pipeline / agents / coûts / warnings ; critères §20 ; canon `docs/ai` ; site EN/FR/DE/ES ; génération CLI docs.

**DoD testable :**

- [ ] Sortie terminal conforme au gabarit §19 (pipeline, agents actifs, coûts, warnings) — test snapshot ou golden stdout
- [ ] Tous les critères §20 vérifiables (checklist DoD global ci-dessous)
- [ ] `06-spec-my-d.md` (ou équivalent canon), `02-architecture.md`, ADR si décision durable, `context-map.md` à jour
- [ ] Pages site EN/FR/DE/ES pour concepts/config/commandes touchés par coordination
- [ ] `make build && ./bin/asa docs generate-cli`

---

## Matrice traçabilité

| ID | Exigence | Lot | Statut |
|----|----------|-----|--------|
| D-ROL-1 | Rôles explicites agents §3–4 | 1 | [ ] |
| D-ISO-1 | Isolation worktree §5 | 1, 5 | [ ] |
| D-ASG-1 | Agent assignment §6 | 2 | [ ] |
| D-PIP-1 | Agent pipeline §7 | 2 | [ ] |
| D-HOF-1 | Handoffs persistés §8 | 3 | [ ] |
| D-XVAL-1 | Cross-validation §9 | 3 | [ ] |
| D-EVT-1 | Runtime events `agent.*` §10 | 5 | [ ] |
| D-POL-1 | Coordination policies §11 | 3 | [ ] |
| D-CTX-1 | Context reduction §12 | 2 | [ ] |
| D-BUD-1 | Budgeting §13 | 4 | [ ] |
| D-RET-1 | Retry & escalation §14 | 4 | [ ] |
| D-CON-1 | Conflict detection §15 | 4 | [ ] |
| D-MRG-1 | Merge policies §16 | 4 | [ ] |
| D-RGR-1 | Runtime coordination graph §17 | 5 | [ ] |
| D-ARCH-1 | Packages `coordination/` §18 | 1 | [ ] |
| D-UX-1 | UX terminal §19 | 6 | [ ] |
| D-ACC-1 | Critères acceptation §20 | all | [ ] |
| D-DOC-1 | Doc canon + site 4 locales | 6 | [ ] |

---

## DoD global spec-my-D (§20)

- [ ] Les agents ont des rôles explicites
- [ ] Les handoffs sont persistés sous `.asagiri/handoffs/`
- [ ] Les reviewers sont indépendants de l'implementer
- [ ] Les conflits sont détectés (fichiers, contrats, flows, trust)
- [ ] Les budgets sont suivis (par agent et pipeline)
- [ ] Les retries et escalations fonctionnent selon config
- [ ] Les événements runtime `agent.*` sont émis
- [ ] Les worktrees isolés fonctionnent (`isolated_worktree`)
- [ ] Les pipelines sont rejouables
- [ ] Les validations trust peuvent bloquer avant merge
- [ ] Tests unitaires / intégration critiques passent

---

## Validation globale (cible)

```bash
cd application && go test ./internal/coordination/... ./internal/config/... ./internal/runtime/...
cd application && go test ./internal/executiongraph/... -count=1
make build && ./bin/asa docs generate-cli
# Smoke coordination (à affiner quand CLI dédiée existe)
./bin/asa graph run <product> --flow <flow> --dry-run --ci --json
# Vérifier artefacts handoffs après run coordonné
ls .asagiri/handoffs/
```

---

## Références

- [`spec-my-D.md`](../../../spec-my-D.md)
- [`06-spec-my-c.md`](../06-spec-my-c.md) (prérequis livré)
- [`06-spec-my-b.md`](../06-spec-my-b.md) (prérequis livré)
- [`06-spec-my-a.md`](../06-spec-my-a.md) (prérequis livré)
- ADR-022 — [`docs/decisions/022-execution-graph.md`](../../decisions/022-execution-graph.md)
- [`spec-phase-finale.md`](../../../spec-phase-finale.md) — hors scope implémentation directe
