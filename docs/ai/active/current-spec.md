# Current spec — Executable Product Layer (spec-better-flow)

**Phase :** `spec-better-flow`  
**Date :** 2026-05-27

## Spec active

- **Mission :** [`spec-better-flow.md`](../../../spec-better-flow.md)
- **Base existante :** [`spec-prototype.md`](../../../spec-prototype.md)
- **Source historique :** [`spec-postv123.md`](../../../spec-postv123.md)
- **Décisions :** ADR-016/017 (contexte rename) dans [`05-decisions.md`](../05-decisions.md)

## Résumé

Le rename **Asagiri / `asa`** est considéré livré. La tranche active étend la couche produit exécutable vers un mode **flow-centric et business-aware**: `business.yaml`, enrichissement des flows (objectif/métriques/implications), dérivation architecture (`asa architecture derive`), review de flows (`asa flows review`), génération des tasks en flow-first, et vérification du couplage métriques/analytics/contracts.

## Handoff actif

[`handoff.md`](handoff.md)

---

# Previous phase — release distribution (spec-release)

**Phase :** `spec-release` (GoReleaser, GitHub Releases, Homebrew)  
**Date :** 2026-05-17

- **Mission :** [`spec-release.md`](../../../spec-release.md)
- **Décision :** ADR-015 (distribution sur repo `asagiri` ; formule Homebrew migrée vers **`asa`** sous ADR-016)

---

# Previous phase — Asagiri rebrand (spec-rename)

**Phase :** `spec-rename` + module path **`github.com/LaProgrammerie/asagiri`**  
**Date :** 2026-05-20

- **Mission :** [`spec-rename.md`](../../../spec-rename.md)
- **Résultat :** rebranding livré (produit, CLI, module, URLs)
- **Migration repo GitHub :** déjà effectuée (runbook: [`docs/migration/github-rename-asagiri.md`](../../migration/github-rename-asagiri.md))

---

# Previous phase — consolidation (post-V3)

**Phase :** `spec-postv123` (consolidation, fiabilisation, OSS)  
**Date :** 2026-05-17

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

- **Passe éditoriale :** [`spec-doc-v2.md`](../../../spec-doc-v2.md)
- **Spec site (structure) :** [`spec-doc.md`](../../../spec-doc.md)
- **Implémentation :** `docs-site/` (Fumadocs, i18n, static export → **Cloudflare Pages**)
- **Déploiement :** [`spec-deploy-doc.md`](../../../spec-deploy-doc.md) — projet Pages **`asagiri-docs`** (ADR-016)
- **Génération CLI :** `asa docs generate-cli`

## Specs historiques

- [`spec-postv123.md`](../../../spec-postv123.md), [`specv3.md`](../../../specv3.md), [`specv2.md`](../../../specv2.md), [`spec.md`](../../../spec.md)

## Scores consolidation

Voir [`docs/consolidation/README.md`](../../consolidation/README.md) : OSS **74/100**, fiabilité **71/100**.
