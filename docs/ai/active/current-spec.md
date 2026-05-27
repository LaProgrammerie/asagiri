# Current spec — Asagiri Executable Product Layer (spec-my-A)

**Phase :** `spec-my-A`  
**Date :** 2026-05-27

## Spec active

- **Mission :** [`spec-my-A.md`](../../../spec-my-A.md) (~3200 lignes, §1–26)
- **Historique :** [`spec-better-flow.md`](../../../spec-better-flow.md), [`spec-prototype.md`](../../../spec-prototype.md)
- **Handoff :** [`handoff.md`](handoff.md)

## Résumé

Extension complète d'Asagiri :

1. **Executable Product Layer** — prototype Vite/React déterministe, flows, contracts, specs/tasks pour `asa work`
2. **Business Intent Layer** — `business.yaml`, flows enrichis, `asa flows review`, `asa architecture derive`, tasks traçables
3. **Persistent Runtime** — `asa daemon`, `asa session`, `asa runtime events`, SQLite `.asagiri/runtime/`
4. **Investigation Engine** — `asa investigate` structuré, context/replay packs, `asa work --investigate-first`

Branding : **Asagiri** / CLI **`asa`** / module `github.com/LaProgrammerie/asagiri`.

## Écarts V1 assumés

| Domaine | Écart | Suite |
|---------|-------|-------|
| Runtime | Pas de process daemon background réel | `asa daemon run` + supervisor |
| Runtime | Memory/skills/hooks exécutables non livrés | phases §24.10–14 |
| Investigation | Pas de root-cause graph | §25.25 phase 7 |
| Investigation | `--from-failed-tests` heuristique | parse sortie `go test` |

## Previous phases

Voir sections historiques ci-dessous dans ce fichier (spec-release, spec-rename, spec-postv123).

---

# Previous phase — spec-better-flow

**Phase :** `spec-better-flow` (fusionnée dans spec-my-A)  
**Date :** 2026-05-27

Flow-centric + business-aware ; détail dans `spec-my-A.md` §23.

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
