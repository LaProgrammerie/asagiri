# Plan de rationalisation

Objectif : **simplicité robuste** sans big-bang.

## 1. Pipeline contexte (fait)

Une seule passe fichiers pertinents : investigation → `CollectForPipeline`.

## 2. Routing unifié (proposé)

Aujourd’hui : routing à l’estimation seulement.  
**Cible :** `intent.Executor.runStep` appelle `routing.Route` pour choisir agent/model avant primitive.

## 3. Cache pack (proposé)

```
.agentflow/cache/context/<feature>/<task>/<git-rev>.json
```

Invalidation : `git rev-parse HEAD`, TTL 24h.

## 4. Template vs produit (proposé)

- Branche ou repo **agentflow** Go-only
- Ce dépôt reste **ai-engineering-template** avec AgentFlow embarqué

## 5. Suppressions candidates (non fait — à valider)

| Élément | Raison |
|---------|--------|
| Stub MCP `get_run_status` | Implémenter ou retirer de tools/list |
| Doublon scan dans `Collect()` fallback | OK si candidats vides |

## 6. Ne pas faire

- Fusionner intent + workflow en un seul package
- Hardcoder prix modèles (déjà interdit ADR-010)
