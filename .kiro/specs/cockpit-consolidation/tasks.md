# cockpit-consolidation — Tasks

> Land in order. Phases 0–2 carry no bus-data change (MVP cockpit). Phase 3 is the
> only genuinely new development. Phase 4 removes the second shell.

## Phase 0 — Shared visual foundation

- [x] **CK-0.1** Add `components.PanelSized(title, body, w, h, theme)` from `eosColumn`
- [x] **CK-0.2** Extract nav-rail + top/bottom status-bar helpers into shared shell helpers (generalise `renderEOSNav` / `renderEOSTopBar` / `renderEOSBottomBar`)
- [x] **CK-0.3** Unit/golden tests for `PanelSized` (border, padding, height clamp)

## Phase 1 — Mission Control cockpit (MVP, no new data)

- [x] **CK-1.1** Rewrite `screens/mission/screen.go` to build panes (Runtime, Trust, Agents, Flow, Runs, Events) via `PanelSized`
- [x] **CK-1.2** Lay panes out with `layout.Engine` + `DashboardColumns` (1/2/3 by width)
- [x] **CK-1.3** Keep flat fallback for `ui.mode=plain|json`
- [x] **CK-1.4** Golden tests: compact / wide / ultra-wide

## Phase 2 — Persistent navigation rail

- [x] **CK-2.1** `renderNavRail(width, height)` in shell; active = `router.Current()`
- [x] **CK-2.2** State badges from snapshot (active runs, trust alert, queued events)
- [x] **CK-2.3** Insert rail in `app.go View()`; collapse below `compact_threshold`
- [x] **CK-2.4** Mouse: rail row click → `navigateTo`; keep `Ctrl+<letter>` + palette
- [x] **CK-2.5** Tests: active highlight follows router; rail collapse; click nav

## Phase 3 — Runs screen + detail (new data path)

- [x] **RUN-3.1** Spike: confirm available run fields from `workflow`/`runtime`/`trust`
- [x] **RUN-3.2** Bus: `RunDetail`, `RunPipelineStep`, `GetRunDetailQuery` + handler (aggregation, no logic in `ui/`)
- [x] **RUN-3.3** `ScreenRuns` in router + palette entry + nav rail entry
- [x] **RUN-3.4** `screens/runs/`: list pane + detail pane (pipeline glyphs, worktree, agents, validation, trust gate, events)
- [x] **RUN-3.5** Selection (keyboard + mouse) + drill-down `t`/`g`/`r`
- [x] **RUN-3.6** Empty state (not onboarded) → prompt `asa onboard --ui`
- [x] **RUN-3.7** Tests: handler unit, screen golden, selection/drill-down

## Phase 4 — Onboarding consolidation + cleanup

- [x] **CK-4.1** Route onboarding through shared shell (`layout.Focus`/`Fullscreen`); drop separate `renderFullscreenWizard` chrome
- [x] **CK-4.2** Remove fake telemetry (`analyzerConfidence`, "AGENT ACTIF" panel, wizard cost/API line)
- [x] **CK-4.3** Collapse dual renderer (`renderEOS*` vs `renderWizard*`) into one responsive path; keep `renderStepPanel` / `renderEOSReady`
- [x] **CK-4.4** Auto-navigate to Mission Control after apply + ready
- [x] **CK-4.5** Delete now-dead `eos_*` code; keep only shared helpers
- [x] **CK-4.6** Tests: onboarding renders in shell, no fake-telemetry strings, apply → mission

## Validation (each phase)

```bash
cd application && go test ./internal/ui/... -count=1
go test ./... -count=1
make build
./bin/asa            # panelised Mission Control + rail
./bin/asa runs       # runs list + detail (after Phase 3)
./bin/asa onboard --ui
```

## Definition of Done

- [x] Mission Control panelised + responsive; rail persistent and state-aware
- [x] Runs first-class screen backed by `RunDetail`
- [x] Onboarding shares the shell, no fake telemetry, hands off to Mission
- [x] Single rendering path; dead `eos_*` removed
- [x] `go test ./... -count=1` green; `make build` ok; CLI equivalents preserved
- [x] `06-spec-ui.md` + ADR updated; `current-spec.md` / `handoff.md` synced

**Statut global :** ✅ **livré** (`2026-05-31`, ADR-029). Phases 0→4 ; `go test ./...` vert, `go vet` propre, `make build` ok ; `asa runs` enregistré ; `eos_*` mort supprimé et télémétrie fictive retirée.
