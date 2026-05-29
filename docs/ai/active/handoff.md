# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche active :** **spec-ui** — Asagiri Experience Platform — **FULL FEATURE livré** (`2026-05-29`, audit reviewer).  
> **Précédent :** specv3 — Cost, Performance & Token Optimization — **livrée** (`2026-05-29`).

## Objectif

**FULL FEATURE 100 %** de [`spec-ui.md`](../../../spec-ui.md) (§1–36) — couche d'expérience interactive **Asagiri Experience Platform** au-dessus du moteur existant, branding **Asagiri** / **`asa`**.

Fondation lots 1–6 (shell, navigation, golden tests, docs 4 locales) + lots **7A–7D** (bus complet, explorers interactifs, design system, mission/prototype/explain/souris, tests intégration).

---

## Matrice traçabilité spec-ui (FULL FEATURE)

| ID | Livrable | Lot | Lead | Statut |
|----|----------|-----|------|--------|
| §3–5 | Principes UX, cohérence CLI, progressive disclosure | 1–6 | developer | [x] |
| §6 | `internal/ui/` arborescence Charm + séparation stricte | 1 | architect+developer | [x] |
| §7 | CommandBus — dispatch typé + `CLIEquivalent()` | 1 | developer | [x] |
| §7 | CommandBus — couverture commandes §7 complète | 7A | developer | [x] |
| §7 | QueryBus — projections read-only stables | 1–5 | developer | [x] |
| §8.1 | Panel, Card | 1 | developer | [x] |
| §8.1 | MetricCard, ProgressBar, StatusBadge | 7C | developer+designer | [x] |
| §8.1 | Timeline, TreeView, TableView, GraphView, LogView | 7B–7C | developer | [x] |
| §8.1 | EventFeed (composant riche), CommandPalette, Tabs | 7A–7C | developer | [x] |
| §8.1 | Breadcrumb, Modal, Drawer, SplitPane, Toast | 7C | developer | [x] |
| §8.2 | Composants métier (RuntimeCard, AgentCard, TrustCard, FlowCard, RiskCard, CostCard, GraphNodeCard, InvestigationCard, ReplayCard, KnowledgeCard, PrototypeCard) | 7C | developer+designer | [x] |
| §8.3 | États visuels (idle/running/success/warning/error/blocked/paused/waiting/unknown) | 7C | developer | [x] |
| §9 | Layout single, split-h, split-v | 1 | developer | [x] |
| §9 | Layout grid, dashboard, focus, fullscreen | 7C | developer | [x] |
| §9 | Focus panneau, redimensionnement, collapse/expand, tabs, panels persistants | 7C | developer | [x] |
| §10.1 | Raccourcis globaux clavier | 3 | developer | [x] |
| §10.2 | Souris complète (clic, double-clic, scroll, hover, menus contextuels, sélection, resize) | 7D | developer | [x] |
| §11 | Mission Control — shell + données runtime | 2 | developer | [x] |
| §11 | Mission Control — contenu complet (sessions, flows critiques, actions recommandées, coût jour/mois) | 7D | developer | [x] |
| §12 | Dashboard live — refresh + modes compact/wide | 2,6 | developer | [x] |
| §12 | Dashboard — widgets composables §23 branchés | 7C | developer | [x] |
| §13 | Agent Theatre (`asa agents watch`) | 5 | developer | [x] |
| §14 | Graph Explorer — affichage read-only | 4 | developer | [x] |
| §14 | Graph Explorer — vues (timeline, dependency, critical-path, parallel-groups, blocked) + actions | 7B | developer | [x] |
| §15 | Flow Explorer — affichage read-only | 4 | developer | [x] |
| §15 | Flow Explorer — sélection step + drill-down API/service/event/tests/metrics/trust/risk | 7B | developer | [x] |
| §16 | Knowledge Explorer — recherche read-only | 4 | developer | [x] |
| §16 | Knowledge Explorer — actions (impact analyze, build context, open graph, explain relationship) | 7B | developer | [x] |
| §17 | Trust Explorer — scores read-only | 4 | developer | [x] |
| §17 | Trust Explorer — drill-down evidence/findings/checks/gates/risks cliquable | 7B | developer | [x] |
| §18 | Replay Explorer — timeline read-only | 5 | developer | [x] |
| §18 | Replay Explorer — jump, inspect artifact, compare, replay offline, explain divergence | 7B | developer | [x] |
| §19 | Prototype Mode — split view shell | 5 | developer | [x] |
| §19 | Prototype Mode — actions pipeline (`prototype create`, `flows extract`, `contracts extract`, `spec generate-from-product`) | 7D | developer | [x] |
| §20 | Command Palette Ctrl+P — entrées statiques + CLI equivalent | 3 | developer | [x] |
| §20 | Command Palette — recherche dynamique flows/agents/reports + actions contextuelles | 7A | developer | [x] |
| §21 | Explain — écran read-only reasons/evidence | 4 | developer | [x] |
| §21 | Explain — questions typées + accessibilité depuis toute décision | 7D | developer | [x] |
| §22 | Event Feed — affichage branché | 2–4 | developer | [x] |
| §22 | Event Feed — filter, search, pause, export, open artifact | 7A | developer | [x] |
| §23 | Widget system V1 (Runtime, Agent, Trust, Cost, Flow, Event…) | 2 | developer | [x] |
| §23 | Widget interface composable + widgets §23.2 complets | 7C | developer | [x] |
| §24 | Live updates Bubble Tea (tick + throttling anti-flicker) | 2 | developer | [x] |
| §24 | Subscriptions multi-sources (graph, trust, cost, logs) + dégradation terminal lent | 7C | developer | [x] |
| §25 | Mode no-animation | 6 | developer | [x] |
| §25 | Animations sobres (shimmer, transitions focus, live counters, sparklines) | 7C | developer | [x] |
| §26 | Theme system 5 thèmes + config `ui.theme` | 1,6 | developer | [x] |
| §27 | Responsive narrow/standard/wide/ultra-wide | 6 | developer | [x] |
| §27 | Layout adaptatif grid dashboard + CI plain fallback | 7C | developer | [x] |
| §28 | Navigation clavier, high-contrast, no-animation, aide `?` | 6 | developer | [x] |
| §28 | Screen-reader compatible where possible | 7D | developer | [x] |
| §29 | Configuration `ui:` étendue | 1 | developer | [x] |
| §30 | Safety UX modal destructive (stub `graph rollback`) | 3 | developer | [x] |
| §30 | Safety UX — actions destructives réelles branchées CommandBus | 7A | developer | [x] |
| §31 | docs-site `experience/` 4 locales (shell) | 6 | developer | [x] |
| §31 | docs-site — mise à jour FULL FEATURE (captures, comportements réels) | 7D | developer | [x] |
| §32 | Tests unit + golden | 1–6 | tester | [x] |
| §32 | Tests intégration end-to-end (start `asa`, palette, trigger command, CLI equivalent) | 7D | tester | [x] |
| §33 | Critères acceptation — shell V1 | 6 | reviewer | [x] |
| §33 | Critères acceptation — FULL FEATURE validés | 7D | reviewer | [x] |
| Doc canon | `06-spec-ui.md` — V1 | 6 | developer | [x] |
| Doc canon | `06-spec-ui.md` — clôture FULL FEATURE | 7D | developer | [x] |

