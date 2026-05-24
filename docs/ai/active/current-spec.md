# Current spec — AgentFlow specv2

**Phase :** `agentflow-specv2` (Intent Layer & sources externes)  
**Date :** 2026-05-17

## Résumé

Couche haut niveau au-dessus des primitives V1 : intentions en langage naturel, plan d’exécution inspectable, sync Notion → `.agentflow/specs/`, inbox multi-sources.

## Critères de phase

| Domaine | Statut |
|---------|--------|
| §4 Commandes work/continue/next/inbox/sync | Livré |
| §5 IntentResolver | Livré |
| §6 HighLevelPlanner | Livré |
| §7–8 Sources Local + Notion | Livré (Notion database si `specs_database_id`) |
| §9 Config intent/work/sources | Livré |
| §10–13 Modes, UX, sécurité | Livré |
| §14 Critères d’acceptation | Livré + tests |

## Spec canonique

- **Évolution / produit :** [`specv2.md`](../../../specv2.md)
- **Historique V1 :** [`spec.md`](../../../spec.md)

## Handoff actif

[`handoff.md`](handoff.md)
