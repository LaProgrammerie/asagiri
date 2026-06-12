# Fiabilité & sécurité

**Date :** 2026-05-17 · **Référence :** spec-postv123 §3

## Contrôles existants

| Zone | Mécanisme |
|------|-----------|
| Agents | `exec.Command` sans shell ; dry-run global |
| Investigation | `RunCommand` avec timeout obligatoire |
| MCP | `enabled: false` par défaut ; `denyPath` ; denylist secrets |
| Config | Chemins relatifs validés sous repo |
| Worktrees | Sous `.agentflow/worktrees/` |
| Secrets Notion | `NOTION_TOKEN` via env uniquement |

## Lacunes traitées

| Lacune | Action |
|--------|--------|
| MCP enabled sans limites | `Config.Validate` exige `max_output_bytes` et timeouts > 0 |
| Logs secrets | Package `redact` (patterns token/bearer/api_key) |
| Double scan (DoS local) | Collecte contexte ciblée post-investigation |

## Restant à faire

| Item | Priorité |
|------|----------|
| Appliquer `redact.String` sur stderr agents en cas d’erreur | P1 |
| WAL SQLite + backup doc corruption | P2 |
| Cancellation propagée pipeline → subprocess (context) | P1 |
| Rate limit MCP tools / IP | P3 |
| Tests worktree isolation concurrent | P2 |

## MCP `run_local_check`

Exécute `go test ./...` — puissant ; garder **disabled** par défaut et documenter risque dans config example.

## Score fiabilité partiel

Garde-fous **corrects pour usage local** ; pas encore « production multi-tenant ».
