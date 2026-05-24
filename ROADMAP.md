# Roadmap AgentFlow

Vision : orchestration **déterministe**, **observable**, **local-first** et **cost-aware** des workflows de développement agentique.

## Livré (V1 → V3)

- Primitives CLI : init, spec, plan, dev, verify, review, worktrees, SQLite
- Intent layer (specv2) : `work`, `continue`, `next`, Notion sync
- Cost/perf (specv3) : estimate, investigate, context, cost, MCP, TUI, métriques

## Consolidation (en cours — spec-postv123)

- Audits architecture, sécurité, performance
- Tests packages critiques, explainability CLI
- Préparation open source (LICENSE Apache 2.0, CONTRIBUTING, examples)

## Prochaines étapes (priorisées)

| Priorité | Item |
|----------|------|
| P0 | Couverture `workflow` / `intent/executor` > 50 % |
| P0 | Tokenizers provider-exacts (estimation) |
| P1 | Cache persistant context pack entre runs |
| P1 | CI GitHub Actions (test, lint, release) |
| P2 | TUI bubbletea event-loop interactif |
| P2 | Embeddings RAG vectoriels |
| P3 | Multi-repo / remote state |

Voir [`docs/consolidation/TODO-prioritized.md`](docs/consolidation/TODO-prioritized.md) pour le détail post-audit.
