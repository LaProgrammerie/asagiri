# Explainability & trust

**Date :** 2026-05-17 · **Référence :** spec-postv123 §9

## Implémentation CLI

### `estimate` / `work --estimate-only`

Boîte **Estimated execution** (`cli/v3_display.go`) :

- Tokens contexte in/out
- Économies contexte (si réduction)
- Coût, durée, budget, confiance
- **Steps (why model / tier)** : pour chaque step planifié — agent, model, tier local/cloud, tokens, raison (`Reason` + `routing=`)

### `work` fin d’exécution

- Rapport intent existant (`intent.PrintWorkReport`)
- **Résumé** : instruction, feature/task, estimation agrégée, étapes exécutées, dernier run ID

### Routing

`cost.BuildEstimate` enrichit les steps cloud avec `routing.Route` → raisons : `prefer_local`, `no_cloud`, `cloud_heavy`, `cloud_fast`, `default`.

## Ce que l’utilisateur doit encore inférer

- Détail fichiers exclus du pack (`--show-context-plan` partiel)
- Coût **réel** post-run (vs estimé) — métriques SQLite présentes, affichage CLI limité

## Pistes

- `inspect` + `cost report` dans le résumé `work`
- Export JSON explainability (`--output json` global)
