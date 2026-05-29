# Spec-my-C — Execution Graph Planner (canon `docs/ai`)

**Statut :** livré (`2026-05-29`)  
**Spec racine :** [`spec-my-C.md`](../archives/specs/spec-my-C.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)  
**Prérequis :** [`06-spec-my-a.md`](06-spec-my-a.md), [`06-spec-my-b.md`](06-spec-my-b.md)

---

## 1. Résumé

Spec-my-C remplace l’exécution linéaire de tâches par une **planification orchestrée par graphe** :

```text
intent / flow / spec → execution graph → dependency-aware planning → agent assignment
  → parallelizable execution → checkpoints → validation gates → rollback / recovery → verified delivery
```

Le planner reste **local-first**, inspectable et déterministe : pas d’orchestrateur cloud, pas de queue distante obligatoire.

---

## 2. Arborescence `.asagiri/graphs/`

```text
.asagiri/graphs/<graph-id>/
  execution-graph.yaml
  execution-graph.json
  plan.md
  metrics.json
  timeline.jsonl
  events.jsonl
  report.md          # après exécution
```

- `<graph-id>` : identifiant stable (ex. `graph-2026-05-29-9307e4be`)
- Config : bloc `execution_graph:` dans `.asagiri/config.yaml` (§24)

---

## 3. Packages Go

| Package | Rôle |
|---------|------|
| `internal/executiongraph/` | Modèle, planner, scheduler, runner, repository, renderer, state machine |
| `internal/executiongraph/` (`dependency.go`) | Inférence dépendances tasks/flows/contracts/fichiers, cycle detection |
| `internal/executiongraph/` (`scheduler.go`) | Topological sort, parallel groups, conflits fichiers |
| `internal/executiongraph/` (`estimator.go`, `risk.go`) | Coût, durée, risque, blast radius |
| `internal/executiongraph/` (`checkpoints.go`, `rollback.go`) | Checkpoints et stratégies rollback |
| `internal/executiongraph/` (`trust_enrichment.go`, `investigation_enrichment.go`) | Gates trust, nœuds investigation auto |
| `internal/config/` | `ExecutionGraphConfig`, gates et rollback |
| `internal/runtime/` | Émission événements `graph.*` (§19) via `GraphEmitter` |
| `internal/cli/` | `plan graph`, `plan explain`, `graph run|status|resume|visualize` |

**Interfaces (§21)** : `ExecutionGraphPlanner`, `DependencyInferer`, `GraphScheduler`, `GraphExecutor` dans `interfaces.go`.

---

## 4. Modèle graphe (§6–9, §20)

**Nœuds** : `investigation`, `enrichment`, `architecture_derivation`, `contract_generation`, `implementation`, `validation`, `review`, `trust_verification`, `documentation`, `release_check`, `manual_approval`, `rollback`.

**Arêtes** : `requires`, `blocks`, `validates`, `produces_context_for`, `must_run_after`, `can_run_after`, `parallel_with`, `rollback_depends_on`, `requires_human_approval`.

**États graphe** : `planned`, `ready`, `running`, `blocked`, `paused`, `failed`, `completed`, `aborted`, `rolled_back`.

**États nœud** : `pending`, `ready`, `running`, `succeeded`, `failed`, `skipped`, `blocked`, `rolled_back`.

Persistance : `execution-graph.yaml` + `execution-graph.json` ; validation à l’écriture.

---

## 5. Planification (§10–16)

| Concern | Comportement |
|---------|--------------|
| **Dependency inference** | Tasks, flows, contracts, fichiers touchés, permissions ; cycles → erreur |
| **Parallelization** | Scopes fichiers disjoints ; `max_parallel` défaut 2 ; jamais deux tâches sur le même fichier |
| **Agent assignment** | Par type nœud, risque, budget, config agents |
| **Cost-aware** | Estimation par nœud et globale (`GraphEstimate`) |
| **Risk-aware** | Niveau, blast radius, `required_checks`, approbation humaine |
| **Checkpoints** | Après nœuds clés ; état Git/worktree, outputs, coût/temps |
| **Rollback** | `worktree_reset`, `patch_revert`, `migration_down`, `feature_flag_disable`, `manual` |

---

## 6. Intégrations (§17–19)

**Trust (§17)** : gates sur contract public, flow critique, sécurité, observability ; `--strict-trust` sur `graph run`.

**Investigation (§18)** : insertion auto de nœuds investigation (ambiguïté, échec précédent, risque élevé).

**Runtime (§19)** — événements : `graph.created`, `graph.started`, `graph.node.started`, `graph.node.completed`, `graph.node.failed`, `graph.checkpoint.created`, `graph.blocked`, `graph.completed`.

---

## 7. Commandes CLI (§5, §22–25)

| Commande | Rôle |
|----------|------|
| `asa plan graph <product> --flow <id>` | Génère graphe sans exécuter ; `--from-product`, `--from-spec`, `--include-reviews`, `--include-docs`, `--estimate`, `--output markdown`, `--json`, `--ci` |
| `asa plan explain <product> --flow <id>` | Explication dépendances, parallélisme, risques |
| `asa graph run <product> --flow <id>` | Plan + exécution ; `--max-parallel`, `--stop-on-risk`, `--strict-trust`, `--budget`, `--dry-run`, `--ci`, `--json` |
| `asa graph status <graph-id>` | État graphe persisté ; `--json` |
| `asa graph resume <graph-id>` | Reprise depuis checkpoint ; `--json` |
| `asa graph visualize <graph-id> --format mermaid\|json\|dot\|markdown` | Export graphe |

CI : `asa plan graph … --ci --json` et `asa graph run … --ci --max-parallel 1` → exit non-zéro si politique violée.

---

## 8. Configuration (§24)

```yaml
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
```

---

## 9. Tests (§26)

- **Unit** : parsing, inference, cycles, scheduling, risk, rollback
- **Golden** : `application/internal/executiongraph/testdata/execution-graph/` (5 scénarios)
- **Integration CLI** : `graph_integration_test.go` (dry-run, mermaid, resume, `--ci --json`)

---

## 10. Documentation publique (site)

Pages par locale `docs-site/content/docs/{en,fr,de,es}/` :

| Sujet | Chemins |
|-------|---------|
| Concept | `concepts/execution-graph` |
| CLI | `cli/plan-graph`, `plan-explain`, `graph-run`, `graph-status`, `graph-resume`, `graph-visualize` |
| Config | section `execution_graph` dans `configuration/config-file` |
| Généré | `en/cli/generated/plan`, `graph` |

---

## 11. Décisions

- **ADR-022** — Execution graph local-first (`.asagiri/graphs/`, planning déterministe, parallélisme conservateur)

---

## 12. Validation

```bash
cd application && go test ./internal/executiongraph/... ./internal/cli/...
make build && ./bin/asa docs generate-cli
./bin/asa plan graph workspace-saas --flow onboarding --json
./bin/asa plan explain workspace-saas --flow onboarding
./bin/asa graph run workspace-saas --flow onboarding --dry-run --ci --json
./bin/asa graph status graph-<id>
./bin/asa graph resume graph-<id>
./bin/asa graph visualize graph-<id> --format mermaid
```

Traçabilité handoff : matrice C-* dans [`active/handoff.md`](active/handoff.md).

**Durcissement phase finale** : **PF-C-01…06** livrés (`2026-05-29`) — voir [`spec-phase-finale.md`](../archives/specs/spec-phase-finale.md) §5.
