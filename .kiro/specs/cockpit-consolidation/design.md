# cockpit-consolidation — Design

## Principle

Consolidation over rewrite. Reuse the existing bricks; add only the Runs detail
data path. UI stays a bus client (ADR-027): no trust/workflow/runtime logic in
`internal/ui`.

## Existing bricks (reused as-is)

| Brick | Location | Use |
|-------|----------|-----|
| Shell loop, router, palette, help, modals, mouse/resize | `ui/app/` | Host the cockpit |
| `layout.Engine` (single/split/grid/dashboard/focus/fullscreen) | `ui/layout/` | Pane geometry |
| 13 widgets (`Runtime…Performance`) | `ui/screens/dashboard/widgets.go` | Mission panes |
| `MissionControlSnapshotResult` | `ui/bus/bus.go` | Mission data (unchanged) |
| Design system (`Brand`, `NavActive`, `RenderBadge`, …) | `ui/theme/styles.go` | Styling |
| Event feed, explorers (graph/flow/trust/knowledge/replay), agents | `ui/components/`, `ui/screens/*` | Drill-downs |

## New / changed components

### Phase 0 — Shared visual foundation

```go
// components/panel.go
func PanelSized(title, body string, w, h int, th theme.Theme) string
```

- Derive from `screens/onboarding.eosColumn` (borders, padding, height clamp).
- Add `shell` helpers (in `ui/app` or `ui/components`) for the nav rail and
  top/bottom status bars, generalised from `renderEOSNav` / `renderEOSTopBar` /
  `renderEOSBottomBar` so both the shell and onboarding consume them.

### Phase 1 — Mission Control cockpit

Rewrite `screens/mission/screen.go`:

```text
ViewModel (unchanged inputs) → build panes → layout.Engine → PanelSized grid
```

- Panes: Runtime, Trust, Agents, Flow, Runs (summary), Events — each a widget
  wrapped in `PanelSized`.
- Column count from `layout.DashboardColumns(width, compactThreshold)`.
- Keep `Render(vm) string` signature; switch internal body from text to panes.
- Plain/json mode: keep a flat fallback when `ui.mode != rich`.

### Phase 2 — Persistent rail in the shell

`app.go View()` gains a left rail before the body:

```text
┌ rail ─┬ body ───────────────┐
│ nav   │ current screen       │
│ +state│                      │
└───────┴──────────────────────┘
```

```go
type navItem struct {
    screen string // router constant
    label  string
    icon   string
    badge  string // optional, from snapshot (runs, trust, queue)
}
func (m model) renderNavRail(width, height int) string
```

- Active item = `m.router.Current()`.
- Badges derived from `m.snapshot` (active runs, trust alert, queued events).
- Click handling: map rail row → `navigateTo(screen, cli)` (reuse existing).
- Collapse below `compact_threshold` (rail hidden, shortcuts/palette remain).

### Phase 3 — Runs screen + detail

Router:

```go
const ScreenRuns = "runs"   // add to router.Set switch + nav rail + palette
```

Bus:

```go
// ui/bus
type RunPipelineStep struct { ID, Label, Status string }
type RunDetail struct {
    ID, Feature, Status string
    Worktree            string
    Pipeline            []RunPipelineStep   // spec→plan→dev→verify→trust→report
    Agents              []ActiveAgentSummary
    Validation          string              // command + state
    TrustGate           TrustSummaryResult   // score + threshold
    CostEUR             float64
    Events              []EventSummary
    CreatedAt, UpdatedAt time.Time
}
type GetRunDetailQuery struct { RunID string }
```

- Handler aggregates from `workflow` (run/steps), `runtime` (agents/events),
  `trust` (gate). Implemented in `ui/bus` adapters, not in screens.
- `screens/runs/`:
  - list pane: reuse `RunSummary` list + selection model (mirror trust/replay
    explorer pattern);
  - detail pane: pipeline glyph row (reuse `eos_widgets.stepGlyph` /
    `renderProgressBar`), worktree, agents, validation, trust gate bar, events.
- Drill-downs: `t`/`g`/`r` from a selected run → existing screens.
- Empty state: not-onboarded → prompt `asa onboard --ui`.

### Phase 4 — Onboarding consolidation

- Route onboarding through `app.go View()` using `layout.Focus` /
  `layout.Fullscreen`, dropping `renderFullscreenWizard` as a separate chrome.
- Delete fake telemetry: `analyzerConfidence`, the "AGENT ACTIF" panel,
  wizard cost/API bottom-bar fields.
- Collapse the dual renderer (`renderEOS*` vs `renderWizard*`) into one
  responsive path; keep `renderStepPanel` / `renderEOSReady` content.
- After apply + ready → `navigateTo(ScreenMission, "asa mission")`.
- Keep `internal/onboarding/form.go` and the bus onboarding commands untouched.

## Files touched (indicative)

| File | Change |
|------|--------|
| `ui/components/panel.go` | add `PanelSized` |
| `ui/app/shell_*.go` (new) | nav rail + status bars helpers |
| `ui/app/app.go` | rail in `View()`, runs route, onboarding via shell |
| `ui/app/router.go` | `ScreenRuns` |
| `ui/screens/mission/screen.go` | panelise |
| `ui/screens/runs/*` (new) | runs list + detail |
| `ui/bus/*` | `RunDetail`, `GetRunDetailQuery` + handler |
| `ui/screens/onboarding/eos_*.go` | reduce to shared helpers / delete |
| `ui/app/palette*.go` | add Runs entry |

## Testing strategy

- Golden render tests: Mission Control panelised at compact / wide / ultra-wide.
- Nav rail: active highlight follows router; badges from a fixture snapshot.
- Runs: `GetRunDetailQuery` handler unit test (aggregation from fake workflow/
  trust/runtime); screen render golden; selection + drill-down navigation.
- Onboarding: render inside shell; no fake-telemetry strings; apply → mission.
- Regression: existing `ui/app/integration_test.go` (mission → dashboard →
  palette → command → CLI equivalent) stays green; explorers unchanged.

## Dependencies

- No new external deps (Charm stack already pinned, ADR-027).
- Bus aggregation reuses existing `workflow` / `runtime` / `trust` services.

## Risks / mitigations

- **`RunDetail` aggregation surface unknown** → spike first (Phase 3 task RUN-3.1)
  to confirm available fields from `workflow`/`runtime` before screen work.
- **Plain/json parity** → keep flat fallbacks; do not gate CI/screen-reader modes
  on the panelised path.
- **Two-shell removal regressions** → land Phases 0–2 (no data change) before
  touching onboarding (Phase 4).
