# ADR-027 — Asagiri Experience Platform (UI foundation)

**Date :** 2026-05-29  
**Status :** accepted  
**Spec :** [`spec-ui.md`](../../spec-ui.md) §6–9, §23, §26–29 ; lot 1 (Phase 1)

## Context

Asagiri dispose déjà d’une façade terminal légère (`internal/tui`) pour le pipeline V3 (`work`, `--rich`, `--plain`, `--json`). La spec **Experience Platform** introduit un cockpit Bubble Tea complet (`asa` → Mission Control) qui compose les mêmes primitives CLI sans dupliquer la logique métier.

Contraintes :

- Go **1.25** (`go.mod` racine).
- Charm déjà partiellement présent (`lipgloss` v1.1.0).
- Moteurs existants : `runtime`, `trust`, `knowledge`, `replay`, `executiongraph`, `workflow`, `pipeline`, `investigation`, `coordination`.
- Invariant spec §3.2 : **aucune logique métier dans `internal/ui`**.

## Decision

### 1. Nouveau package `application/internal/ui/`

Arborescence lot 1 (spec §6.2, périmètre réduit lot 1) :

```
application/internal/ui/
  app/           # tea.Model racine, router, bootstrap CLI
  bus/           # CommandBus, QueryBus, handlers → services existants
  layout/        # engine minimal : single, split-h, split-v
  theme/         # tokens Lip Gloss, 5 thèmes nommés
  components/    # panel, card (stubs lot 1)
  input/         # keyboard base, mouse toggle config
  state/         # état UI local (écran, focus, layout) — pas d’état métier
  screens/
    mission/     # shell Mission Control (placeholder lot 1)
    settings/    # aperçu thème (lot 1)
```

**Hors lot 1** (répertoires vides ou `.gitkeep` uniquement) : `dashboard/`, `agents/`, `flows/`, `trust/`, `graph/`, `knowledge/`, `replay/`, `prototype/`.

### 2. Coexistence `internal/tui` vs `internal/ui`

| Package | Rôle | Consommateurs |
|---------|------|---------------|
| `internal/tui` | Façade **sortie** pipeline V3 (rich/plain/json, progress, live logs) — **pas** de event loop | `pipeline`, `cli/work`, `cli/estimate`, … |
| `internal/ui` | **Application** Bubble Tea (navigation, layout, buses) | `cli/root` (entrée `asa`), futures commandes `asa mission`, `asa dashboard` |

Règles :

- **Pas d’import** `ui` → `tui` ni `tui` → `ui`.
- Les deux lisent `config.UI` ; `ui.mode` (specv3) reste le fallback non-TTY ; champs Experience Platform (§29) ignorés par `tui`.
- Réutilisation visuelle autorisée : `theme/` peut s’appuyer sur les mêmes tokens couleur que `tui/rich_ui.go` via Lip Gloss, sans partager de struct.

### 3. Command / Query Bus

Couche **`internal/ui/bus/`** — seul endroit autorisé à importer les packages métier.

- Interfaces `CommandBus` / `QueryBus` (spec §7).
- Pattern **command/query typées** + dispatch par `Name()` (pas de reflection).
- Chaque command expose `CLIEquivalent() string` pour `ui.show_cli_equivalents`.
- Handlers délèguent aux **mêmes entrypoints** que les commandes Cobra (`cli/work.go`, `cli/trust_cmd.go`, etc.), idéalement via fonctions partagées extraites de `cli/` en lot 1 minimal ou appels directs documentés.

**Interdit** dans `components/`, `screens/`, `layout/` : import de `store/sqlite`, `trust`, `workflow`, `executiongraph`, etc.

### 4. Layout engine minimal (lot 1)

Implémenter uniquement (spec §9, lot 1) :

- `single`
- `split-horizontal`
- `split-vertical`

Reporter `grid`, `dashboard`, `focus`, `fullscreen` au lot 2+.

Entrée : largeur/hauteur terminal + `ui.compact_threshold`. Sortie : rectangles `PaneBounds` pour Lip Gloss.

### 5. Theme system (lot 1)

Thèmes nommés (spec §26) : `asagiri-dark` (défaut), `asagiri-light`, `minimal`, `high-contrast`, `cyber`.

Struct `theme.Theme` : palette Lip Gloss (primary, muted, success, warning, error, border, background). Sélection via `ui.theme`. `ui.animations: false` désactive spinners/transitions (accessibilité §28).

### 6. Extension `config.UIConfig`

Conserver champs specv3 existants et ajouter (spec §29) :

```yaml
ui:
  mode: auto                    # specv3 — inchangé
  live_logs: true
  progress_bars: true
  compact: false
  default_screen: mission
  theme: asagiri-dark
  mouse: true
  animations: true
  refresh_interval_ms: 500
  compact_threshold: 100
  show_cli_equivalents: true
  confirm_destructive_actions: true
```

Defaults dans `applyV3Defaults()` : `default_screen=mission`, `theme=asagiri-dark`, `refresh_interval_ms=500`, `compact_threshold=100`, booléens `true` sauf `compact`.

### 7. Entrée CLI

`application/internal/cli/root.go` :

- **`asa`** sans argument + stdout TTY → `ui/app.Run(ctx, opts)` (Mission Control / `default_screen`).
- **`asa`** sans argument + non-TTY → `cobra` help (comportement actuel implicite).
- Sous-commandes existantes **inchangées** (`asa work`, `asa graph`, …).

Détection TTY : `tui.DetectTTY(os.Stdout)` (réutilisation utilitaire, pas de dépendance UI).

### 8. Stack Charm (Go 1.25)

Ajouter au `go.mod` racine (versions compatibles Go 1.25, écosystème v1.x) :

| Module | Version cible |
|--------|---------------|
| `github.com/charmbracelet/bubbletea` | `v1.3.4` |
| `github.com/charmbracelet/lipgloss` | `v1.1.0` (déjà présent) |
| `github.com/charmbracelet/bubbles` | `v0.21.0` |
| `github.com/charmbracelet/huh` | `v0.7.0` |
| `github.com/charmbracelet/glamour` | `v0.10.0` |

Lot 1 : Bubble Tea + Lip Gloss obligatoires ; `bubbles` (viewport/table), `huh` (confirm destructive), `glamour` (markdown trust/report) — imports minimaux acceptables même si widgets complets lot 2+.

### 9. Tests lot 1

- Unitaires : `layout/` (bounds single/split), `theme/` (rendu stable), `bus/` (dispatch + `CLIEquivalent`).
- Golden : frame Mission Control shell (single layout).
- Intégration : `asa` en non-TTY → help ; smoke TTY optionnel CI (`script` / `test -t` skip).

## Consequences

- Nouveau surface area : maintenir parité CLI ↔ TUI pour chaque command bus (spec §3.1).
- Refactor léger `cli/` possible pour extraire `runWork`, `runVerifyTrust` partagés avec `bus` — acceptable lot 1 si documenté.
- Daemon runtime **optionnel** : queries runtime dégradent gracieusement si `runtime.Open` échoue (artefacts `.asagiri/runtime/` absents).
- Documentation publique Experience Platform (spec §31) — **lot 7**, hors scope lot 1.

## Related

- [`spec-ui.md`](../../spec-ui.md)
- ADR-010 (`internal/tui` isolée)
- [`02-architecture.md`](../ai/02-architecture.md) — section Experience Platform
