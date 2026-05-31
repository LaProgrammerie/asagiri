# Current spec — cockpit-consolidation (active)

**Phase :** Direction 4 « Asagiri Operations Cockpit » — consolidation de
l'Experience Platform sur l'infrastructure existante — **livré** (`2026-05-31`,
ADR-029 ; `go test ./...` vert, `make build` ok).
**Handoff :** [`handoff.md`](handoff.md)

## Objet

Corriger l'inversion d'effort UI : Mission Control (écran quotidien par défaut)
est rendu en texte brut alors que le wizard d'onboarding (one-shot) a reçu le
chrome premium « EOS ». Consolider sur les briques déjà livrées (layout engine,
widgets, palette, event feed, design system, bus, explorers) ; ajouter le seul
manque réel : l'écran **Runs** + sa requête `RunDetail`.

## Spec

- **Kiro :** `.kiro/specs/cockpit-consolidation/` (requirements, design, tasks)
- **Canon UI :** [`06-spec-ui.md`](../06-spec-ui.md) (ADR-027, Experience Platform)
- **Code cible :** `application/internal/ui/` (app, screens/mission, screens/runs,
  bus, components, screens/onboarding)

## Phasage

| Phase | Focus | Donnée bus | Statut |
|-------|-------|-----------|--------|
| 0 | Fondation visuelle partagée (`PanelSized`, rail/status-bar helpers) | inchangée | livré |
| 1 | Mission Control panelisé (MVP cockpit) | inchangée | livré |
| 2 | Rail de navigation persistant + badges d'état | inchangée | livré |
| 3 | Écran Runs + `GetRunDetailQuery`/`RunDetail` | **nouvelle** | livré |
| 4 | Fusion onboarding dans le shell + suppression second shell | inchangée | livré |

MVP cockpit = Phases 0–2 (aucun changement de données). Phase 3 = seul vrai
développement neuf. Phase 4 = nettoyage du shell parallèle.

## Invariants

- UI = client du bus (ADR-027) ; aucune logique métier dans `internal/ui`.
- Pas de régression `internal/tui` (specv3 rich/plain/json).
- Équivalents CLI conservés ; parité plain/json.

## Précédent (livré)

- **project-onboarding** — Project Onboarding & Readiness + TUI wizard — FULL
  livré (`2026-05-30`).
- **spec-ui** — Experience Platform FULL FEATURE — livré (`2026-05-29`).

Branding : **Asagiri** / **`asa`**.
