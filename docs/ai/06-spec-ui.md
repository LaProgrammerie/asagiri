# Spec-ui — Asagiri Experience Platform (canon `docs/ai`)

**Statut :** FULL FEATURE livré (`2026-05-29`, audit reviewer) — réserves P1/P2 : voir [`active/handoff.md`](active/handoff.md#audit-reviewer-2026-05-29)  
**Spec racine :** [`spec-ui.md`](archives/specs/spec-ui.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)

---

## 1. Résumé

Spec-ui introduit l’Experience Platform terminal au-dessus des primitives CLI existantes :

- `asa` (TTY) ouvre Mission Control ;
- les commandes directes restent stables (`asa work`, `asa investigate`, `asa verify trust`, etc.) ;
- la TUI est un client du moteur via CommandBus/QueryBus ;
- aucune logique métier n’est portée dans `internal/ui`.

---

## 2. Entrées UI livrées

| Entrée | Comportement |
|---|---|
| `asa` (TTY) | Mission Control |
| `asa mission` | Mission Control |
| `asa dashboard` | Dashboard live |
| `asa agents watch` | Agent Theatre |
| `asa graph` | Graph Explorer |
| `asa flow` / `asa flow open <name>` | Flow Explorer |
| `asa knowledge` | Knowledge Explorer |
| `asa trust` | Trust Explorer |
| `asa replay open <id>` | Replay Explorer |
| `asa runs` | Runs screen (liste + détail) |
| `asa prototype` | Prototype Mode |
| `asa explain` | Explain screen |

---

## 3. Architecture runtime UI

| Zone | Rôle |
|---|---|
| `application/internal/ui/app` | Bubble Tea model principal, router, palette, help, refresh loop |
| `application/internal/ui/bus` | Adaptateurs CommandBus/QueryBus vers workflow/runtime/trust/knowledge/replay/graph |
| `application/internal/ui/screens/*` | Rendu read-only orienté cockpit |
| `application/internal/ui/components` | Panels, cards, event feed |
| `application/internal/ui/layout` | Layout single/split + focus navigation |
| `application/internal/ui/theme` | Palettes + styles (dark/light/minimal/high-contrast/cyber) |
| `application/internal/ui/input` | Raccourcis clavier + support souris |

Invariant : **coexistence** avec `internal/tui` (specv3) sans régression.

---

## 4. Polish lot 6 (§25–28)

Livré :

- resize souris de panneaux (basique) via wheel et drag de divider ;
- mode no-animation (glyphes statiques pour états `running`) ;
- usage explicite high-contrast (thème + aide accessibilité) ;
- responsive raffiné (dashboard compact/wide/ultra-wide) ;
- écran d’aide accessibilité + raccourcis (`?`).

---

## 5. Configuration `ui:`

Champs supportés :

- `mode`, `live_logs`, `progress_bars`, `compact` ;
- `default_screen`, `theme`, `mouse`, `animations` ;
- `refresh_interval_ms`, `compact_threshold` ;
- `show_cli_equivalents`, `confirm_destructive_actions`.

Exemple :

```yaml
ui:
  default_screen: mission
  theme: asagiri-dark
  mouse: true
  animations: true
  refresh_interval_ms: 500
  compact_threshold: 100
  show_cli_equivalents: true
  confirm_destructive_actions: true
```

---

## 6. Documentation publique (§31)

Les sections `experience/` sont publiées en **4 locales** :

- `en/experience/*`
- `fr/experience/*`
- `de/experience/*`
- `es/experience/*`

Pages : `index`, `mission-control`, `dashboard`, `command-palette`, `keyboard-shortcuts`, `mouse-support`, `themes`, `accessibility`.

---

## 8. Lot 7D — Mission / Prototype / Explain / souris / intégration

Livré :

- **Mission Control §11** : panneau *Recommended actions* (`GetRecommendedActionsQuery`) basé runtime/trust/queue/graph ;
- **Prototype Mode §19** : commandes `PrototypeCreate`, `FlowsExtract`, `ContractsExtract`, `SpecGenerateFromProduct` via CommandBus ; touches `1–4` dans l’écran prototype ;
- **Explain §21** : `FocusContext` + `ExplainContext` sur `GetExplainQuery` ; questions typées ; accessible depuis graph/flow/trust/knowledge (touche `e`) ;
- **Souris §10.2** : double-clic (ouvrir détail), hover/sélection, menu contextuel (clic droit), resize panneaux ;
- **Tests intégration §32** : `application/internal/ui/app/integration_test.go` (mission → dashboard → palette → commande → CLI equivalent).

Contrats produit : affichage `pending: <ref>` au lieu de `TODO:*` dans le rendu TUI.

---

## 7. Validation et clôture §33

Validation exécutée :

```bash
go test ./... -count=1
make build
make build && ./bin/asa docs generate-cli
cd docs-site && pnpm docs:check
```

**FULL FEATURE** : matrice handoff 100 % `[x]` ; `go test ./...` vert ; revue §33 signée `2026-05-29` (COMMENT — durcissement shimmer/tabs/test binaire optionnel).

---

## 9. Operations Cockpit — consolidation Direction 4 (ADR-029, `2026-05-31`)

Consolidation (pas refonte) corrigeant l'inversion d'effort UI : Mission Control
(écran quotidien) était en texte brut, le wizard onboarding (one-shot) portait le
chrome premium « EOS ». Livré en 5 phases, Phases 0–2 sans changement de données.

### 9.1 Fondation visuelle partagée (Phase 0)

- `components.PanelSized(title, body, w, h, theme)` — panneau borné (bordure,
  padding, clamp hauteur) dérivé de `eosColumn`.
- `components.RenderNavRail` / `RenderTopBar` / `RenderBottomBar` — shell partagé
  généralisé depuis `renderEOS*`. **Un seul jeu de helpers** pour le cockpit et
  l'onboarding.

### 9.2 Mission Control panelisé (Phase 1)

- `mission.RenderCockpit` construit les panneaux (Runtime, Trust, Agents, Active
  Flow, Runs, Events) via `PanelSized`, disposés par `layout.DashboardColumns`
  (1/2/3 colonnes selon la largeur). Fallback plat conservé pour `plain|json`.

### 9.3 Rail de navigation persistant (Phase 2)

- `app.renderNavRail` injecté dans `View()` ; entrée active = `router.Current()`.
- Badges d'état depuis le snapshot (runs actifs, alerte trust, file d'events).
- Collapse sous `compact_threshold` ; clic souris sur une ligne → `navigateTo` ;
  `Ctrl+<lettre>` + palette conservés.

