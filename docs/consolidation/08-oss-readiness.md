# Open source readiness

**Date :** 2026-05-17 · **Score : 74/100**

## Grille

| Critère | Poids | Score | Notes |
|---------|-------|-------|-------|
| LICENSE claire | 15 | 14 | Apache 2.0 LaProgrammerie (remplace MIT JoliCode) |
| README quickstart | 15 | 13 | Install, commandes, philosophie, liens |
| CONTRIBUTING | 10 | 9 | Présent |
| ROADMAP | 5 | 5 | ROADMAP.md |
| Examples | 10 | 8 | `examples/quickstart/` |
| Pas de secrets | 10 | 9 | .gitignore OK ; audit manuel recommandé avant publish |
| Structure repo | 10 | 7 | Legacy PHP/E2E dans template |
| Tests / CI | 15 | 8 | Tests race verts ; pas de workflow GitHub release |
| Docs architecture | 10 | 8 | docs/ai + consolidation |
| Install reproductible | 10 | 9 | make build, go.mod |

## Livrables OSS créés / mis à jour

- `LICENSE` (Apache 2.0)
- `CONTRIBUTING.md`
- `ROADMAP.md`
- `examples/quickstart/README.md`
- `scripts/benchmark-workflow.sh` + `make benchmark`
- `docs/consolidation/*`

## Avant annonce publique

1. Workflow GitHub Actions (test, lint)
2. Badges README (build, license)
3. Renommer ou scinder artefact template PHP
4. Vérifier aucun secret dans historique git
