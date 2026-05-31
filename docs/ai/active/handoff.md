# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.
> **Tranche active :** **cockpit-consolidation** — Direction 4 « Asagiri
> Operations Cockpit ». **Livré** (`2026-05-31`, ADR-029) — Phases 0→4 ;
> `go test ./...` vert, `go vet` propre, `make build` ok.
> **Précédent :** project-onboarding — TUI wizard interactif — FULL (`2026-05-30`).

## Objectif

Consolider l'Experience Platform (ADR-027) sur l'infrastructure UI existante.
Corriger l'inversion d'effort : Mission Control (écran quotidien) est en texte
brut ; le wizard onboarding (one-shot) porte le chrome premium « EOS ». Aucune
grosse refonte : ~75 % des briques existent (layout engine, 13 widgets, palette,
event feed, design system, explorers, bus). Seul vrai développement neuf :
l'écran **Runs** + la requête bus `RunDetail`.

## Scope (autorisé)

- `application/internal/ui/components/` — `PanelSized`.
- `application/internal/ui/app/` — rail persistant dans `View()`, helpers shell,
  route Runs, onboarding via shell commun.
- `application/internal/ui/screens/mission/` — panelisation.
- `application/internal/ui/screens/runs/` — **nouveau** (liste + détail).
- `application/internal/ui/bus/` — `RunDetail`, `RunPipelineStep`,
  `GetRunDetailQuery` + handler d'agrégation.
- `application/internal/ui/screens/onboarding/` — réduction `eos_*` aux helpers
  partagés, suppression télémétrie fictive.
- `application/internal/ui/app/router.go`, `palette*.go` — `ScreenRuns` + entrée.

## Hors scope (interdit sans MAJ spec)

- Nouveau moteur d'orchestration ; toute logique métier dans `internal/ui`.
- `internal/tui` (specv3) ; `internal/onboarding/form.go` (logique onboarding).
- UI web/desktop.

## Ordre d'exécution

| Phase | Tâches | Donnée bus | DoD partiel | Statut |
|-------|--------|-----------|-------------|--------|
| 0 | CK-0.1 → CK-0.3 | inchangée | `PanelSized` + helpers shell testés | ✅ |
| 1 | CK-1.1 → CK-1.4 | inchangée | Mission Control panelisé responsive | ✅ |
| 2 | CK-2.1 → CK-2.5 | inchangée | rail persistant + nav sans régression | ✅ |
| 3 | RUN-3.1 → RUN-3.7 | **nouvelle** | écran Runs + `RunDetail` | ✅ |
| 4 | CK-4.1 → CK-4.6 | inchangée | shell unique, `eos_*` mort supprimé | ✅ |

MVP livrable après Phase 2 (cockpit crédible sans changement de données).

## Definition of Done

- [x] Mission Control panelisé + responsive ; rail persistant et state-aware
- [x] Runs écran de premier rang adossé à `RunDetail`
- [x] Onboarding dans le shell commun, sans télémétrie fictive, bascule Mission
- [x] Un seul chemin de rendu ; code `eos_*` mort supprimé
- [x] `go test ./... -count=1` vert ; `make build` ok ; équivalents CLI conservés
- [x] `06-spec-ui.md` + ADR (ADR-029) mis à jour ; `current-spec.md` / `handoff.md` synchro

## Garde-fous

- UI = client du bus (ADR-027). Pas de logique trust/workflow/runtime dans les écrans.
- Parité plain/json : fallback plat conservé, jamais conditionné au rendu panelisé.
- Landing Phases 0–2 (sans changement de données) avant de toucher l'onboarding.

## Références

- `.kiro/specs/cockpit-consolidation/` (requirements, design, tasks)
- [`06-spec-ui.md`](../06-spec-ui.md) — ADR-027 Experience Platform
- Code : `ui/app/app.go`, `ui/layout/`, `ui/screens/dashboard/widgets.go`,
  `ui/screens/mission/screen.go`, `ui/bus/bus.go`, `ui/theme/styles.go`
