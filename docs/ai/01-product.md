# Produit

## Problème résolu

**AgentFlow** industrialise les workflows de développement assistés par agents : spec → tâches → worktree → implémentation → validation → review → rapport, de façon déterministe et traçable en local.

**Canon produit détaillé :** [`spec.md`](../../spec.md) (vision, CLI, config, workflow cible). Ce fichier résume ; ne pas dupliquer la spec complète ici.

## Utilisateurs / contexte

Développeurs et équipes qui utilisent déjà Kiro, Cursor, Codex, Claude Code ou Ollama et veulent un orchestrateur CLI reproductible (pas un agent autonome généraliste).

## État V1 (tranches)

| Tranche | Statut |
|---------|--------|
| `agentflow-init` | Livré — `init`, `doctor`, config YAML, SQLite schéma V1 |
| `plan` + lecture spec Kiro | Prochaine recommandée |
| Agents, worktrees, workflow complet | À venir (voir `spec.md` §3) |

## Hors scope (V1)

- UI desktop, dashboard web, orchestration distribuée, multi-user.
- Castor, Yoimachi, déploiement production — sauf ADR future.
- Voir § non-objectifs V1 dans `spec.md`.

## Critères de succès (V1 globale)

Voir objectifs fonctionnels dans `spec.md` §3.1 ; la fondation locale (`init` + état persistant) est en place.