### 9.4 Écran Runs + `RunDetail` (Phase 3 — seul vrai dev neuf)

- Bus : `RunDetail`, `RunPipelineStep`, `GetRunDetailQuery` + handler
  d'agrégation (`handleGetRunDetail`) combinant sqlite (`runs`/`tasks`/
  `run_metrics`), trust gate, agents actifs et events filtrés par run. **Aucune
  logique métier dans `ui/`** : pure agrégation de données déjà exposées.
- `screens/runs/` : liste + panneau détail (glyphes pipeline, worktree, agents,
  validation, trust gate, events) ; sélection clavier/souris ; drill-down
  `t` (trust) / `g` (graph) / `r` (replay) ; empty state → `asa onboard --ui`.
- Router `ScreenRuns` + entrée palette `nav.runs` + entrée rail + CLI `asa runs`.

### 9.5 Consolidation onboarding (Phase 4)

- Wizard fullscreen routé dans le **shell commun** (`components.*`) ; chrome
  `renderFullscreenWizard` bespoke supprimé.
- Télémétrie fictive retirée : `analyzerConfidence`, panneau « AGENT ACTIF »,
  ligne cost/API du bottom bar.
- Code `eos_*` mort supprimé ; seuls les helpers partagés et le contenu réutilisé
  (`renderStepPanel`, `renderWelcomePanel`, progression/readiness) subsistent.
- Bascule automatique vers Mission Control après apply + readiness `Ready`.

### 9.6 Validation

```bash
go test ./... -count=1   # vert
go vet ./...             # propre
make build               # ok (bin/asa)
```

Équivalents CLI préservés (`asa runs --help`) ; parité plain/json maintenue.