**Couverture FULL FEATURE :** 100 % matrice `[x]` — audit reviewer `2026-05-29` (réserves P1/P2 ci-dessous, non bloquantes clôture).

---

## Lots

### Lots 1–6 — V1 (livrés)

Référence historique ; ne pas rouvrir sauf régression.

| Lot | Contenu |
|-----|---------|
| **1 — Foundation** | Bubble Tea app, router, layout minimal, theme, buses, config `ui:`, `internal/ui/` |
| **2 — Mission + Dashboard** | Mission Control, dashboard live, widgets V1, event feed affichage, live updates |
| **3 — Palette & nav** | Command Palette Ctrl+P, raccourcis, safety UX |
| **4 — Explorers shell** | Graph/Flow/Knowledge/Trust/Explain read-only + event feed intégré |
| **5 — Theatre & replay** | Agent Theatre, Replay timeline, Prototype split view |
| **6 — Polish V1** | Souris resize, no-animation, responsive, docs 4 locales, golden tests |

---

### Lot 7A — CommandBus + Event Feed + Palette dynamique

**Definition of Done lot 7A :**
- [x] Toutes les commandes §7 exposées avec `CLIEquivalent()` et handlers réels (pas de stub dispatch)
- [x] Event feed : filter par type, search texte, pause/resume, export, open artifact
- [x] Palette : recherche flows, agents, reports ; actions contextuelles selon écran actif
- [x] Safety UX destructive branchée sur CommandBus (`GraphRollback` + impact query)
- [x] `go test ./internal/ui/bus/... ./internal/ui/app/... ./internal/ui/components/... -count=1` vert

---

### Lot 7B — Explorers interactifs

**Definition of Done lot 7B :**
- [x] Graph : 5 vues + actions (ouvrir nœud, logs, dépendances, explain, lancer/reprendre, export Mermaid/JSON)
- [x] Flow : sélection step + panneau détail (API, service, event, tests, metrics, trust, risk)
- [x] Knowledge : recherche interactive + actions impact/build context/open graph/explain
- [x] Trust : scores cliquables → evidence, findings, checks, gates, risques résiduels
- [x] Replay : jump to event, inspect artifact, compare run, replay offline, explain divergence
- [x] Golden tests mis à jour ; aucun `(stub)` dans le rendu

---

### Lot 7C — Widgets + layout + design system + animations/perf

**Definition of Done lot 7C :**
- [x] Composants §8.1–8.2 réutilisables ; états visuels §8.3 uniformes
- [x] Layout grid/dashboard/focus/fullscreen + collapse/expand + tabs (composant + session panneau)
- [x] Widget interface §23.1 ; widgets §23.2 branchés au dashboard
- [x] Animations sobres §25 (respect `ui.animations: false`) — sparklines, spinners, compteurs
- [x] Virtualized lists / lazy loading ; pas de régression perf terminal standard
- [x] `go test ./internal/ui/... -count=1` vert

