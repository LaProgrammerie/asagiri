# Current spec — AgentFlow specv3

**Phase :** `agentflow-specv3` (Cost, Performance & Token Optimization)  
**Date :** 2026-05-17

## Résumé

Couche cost/performance au-dessus de V1 + V2 : estimation locale tokens/coût/temps, investigation repo, optimisation contexte, budgets, métriques SQLite, TUI (rich/plain/json), MCP local.

## Critères de phase

| Domaine | Statut |
|---------|--------|
| §4–7 Estimation / pricing / budget / durée | Livré |
| §8–9 Context opt + investigation | Livré |
| §10 MCP local (stdio, tools §10.2) | Livré (`mcp.enabled: false` par défaut) |
| §11 Routing cost-aware | Livré (heuristique) |
| §12 Extension `work` + flags V3 | Livré |
| §13 TUI (lipgloss + fallback) | Livré |
| §14–15 Métriques + rapport Cost & Performance | Livré |
| §16 Commandes estimate/investigate/context/cost/inspect/mcp | Livré |
| §17 Critères d'acceptation | Livré + tests |

## Specs

- **Courante :** [`specv3.md`](../../../specv3.md)
- **Intent layer :** [`specv2.md`](../../../specv2.md)
- **V1 :** [`spec.md`](../../../spec.md)

## Handoff actif

[`handoff.md`](handoff.md)
