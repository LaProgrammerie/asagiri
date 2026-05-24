# Audit architecture & drift

**Date :** 2026-05-17 · **Référence :** spec-postv123 §1

## Cartographie actuelle

| Couche | Packages | Rôle |
|--------|----------|------|
| CLI | `internal/cli` | Cobra, flags V1/V2/V3 |
| Intent | `internal/intent` | Resolver, planner, executor → primitives |
| Workflow | `internal/workflow` | State machine, runs SQLite, agents |
| V3 pipeline | `internal/pipeline`, `cost`, `contextopt`, `investigation`, `routing`, `telemetry`, `tui` | Cost-aware preprocessing |
| État | `internal/store/sqlite` | Runs, tasks, métriques |
| Contrat public | `pkg/agentflow` | Types tâches canoniques |

Alignement specs : **bon** (V1 primitives, V2 intent, V3 cost/perf coexistent sans fork CLI).

## Écarts identifiés

| Sévérité | Écart | Risque |
|----------|-------|--------|
| **Critique (corrigé)** | Double scan repo : `investigation.Run` + `contextopt.Collect` (arbre complet) | Latence, I/O inutile sur gros dépôts |
| Moyen | `routing.Route` peu utilisé hors estimation | Décisions tier incohérentes vs config |
| Moyen | `telemetry` sans tests d’intégration SQLite (cycle d’import) | Régression métriques non détectée |
| Faible | RAG + template PHP legacy (`application/public`, Cypress) dans le même module | Bruit pour adopteurs OSS Go-only |
| Faible | MCP `get_run_status` stub | Outils MCP incomplets |

## Dette / couplage

- **Intent → workflow** : correct (façade haut niveau).
- **Pipeline → intent.Executor** : acceptable ; pas de cycle.
- **Duplication** : scoring contexte + grep partiellement redondants avec investigation (acceptable si borné).

## Corrections appliquées (code)

- `contextopt.CollectForPipeline` : collecte ciblée sur candidats investigation + chemins spec.
- `pipeline.RunV3Pipeline` : utilise la collecte ciblée.
- `cost.BuildEstimate` : enrichit `Reason` avec `routing.Route`.
- Package `internal/redact` pour masquage logs.

## Décisions proposées (ADR)

- **ADR-011** : consolidation post-V3 ; spec active `spec-postv123.md`.
- Reporter nettoyage template PHP dans un dépôt fork ou branche `template-only`.

## Recommandations

1. Extraire progressivement le squelette PHP/E2E du module Go public (ou documenter clairement « template full-stack »).
2. Unifier routing dans executor (choix agent par step class).
3. Persister cache context pack (hash feature+task) — voir rationalization-plan.
