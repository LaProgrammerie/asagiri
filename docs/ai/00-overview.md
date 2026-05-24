# Vue d’ensemble du contexte

## Rôle du dépôt

**Asagiri** — orchestrateur CLI local (Go) pour workflows de développement agentique : spec → tâches → worktrees → agents → validation → review, de façon déterministe et traçable. Ce dépôt contient le binaire **`asa`**, la couche AI Engineering (`docs/ai/`, `.kiro/`, `.cursor/`) et le runtime local optionnel (Docker Compose).

Détail produit : [`spec.md`](../../spec.md) à la racine (historique V1 ; branding courant : [`spec-rename.md`](../../spec-rename.md)).

## Carte du système de contexte

Pour les **sources de vérité**, Kiro vs Cursor vs Copilot, et specs vs `docs/ai/active/`, lire **`context-map.md`** en premier.

## Chemins importants

| Chemin | Rôle |
|--------|------|
| `spec-rename.md` | Spec active — rebranding Asagiri / `asa` |
| `spec.md` | Spec produit & technique V1 (historique) |
| `AGENTS.md` | Routeur court (toujours inclus par Kiro) |
| `docs/ai/context-map.md` | Carte des fichiers et anti-dérive |
| `docs/ai/` | Canon projet (stack, standards, ADR) |
| `docs/ai/active/` | Résumé de spec + handoff d’exécution |
| `.asagiri/` | État local Asagiri (config, SQLite, runs, logs, worktrees) |
| `.asagiri/config.yaml.example` | Schéma de configuration versionnable |
| `bin/asa` | CLI compilée (`make build`) |
| `.kiro/specs/` | Artefacts Kiro (requirements, design, tasks) |
| `application/` | Code Go (`cmd/asa`, `internal/`) |
| `Makefile` | Entrées nommées pour dev, test, Docker |
| `infrastructure/docker/` | Compose et images locales |
| `docs/migration/github-rename-asagiri.md` | Runbook phase 2 (repo GitHub, module Go) |

## Liens rapides

- Produit : `01-product.md` → `spec-rename.md`
- Architecture : `02-architecture.md`
- Spec / handoff : `active/current-spec.md`, `active/handoff.md`
