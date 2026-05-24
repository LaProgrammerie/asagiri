# Produit

## Problème résolu

**AgentFlow** industrialise les workflows de développement assistés par agents : spec → tâches → worktree → implémentation → validation → review → rapport, de façon déterministe et traçable en local.

**Canon produit (évolution) :** [`specv2.md`](../../specv2.md) — intent layer, sources, Notion.  
**Historique V1 :** [`spec.md`](../../spec.md). Ce fichier résume ; ne pas dupliquer la spec complète ici.

## Utilisateurs / contexte

Développeurs et équipes qui utilisent déjà Kiro, Cursor, Codex, Claude Code ou Ollama et veulent un orchestrateur CLI reproductible (pas un agent autonome généraliste).

## État V1 (tranches)

| Tranche | Statut |
|---------|--------|
| `agentflow-init` | Livré |
| `agentflow-spec-7-12` | Livré — config, tâches, RAG, state machine |
| `agentflow-specv2` | Livré — `work`, `continue`, `next`, `inbox`, `sync`, Notion |

## Hors scope (V1)

- UI desktop, dashboard web, orchestration distribuée, multi-user.
- Castor, Yoimachi, déploiement production — sauf ADR future.
- Voir § non-objectifs V1 dans `spec.md`.

## Critères de succès (V1 globale)

Voir objectifs fonctionnels dans `spec.md` §3.1 ; la fondation locale (`init` + état persistant) est en place.
