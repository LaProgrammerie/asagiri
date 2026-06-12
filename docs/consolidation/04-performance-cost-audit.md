# Performance & coût/tokens

**Date :** 2026-05-17 · **Référence :** spec-postv123 §4

## Pipeline V3 (ordre)

1. `investigation.Run` — grep, scan borné, symboles Go
2. `contextopt.CollectForPipeline` — **candidats + specs** (plus d’arbre `.` systématique si candidats présents)
3. `Reduce` + `BuildPack`
4. `cost.BuildEstimate` + budget
5. Exécution intent → primitives

## Quick win appliqué

- **Évite double scan** : économie I/O proportionnelle à la taille du repo hors chemins candidats.

## Pistes non implémentées

| Piste | Gain estimé |
|-------|-------------|
| Cache pack (feature, task, git rev) | 30–60 % temps estimate répété |
| Réutiliser contenu fichiers déjà lus en investigation | 10–20 % I/O |
| Tokenizer tiktoken/provider | Précision coût (pas vitesse) |
| Compression excerpts (compressor.go) activée par défaut sur gros fichiers | Tokens cloud |

## Benchmark

```bash
make benchmark
# scripts/benchmark-workflow.sh — dry-run estimate + work --plan-only
```

Métriques affichées par `estimate` : tokens in/out, coût, durée, budget, steps explainability.

## Routing

`routing.Route` intégré à l’estimation (raison tier par step). Executor utilise encore agents du plan — alignement futur recommandé.
