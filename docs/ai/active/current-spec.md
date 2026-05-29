# Current spec — specv3 livrée

**Phase :** [`specv3.md`](../../../specv3.md) — **livrée** (`2026-05-29`)  
**Handoff :** [`handoff.md`](handoff.md) — tranche specv3, matrice 100 %

## Spec active

- **Registre :** [`specv3.md`](../../../specv3.md) — Cost, Performance & Token Optimization
- **Canon :** [`06-spec-v3.md`](../06-spec-v3.md)
- **Handoff :** [`handoff.md`](handoff.md)

## Résumé — livrables V3

| Bloc | Contenu |
|------|---------|
| **Pipeline** | `RunV3PreFlight` → affichage → `RunV3Execute` |
| **CLI** | `estimate`, `work --estimate-only`, `context`, `cost`, `investigate`, `mcp serve` |
| **Métriques** | SQLite `run_metrics` / `step_metrics`, `cost report` |
| **Contexte** | reduce/pack, savings mesurés, `--no-context-reduction` |
| **TUI** | rich / plain / json (`ui.mode`) |
| **MCP** | tools `asagiri.*`, désactivé par défaut |
| **Rapport** | section Cost & Performance dans `report.md` |

## Stack A–F + PF

Toutes livrées — voir [`context-map.md`](../context-map.md).

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.
