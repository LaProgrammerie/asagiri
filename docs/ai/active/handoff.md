# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-better-flow` : Executable Product Layer flow-centric & business-aware** (`2026-05-27`).

## Immediate objective

Implémenter l'extension `spec-better-flow.md` en s'appuyant sur `spec-prototype.md`, avec périmètre strict:
- core `application/internal/product/**` (modèle, validation, repository, génération d'artefacts, dérivation architecture);
- commandes CLI `prototype`, `flows`, `contracts`, `product review`, `architecture derive`, `spec generate-from-product`;
- artefacts `.asagiri/products/<product>/**`, `.asagiri/specs/<product>/**`, `.asagiri/tasks/<product>/**`;
- tests unitaires/golden/integration associés;
- mise à jour docgen/docs-site uniquement pour ces nouvelles commandes.

## Allowed scope (spec-better-flow)

- `application/internal/product/**`
- `application/internal/product/derivation/**`
- `application/internal/cli/root.go` + nouveaux fichiers de commandes CLI liés à la couche produit
- `application/internal/cli/docgen/**` (si nécessaire pour intégrer les nouvelles commandes)
- `application/internal/**/tests` + `*_test.go` liés à la couche produit et aux commandes ajoutées
- fixtures sous `application/internal/product/testdata/**` et fixtures d'intégration CLI dédiées
- `docs-site/**` (uniquement pages liées aux nouvelles commandes produit, si docgen requis)
- `docs/ai/active/handoff.md`, `docs/ai/active/current-spec.md`, `docs/ai/05-decisions.md`
- `spec-prototype.md`, `spec-better-flow.md`

## Definition of Done — spec-better-flow

- [ ] `asa prototype create/run/patch` implémenté (V1 déterministe).
- [ ] `asa flows extract/inspect` et `asa contracts extract` implémentés.
- [ ] `asa product review` et `asa spec generate-from-product` implémentés.
- [ ] `asa flows review` et `asa architecture derive` implémentés.
- [ ] `business.yaml` et enrichissement business/métriques des flows implémentés.
- [ ] Génération des tasks en flow-first avec sources explicites `flow/step/action/business objective`.
- [ ] Couplage métriques/analytics/contracts vérifié dans les rapports de dérivation/review.
- [ ] Artefacts générés sous `.asagiri/products/<product>/...`, `.asagiri/specs/<product>/...`, `.asagiri/tasks/<product>/...`.
- [ ] Tests unitaires/golden/integration ajoutés et stables.
- [ ] `go test ./...` vert et non-régression `asa work --plan-only`, `asa continue/next`.

## Hors scope

- Nouvelles features hors `spec-better-flow`.
- Refactor large sans lien direct avec la couche produit exécutable.
- Changement de branding/release déjà livré (`spec-rename`) sauf correctif bloquant.
- Commit / push / tag / release réelle par l’agent.

## References

- [`spec-prototype.md`](../../../spec-prototype.md)
- [`spec-better-flow.md`](../../../spec-better-flow.md)
- ADR-016, ADR-017 dans [`05-decisions.md`](../05-decisions.md)
