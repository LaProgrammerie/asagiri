# Produit

## Problème résolu

**Asagiri** industrialise les workflows de développement assistés par agents : spec → tâches → worktree → implémentation → validation → review → rapport, de façon déterministe et traçable en local. La commande CLI officielle est **`asa`**.

**Rebrand (courant) :** [`spec-rename.md`](../archives/specs/spec-rename.md).  
**Canon produit (évolution intent) :** [`specv2.md`](../archives/specs/specv2.md) — intent layer, sources, Notion.  
**Historique V1 :** [`spec.md`](../archives/specs/spec.md). Ce fichier résume ; ne pas dupliquer la spec complète ici.

## Utilisateurs / contexte

Développeurs et équipes qui utilisent déjà Kiro, Cursor, Codex, Claude Code ou Ollama et veulent un orchestrateur CLI reproductible (pas un agent autonome généraliste).

## État (tranches)

| Tranche | Statut |
|---------|--------|
| `agentflow-init` / spec V1 | Livré (chemins migrés vers `.asagiri/` — ADR-016) |
| `agentflow-spec-7-12` | Livré — config, tâches, RAG, state machine |
| `agentflow-specv2` | Livré — `work`, `continue`, `next`, `inbox`, `sync`, Notion |
| `spec-rename` | En cours — Asagiri / `asa`, releases, docs, CI |

## Hors scope (V1)

- UI desktop, dashboard web, orchestration distribuée, multi-user.
- Castor, Yoimachi, déploiement production — sauf ADR future.
- Voir § non-objectifs V1 dans `spec.md`.

## Critères de succès (rebrand)

Voir critères d’acceptation dans `spec-rename.md` ; handoff : `docs/ai/active/handoff.md`.
