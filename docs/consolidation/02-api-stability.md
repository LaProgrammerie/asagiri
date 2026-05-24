# Stabilisation API & primitives

**Date :** 2026-05-17 · **Référence :** spec-postv123 §2

## CLI — cohérence

| Famille | Commandes | Statut |
|---------|-----------|--------|
| V1 | init, doctor, spec, plan, enrich, dev, verify, review, status, resume, clean, index | Stable |
| V2 | work, continue, next, inbox, sync | Stable |
| V3 | estimate, investigate, context, cost, inspect, mcp | Stable |

Flags `work` V3 documentés dans README ; pas de breaking change observé sur intégration CLI.

## Séparation des responsabilités

| Domaine | Package | Import depuis CLI |
|---------|---------|-------------------|
| Moteur workflow | `workflow` | via executor / commandes V1 |
| Agents | `agent`, `agent/exec` | workflow |
| Intent | `intent` | cli work/continue/next |
| Cost | `cost` | pipeline, cli estimate |
| Investigation | `investigation` | pipeline, mcp |
| TUI | `tui` | cli display uniquement |

**Règle respectée :** `pkg/asagiri` = types stables ; `internal/*` non importable hors module.

## `pkg/asagiri`

- Types tâches / statuts alignés spec §8.
- Pas d’API runtime exportée (volontaire).

## Incohérences mineures (non bloquantes)

- `context.go` renommé en `app_context.go` / `context_cmd.go` (clarification) — OK.
- Nom module Go `asagiri` vs branding **Asagiri** — documenté README ; renommage module = breaking (phase 2, ADR-016).

## Actions

- Aucun rename bloquant requis avant publication.
- Documenter mapping agent config → model profile dans `03-standards.md` (suivi).
