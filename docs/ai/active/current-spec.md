# Current spec — AgentFlow consolidation (post-V3)

**Phase :** `spec-postv123` (consolidation, fiabilisation, OSS)  
**Date :** 2026-05-17

## Résumé

Phase de **consolidation** après livraison V3 : cohérence architecture, sécurité, performance, tests, explainability, documentation OSS. Pas d’extension fonctionnelle majeure.

## Critères de phase

| Domaine | Statut |
|---------|--------|
| Audit architecture & drift | Livré (`docs/consolidation/01-*`) |
| API / primitives | Documenté (`02-*`) |
| Sécurité & fiabilité | Audit + garde-fous MCP/redact/collecte |
| Performance / coût | Quick win double-scan + benchmark |
| Workflows agents | Matrice manuelle (`05-*`) |
| Qualité | Tests ajoutés ; workflow/intent <50 % → roadmap |
| UX CLI explainability | estimate + work résumé |
| OSS readiness | Score 74/100 (`08-*`) |

## Documentation publique

- **Passe éditoriale (courante) :** [`spec-doc-v2.md`](../../../spec-doc-v2.md) — ton, prose, i18n (en/fr/de/es), sans changement fonctionnel
- **Spec site (structure) :** [`spec-doc.md`](../../../spec-doc.md)
- **Implémentation :** `docs-site/` (Fumadocs, i18n, static export → **Cloudflare Pages** via Wrangler CI)
- **Déploiement :** [`spec-deploy-doc.md`](../../../spec-deploy-doc.md) — secrets `CLOUDFLARE_*` ; `docs-check.yml` pour forks sans deploy
- **Génération CLI :** `agentflow docs generate-cli`

## Specs

- **Mission courante :** [`spec-postv123.md`](../../../spec-postv123.md)
- **V3 cost/perf :** [`specv3.md`](../../../specv3.md)
- **Intent :** [`specv2.md`](../../../specv2.md)
- **V1 :** [`spec.md`](../../../spec.md)

## Handoff actif

[`handoff.md`](handoff.md)

## Scores consolidation

Voir [`docs/consolidation/README.md`](../../consolidation/README.md) : OSS **74/100**, fiabilité **71/100**.
