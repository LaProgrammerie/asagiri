# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `agentflow-specv3` : livrée** (`2026-05-17`).

## Immediate objective

Couche cost/performance AgentFlow (specv3) : estimation, budgets, investigation locale, context optimization, métriques, TUI, MCP — sans régression V1/V2.

## Allowed scope (agentflow-specv3)

- `application/internal/cost/`, `contextopt/`, `investigation/`, `telemetry/`, `tui/`, `routing/`, `pipeline/`, `mcp/`
- `application/internal/cli/` (estimate, investigate, context, cost, inspect, mcp, work flags)
- `application/internal/config/`, `store/sqlite/` (migration 003)
- `application/internal/report/` (section Cost & Performance)
- `.agentflow/config.yaml.example`
- `docs/ai/` (handoff, current-spec, 02–05, context-map)
- `README.md`

## Definition of Done — agentflow-specv3

- [x] Modules §4 : cost, contextopt, investigation, telemetry, tui
- [x] Config models/budgets/pricing/token_estimation/routing/ui/mcp
- [x] Commandes §16 + extension `work` §12
- [x] Pipeline §12.1 via `pipeline.RunV3Pipeline`
- [x] Migration `003_run_step_metrics.sql` + persistance métriques
- [x] MCP tools §10.2 (stdio, désactivé par défaut)
- [x] Rapport §15 Cost & Performance
- [x] `go test -race ./...` vert
- [x] V1/V2 commandes inchangées (régression CLI intégration)

## Hors scope

- Tokenizers provider-exacts (V3.2)
- Bubbletea event-loop interactif complet (progress plain/rich suffisant)
- Embeddings RAG vectoriels

## References

- [`specv3.md`](../../../specv3.md)
- [`specv2.md`](../../../specv2.md)
- [`current-spec.md`](current-spec.md)
- ADR-010 dans [`05-decisions.md`](../05-decisions.md)
