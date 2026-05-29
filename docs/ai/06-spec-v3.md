# Spec V3 — Cost, Performance & Token Optimization (Asagiri)

**Source :** [`specv3.md`](../../specv3.md) (racine) — branding **Asagiri** / CLI **`asa`** (pas `agentflow`).

## Objectif

Couche cost/performance-aware : investigation locale → optimisation contexte → estimation → budgets → exécution → métriques → rapport.

## Packages

| Package | Rôle |
|---------|------|
| `internal/cost/` | Tokens, pricing config, budgets, `ExecutionEstimate`, durée |
| `internal/contextopt/` | Collect, reduce, pack, `ComputeOptimize` |
| `internal/investigation/` | Grep, candidats, tests liés, symboles |
| `internal/routing/` | Routage local / cloud (`cost_aware`) |
| `internal/telemetry/` | `run_metrics`, `step_metrics`, rapports |
| `internal/tui/` | Rich / plain / JSON (isolé du moteur) |
| `internal/mcp/` | MCP stdio (`asagiri.*` tools), désactivé par défaut |
| `internal/pipeline/` | `RunV3PreFlight`, `RunV3Execute`, `RunV3Pipeline` |

## CLI

| Commande | Spec |
|----------|------|
| `asa work "<…>"` | Pipeline V3 ; `--estimate-only` ; estimation **avant** exécution |
| `asa estimate <feature>` | Estimation sans cloud |
| `asa investigate "<symptom>"` | Investigation structurée (`--task`) |
| `asa context <feature> --show\|--optimize` | Pack + savings tokens |
| `asa cost report\|models` | Historique SQLite |
| `asa mcp serve` | MCP local (config `mcp.enabled`) |

## Config (`.asagiri/config.yaml`)

Blocs : `pricing`, `budgets`, `token_estimation`, `routing`, `ui`, `mcp`.

## Critères d'acceptation (§17)

- [x] Estimation avant `work` (pre-flight + affichage)
- [x] `estimate` sans agent cloud
- [x] Investigation locale avant modèle (pipeline)
- [x] Contexte mesuré et réductible (`--no-context-reduction`)
- [x] Budgets blocage / confirmation
- [x] Métriques SQLite + `cost report`
- [x] Rapport Cost & Performance (`report.CostPerformance`)
- [x] TUI rich + fallback plain/json
- [x] MCP search/read/estimate/context (scope + désactivé par défaut)

## ADR

ADR-010 — `docs/ai/05-decisions.md`.

## Validation

```bash
cd application && go test ./... -count=1
make build && ./bin/asa docs generate-cli
./bin/asa estimate agentflow-test --dry-run
./bin/asa work "develop agentflow-test" --dry-run --estimate-only
./bin/asa context agentflow-test --optimize --dry-run
./bin/asa cost report --since 7d
```