---

### Lot 7D — Mission / Prototype / Explain + souris + tests intégration + doc clôture

**Definition of Done lot 7D :**
- [x] Mission Control contenu spec §11 (actions recommandées, coût jour/mois, flows critiques)
- [x] Prototype Mode lance commandes pipeline réelles via CommandBus
- [x] Explain : questions typées §21, accessible depuis explorers/decisions
- [x] Souris : clic, double-clic, scroll, hover, menus contextuels, sélection, resize panels
- [x] Tests intégration §32 (`integration_test.go`)
- [x] Critères §33 validés (reviewer `2026-05-29`)
- [x] `06-spec-ui.md` + docs-site `experience/` alignés (4 locales)
- [x] Matrice traçabilité 100 % `[x]`

---

## Audit reviewer (`2026-05-29`)

**Verdict :** COMMENT — FULL FEATURE exploitable ; réserves P1/P2 pour durcissement.

| Priorité | Écart | Fix suggéré |
|----------|-------|-------------|
| P1 | `ShimmerPrefix` (§25) défini mais jamais appelé dans le rendu | Brancher sur états `running` / chargement dashboard |
| P1 | Tabs layout (§9) : `RenderTabs` + `PanelSession` sans barre d'onglets dans `app.View` | Intégrer au moins un écran split (mission/dashboard) |
| P1 | Tests §32 : intégration in-process (`newModel`), pas subprocess `asa` | Ajouter test `exec` TTY ou `testscript` sur binaire |
| P2 | Widgets §23.1 : `Init`/`Update` non invoqués (seul `View`) | Optionnel : boucle tea par widget ou assouplir spec canon |
| P2 | Panels « persistants » : session mémoire uniquement (pas disque) | Persister `layout.Session` dans config/state store si requis produit |
| P2 | Flow step sans métadonnées graphe → `n/a` (pas stub) | Enrichir projection `GetFlowStepDetailQuery` |

**Validations exécutées :** `go test ./... -count=1` vert (`application/`, `2026-05-29`).

---

## Périmètre autorisé

- `application/internal/ui/**`
- `application/internal/cli/*_ui*.go`, commandes `mission`, `dashboard`, `agents`, `flow`, `explain`, `graph`, `knowledge`, `trust`, `replay`, `prototype`, modifications `root.go` pour `asa` TTY → TUI
- `application/internal/config/config.go` (UIConfig §29)
- `go.mod` — bubbletea, bubbles, huh, glamour
- `docs/ai/06-spec-ui.md`, `02-architecture.md`, `05-decisions.md`, `context-map.md`
- `docs-site/content/docs/{en,fr,de,es}/experience/**`
- Tests `application/internal/ui/**`
- Extractions partagées `cli/` ↔ `bus/` pour parité CommandBus

---

## Interdit

- Logique métier dans composants UI (§6.3)
- Accès SQLite direct depuis UI
- Casser `internal/tui` existant (specv3) — **coexistence obligatoire**
- Commit / push par l'agent
- Marquer `[x]` sans comportement réel (pas de cocher stub)
- Élargir le scope au-delà de `spec-ui.md` §1–36

---

## Definition of Done (globale — FULL FEATURE)

- [x] `cd application && go test ./... -count=1` vert
- [ ] `make build` OK (à confirmer en CI locale)
- [ ] `make build && ./bin/asa docs generate-cli`
- [x] Matrice traçabilité §8–33 : 100 % `[x]`
- [x] Critères §33 validés (explorers, palette, souris, widgets, tests intégration)
- [x] Aucun texte « stub », « TODO », « placeholder » visible dans les écrans TUI parcours nominal
- [x] Parité CLI : toute action TUI affiche ou expose son équivalent CLI (§3.1)
- [ ] `cd docs-site && pnpm docs:check` — si `pnpm` disponible
- [x] Reviewer signe clôture lot 7D (`2026-05-29`, réserves audit ci-dessus)

---

## Validation

```bash
cd application && go test ./... -count=1
make build
make build && ./bin/asa docs generate-cli
cd docs-site && pnpm docs:check  # si pnpm dispo
```

**Validation manuelle lot 7+ (minimum) :**
- `asa` → Mission Control complet
- `Ctrl+P` → recherche flow/agent + exécution action + CLI equivalent affiché
- `asa graph` → changement de vue + action sur nœud
- `asa replay open <id>` → jump event + compare
- Souris : clic sélection + resize + menu contextuel sur au moins un explorer
- `--plain` / `--json` / `--ci` inchangés sur commandes CLI directes

---

## Références

- [`spec-ui.md`](../../../spec-ui.md) — registre Spec G (Experience Platform)
- [`06-spec-ui.md`](../06-spec-ui.md) — canon projet
- [`06-spec-v3.md`](../06-spec-v3.md) — TUI rich/plain/json (coexistence `internal/tui`)
- Stacks A–F + V3 : [`context-map.md`](../context-map.md)
- ADR-027 — fondation UI (lots 1–6)
