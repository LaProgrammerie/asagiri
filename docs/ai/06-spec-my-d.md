# Spec-my-D — Multi-Agent Coordination (canon `docs/ai`)

**Statut :** livré — fondation + UX terminal + **D-FULL** (worktrees git réels, exécution nœuds via `NodeExecutor`, scoring avec historique) ; budget/conflict/escalation : trackers mémoire (lot 4)  
**Spec racine :** [`spec-my-D.md`](../../spec-my-D.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)  
**Prérequis :** [`06-spec-my-c.md`](06-spec-my-c.md) (execution graph)

---

## 1. Résumé

Spec-my-D ajoute une couche **local-first** de gouvernance multi-agents au-dessus du graphe d'exécution :

```text
execution graph → agent assignment (rôles, isolation) → handoffs structurés
  → cross-validation → trust gates → merge / reject
```

Objectifs : rôles explicites, reviewers indépendants, handoffs persistés, budgets et conflits traçables, événements `agent.*` inspectables.

---

## 2. Arborescence `.asagiri/handoffs/`

```text
.asagiri/handoffs/<handoff-id>/
  handoff.yaml
```

Champs : `from`, `to`, `summary`, `files`, `constraints`, `confidence`, `created_at`.

Config : bloc `coordination:` dans `.asagiri/config.yaml` (§11).

---

## 3. Packages Go

| Package | Rôle |
|---------|------|
| `internal/coordination/` | Coordinator, assigner, policies, handoffs, runtime emitter, display |
| `internal/coordination/` (`assignment.go`) | Rôles agents, isolation, `DefaultAssigner` + overrides `coordination.assignment` |
| `internal/coordination/` (`handoff.go`) | `DefaultHandoffBuilder` → `.asagiri/handoffs/` |
| `internal/coordination/` (`policies.go`) | `require_independent_review`, `allow_self_review`, security flows |
| `internal/coordination/` (`display.go`) | `FormatMultiAgentRuntime` — gabarit terminal §19 |
| `internal/coordination/` (`isolation.go`) | `EnsureWorktree` — `git worktree add/remove` sous `.asagiri/worktrees/<graph>/<node>/` |
| `internal/coordination/` (`agent_runner.go`) | `NodeExecutor`, `MarkerNodeAgentRunner`, `RunOptions` pour le runner |
| `internal/coordination/` (`assignment_history.go`) | Historique succès pour enrichissement `ScoringAssigner` (§6) |
| `internal/coordination/` (`budgeting.go`, `conflict.go`, `escalation.go`) | Trackers / détecteurs lot 4 (mémoire) |
| `internal/executiongraph/` (`node_executor.go`, `runner.go`) | Hook `NodeExecutor` + champs `DefaultRunner.Coordinator` / `NodeExecutor` |
| `internal/config/` | `CoordinationConfig`, validation `profiles.*.agent` ∈ `agents` |
| `internal/runtime/` | Types événements `agent.*` (§10) |

**Interfaces (§18)** : `AgentCoordinator`, `AgentAssigner`, `HandoffBuilder` dans `coordinator.go` / `assignment.go` / `handoff.go`.

---

## 4. Rôles et isolation (§3–5)

**Rôles :** investigator, architect, planner, implementer, reviewer, validator, security_auditor, documenter, …

**Isolation :** `shared`, `isolated_worktree` (défaut), `readonly`, `sandbox`.

`EnsureWorktree(ctx, repoRoot, graphID, nodeID, branch)` crée ou réutilise un worktree git ; `cleanup` appelle `git worktree remove --force`.

`RoleForNodeType` mappe les types de nœuds execution graph vers les rôles coordination.

---

## 5. Configuration `coordination:`

Voir [`.asagiri/config.yaml.example`](../../.asagiri/config.yaml.example) et section site [configuration/config-file](/docs/configuration/config-file#coordination).

Champs clés : `max_parallel_agents`, `default_isolation`, policies review, `assignment:` (overrides par type de tâche), `profiles:` (rôle + agent + capabilities), `pipeline:` (ordre des rôles).

---

## 6. Runtime et UX (§10, §19)

**Événements** (`CoordinationEmitter`) : `agent.started`, `agent.completed`, `agent.failed`, `agent.blocked`, `agent.review_requested`, `agent.review_rejected`, `agent.handoff.created`, `agent.context_reduced`.

**Terminal** : `FormatMultiAgentRuntime(MultiAgentRuntimeView)` affiche pipeline (✓ / ⠋ / ○), agents actifs, coûts, warnings — golden `testdata/multi_agent_runtime.txt`.

---

## 7. Tests

```bash
cd application && go test ./internal/coordination/... -count=1
UPDATE_GOLDEN=1 go test ./internal/coordination/... -run FormatMultiAgentRuntime -count=1  # regen golden
```

---

## 8. Documentation publique (site)

Pages par locale `docs-site/content/docs/{en,fr,de,es}/` :

| Sujet | Chemins |
|-------|---------|
| Concept | `concepts/multi-agent-coordination` |
| Config | section `coordination` dans `configuration/config-file` |

---

## 9. Décisions

- **ADR-023** — Fondation coordination (`internal/coordination/`, `coordination:`, handoffs, `agent.*`)

---

## 10. Validation

```bash
cd application && go test ./internal/coordination/... ./internal/config/...
make build && ./bin/asa docs generate-cli
```

Traçabilité handoff : matrice D-* dans [`active/handoff.md`](active/handoff.md).
