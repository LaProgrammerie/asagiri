# cockpit-consolidation — Requirements

## Problem

The Experience Platform (ADR-027) ships the full daily-driver surface — Mission
Control plus the explorers (Dashboard, Agents, Graph, Flow, Trust, Knowledge,
Replay, Logs, Explain) — but the **investment is inverted**:

- **Mission Control** (the default `asa` screen, used daily) renders as plain
  stacked text (`"Mission Control\n=====\n"` + `\n\n` sections). It does not use
  `components.Panel`, the dashboard widgets, or `layout.Engine`, even though all
  three exist and are exercised by the Dashboard screen.
- The **onboarding wizard** (one-shot) received a bespoke premium chrome
  (3-column "EOS" layout, nav rail, status bars, badges, gauges) via a parallel
  rendering path (`renderFullscreenWizard` + `eos_layout.go`/`eos_decor.go`/
  `eos_widgets.go`) that bypasses the shared shell.
- There is **no persistent navigation rail** in the shared shell: navigation
  relies on memorised `Ctrl+<letter>` shortcuts and the palette.
- **Runs** — the central object of the product (spec → worktree → agents →
  validation → trust → report) — are not a first-class screen; they are a
  sub-section of Mission Control, and `bus.RunSummary` exposes only
  `{ID, Feature, Status, CreatedAt, UpdatedAt}` with no run detail.
- The wizard fabricates **fake telemetry** ("Agent Analyzer / Analyzing /
  confidence %") computed from form-fill, contradicting the product's
  deterministic / auditable positioning.

Goal: consolidate onto the **existing** cockpit infrastructure (Direction 4 —
"Asagiri Operations Cockpit"). This is consolidation, not a rewrite: ~75% of the
bricks exist. The only genuinely new development is the Runs screen and its bus
detail query.

## Goals

1. **One shell, one visual language** — promote the EOS visual bricks into the
   shared shell; remove the second (wizard) rendering path.
2. **Mission Control as a real cockpit** — panelised, responsive, built from the
   existing widgets and `layout.Engine`, with no new data.
3. **Persistent navigation rail** — visible, state-aware, wired to the existing
   router.
4. **Runs as a first-class screen** — list + detail pane, backed by a new bus
   `RunDetail` query (pipeline, worktree, per-run agents, validation, trust
   gate, cost, events).
5. **Honest onboarding** — slim, focused, runs inside the shared shell, hands off
   to Mission Control; no fake telemetry.

## Users

- Developers running `asa` (TTY) as their daily orchestration cockpit.
- The same template-fork audience as `project-onboarding`.

## Functional requirements

### FR-1 Shared visual foundation

- FR-1.1 Add `components.PanelSized(title, body string, w, h int, th theme.Theme)`
  (width/height-aware), derived from `screens/onboarding.eosColumn`.
- FR-1.2 Move nav-rail rendering and top/bottom status bars out of
  `screens/onboarding` into reusable `components`/shell helpers.
- FR-1.3 Keep the existing `theme/styles.go` design-system helpers (`Brand`,
  `NavActive`, `SectionHead`, `HeroTitle`, `RenderBadge`, `RenderBarGauge`,
  `RenderKVGrid`, `RenderStatusBarFull`) as the single source of styling.

### FR-2 Mission Control cockpit

- FR-2.1 Re-render `screens/mission` using `PanelSized` panes laid out via
  `layout.Engine` (split/grid by width), reusing the 13 existing widgets.
- FR-2.2 No new bus data: consume `MissionControlSnapshotResult` as-is.
- FR-2.3 Responsive: 1 / 2 / 3 columns via `layout.DashboardColumns`; degrade to
  stacked panes below `compact_threshold`.
- FR-2.4 Preserve plain/json output mode parity (`ui.mode`) and CLI equivalents.

### FR-3 Persistent navigation rail

- FR-3.1 Render a left rail in `app.go View()` listing the real router screens,
  highlighting the current one (driven by `router.Current()`).
- FR-3.2 Show state badges per destination when data is available (e.g. active
  runs count, trust alert, queued events).
- FR-3.3 Rail entries are navigable by click and keep existing `Ctrl+<letter>` /
  palette shortcuts; no regression to current navigation.
- FR-3.4 Rail collapses below `compact_threshold`.

### FR-4 Runs screen + detail

- FR-4.1 Add `ScreenRuns` to the router and `screens/runs/`.
- FR-4.2 Add bus `GetRunDetailQuery` → `RunDetail` aggregating pipeline steps,
  worktree, per-run agents, validation status, trust gate, cost, recent events
  from `workflow` / `runtime` / `trust` (no business logic in `ui/`).
- FR-4.3 Runs screen = list pane + detail pane; selection updates detail;
  `Enter` opens, keyboard + mouse selection consistent with other explorers.
- FR-4.4 Drill-downs from a run to `trust` / `graph` / `replay` reuse existing
  navigation (`t` / `g` / `r`).
- FR-4.5 Empty state when the repo is not onboarded: prompt `asa onboard --ui`.

### FR-5 Onboarding consolidation

- FR-5.1 Onboarding renders inside the shared shell (Focus/Fullscreen layout),
  not via a separate `renderFullscreenWizard` path.
- FR-5.2 Remove the fake "Agent Analyzer / confidence" panel and
  `analyzerConfidence()`; remove the empty cost/API line from the wizard.
- FR-5.3 Reduce progression to a single representation.
- FR-5.4 On apply + ready, auto-navigate to Mission Control.
- FR-5.5 Preserve wizard business logic (`internal/onboarding/form.go`, field
  rows, readiness apply/autofix) unchanged.

## Non-goals

- No new orchestration engine; UI stays a bus client (ADR-027 invariant).
- No change to `internal/tui` (specv3 rich/plain/json pipeline output).
- No web/desktop UI; terminal only.
- No change to the onboarding domain logic or readiness scoring.

## Acceptance

- `asa` (TTY) opens a panelised Mission Control with a persistent rail.
- `asa runs` (or palette) opens a Runs list + detail backed by `RunDetail`.
- Onboarding shares the shell, has no fake telemetry, and hands off to Mission
  Control.
- The `eos_layout.go` / `eos_decor.go` / `eos_widgets.go` parallel shell is
  removed or reduced to shared helpers; one rendering path remains.
- `go test ./... -count=1` green; `make build` ok; CLI equivalents preserved.
