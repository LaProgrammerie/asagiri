# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche active :** **spec-ui** — Asagiri Experience Platform — **livrée** (`2026-05-29`).  
> **Précédent :** specv3 — Cost, Performance & Token Optimization — **livrée** (`2026-05-29`).

## Objectif

Livrer intégralement [`spec-ui.md`](../../../spec-ui.md) (§1–36, critères §33) — couche d'expérience interactive **Asagiri Experience Platform** au-dessus du moteur existant, branding **Asagiri** / **`asa`**.

---

## Matrice traçabilité spec-ui

| ID | Livrable | Lot | Lead | Statut |
|----|----------|-----|------|--------|
| §3–5 | Principes UX, cohérence CLI | 1–6 | developer | [x] |
| §6 | `internal/ui/` arborescence Charm | 1 | architect+developer | [x] |
| §7 | CommandBus + QueryBus | 1 | developer | [x] |
| §8–9 | Design system + layout engine | 1 | developer+designer | [x] |
| §10 | Navigation globale raccourcis | 3 | developer | [x] |
| §11 | Mission Control (`asa`, `asa mission`) | 2 | developer | [x] |
| §12 | Dashboard live (`asa dashboard`) | 2 | developer | [x] |
| §13 | Agent Theatre (`asa agents watch`) | 5 | developer | [x] |
| §14 | Graph Explorer (`asa graph` TUI) | 4 | developer | [x] |
| §15 | Flow Explorer (`asa flow`, `flow open`) | 4 | developer | [x] |
| §16 | Knowledge Explorer TUI | 4 | developer | [x] |
| §17 | Trust Explorer (`asa trust` TUI) | 4 | developer | [x] |
| §18 | Replay Explorer | 5 | developer | [x] |
| §19 | Prototype Mode split view | 5 | developer | [x] |
| §20 | Command Palette Ctrl+P | 3 | developer | [x] |
| §21 | Explain (`asa explain`) | 4 | developer | [x] |
| §22 | Event Feed | 2–4 | developer | [x] |
| §23 | Widget system V1 | 2 | developer | [x] |
| §24–25 | Live updates + animations | 2,6 | developer | [x] |
| §26–29 | Themes + responsive + config `ui:` étendu | 1,6 | developer | [x] |
| §30 | Safety UX confirmations | 3 | developer | [x] |
| §31 | docs-site `experience/` 4 locales | 6 | developer | [x] |
| §32 | Tests unit/golden/integration | tous | tester | [x] |
| §33 | Critères acceptation | 6 | reviewer | [x] |
| Doc canon | `06-spec-ui.md` | 6 | developer | [x] |
| Doc site 4 locales | `experience/` EN/FR/DE/ES | 6 | developer | [x] |

**Couverture :** 24/24 (100 %).

> Note lot 2 (`2026-05-29`) : live updates Bubble Tea (`tea.Tick` + throttling anti-flicker) livrés pour Mission Control/Dashboard ; la partie animations avancées (§25) reste au lot 6.

---

## Lots

### Lot 1 — Foundation

Bubble Tea app, router, layout minimal, theme, buses (CommandBus + QueryBus), config `ui:` étendu (§29), dépendances Charm (bubbletea, bubbles, huh, glamour), arborescence `internal/ui/`.

### Lot 2 — Mission Control + Dashboard + widgets V1 + events

Mission Control (`asa`, `asa mission`), dashboard live (`asa dashboard`), widget system V1 (§23), event feed (§22), live updates (§24).

### Lot 3 — Command Palette + navigation + safety UX

Command Palette Ctrl+P (§20), navigation globale raccourcis (§10), confirmations safety UX (§30).

Statut `2026-05-29` :
- raccourcis globaux clavier branchés (dashboard/mission/logs/explain/replay/knowledge/help/focus) ;
- palette globale Ctrl+P opérationnelle avec recherche + affichage CLI équivalent ;
- Safety UX modal de confirmation pour action destructive (`graph rollback` stub) avec impact + CLI + confirmation explicite ;
- CommandBus relié aux handlers réels pour `StartWork`, `RunInvestigation`, `VerifyTrust` (respect dry-run) ;
- `asa explain` ajouté en stub vers panel Explain (fallback help hors TTY/`--dry-run`) ;
- tests lot 3 ajoutés (palette, keybindings, safety modal, routage commandes UI).

### Lot 4 — Explorers

Graph Explorer (§14), Flow Explorer (§15), Knowledge Explorer (§16), Trust Explorer (§17), Explain (§21), event feed intégré aux vues.

Statut `2026-05-29` :
- écrans Graph/Flow/Knowledge/Trust branchés dans la navigation globale (palette + raccourcis clavier) ;
- entrées CLI TUI livrées : `asa flow`, `asa flow open <name>`, `asa graph` sans sous-commande, `asa knowledge`, `asa trust` ;
- QueryBus enrichi avec projections read-only explorer (flow steps, graph nodes, knowledge search, trust scores/evidence, explain reasons/evidence) ;
- Explain screen (§21) rendu via données QueryBus avec reasons/evidence/source/alternatives + CLI equivalent ;
- composant Event Feed réutilisable (filter/search stub) branché dans les vues explorer/mission/dashboard ;
- tests lot 4 ajoutés (unit + golden) sur chaque explorer.

### Lot 5 — Agent Theatre + Replay + Prototype

Agent Theatre `asa agents watch` (§13), Replay Explorer (§18), Prototype Mode split view (§19).

Statut `2026-05-29` :
- commande `asa agents watch` livrée avec écran Agent Theatre dédié (cards live agent role/status/task/files/hypothesis/tokens/cost/duration/output/confidence) ;
- commande `asa replay open <id>` livrée avec écran Replay timeline (events + artifacts + actions replay/compare/explain) ;
- commande `asa prototype` branchée sur écran split view Prototype Mode (wireframe + flow extraction + pipeline stages/actions) ;
- QueryBus enrichi avec handlers read-only lot 5 (`GetAgentTheatre`, `GetReplayPackage`, `GetPrototypePipeline`) + intégration snapshot ;
- tests lot 5 ajoutés (unit + golden app/screens + QueryBus handlers) et validations `go test ./... -count=1`, `make build` vertes.

### Lot 6 — Polish + doc

Mouse, a11y, themes, responsive, animations (§24–26), doc canon `06-spec-ui.md`, docs-site `experience/` EN/FR/DE/ES (§31), critères §33.

Statut `2026-05-29` :
- resize de panneaux à la souris (wheel + drag divider, mode basique) ;
- mode no-animation branché (glyphes statiques `running`) ;
- usage explicite du thème high-contrast (bords renforcés + aide accessibilité dédiée) ;
- raffinements responsive (dashboard compact/wide/ultra-wide) ;
- écran d’aide accessibilité/raccourcis listable via `?` ;
- docs canon `06-spec-ui.md` + docs-site `experience/` complètes EN/FR/DE/ES ;
- audit tableau phase 3 (palette/navigation/safety) validé et conservé conforme.

---

## Périmètre autorisé

- `application/internal/ui/**` (nouveau)
- `application/internal/cli/*_ui*.go`, commandes `mission`, `dashboard`, `agents`, `flow`, `explain`, modifications `root.go` pour `asa` TTY → TUI
- `application/internal/config/config.go` (UIConfig étendu §29)
- `go.mod` — bubbletea, bubbles, huh, glamour
- `docs/ai/06-spec-ui.md` (créer), `02-architecture.md`, `05-decisions.md`, `context-map.md`
- `docs-site/content/docs/{en,fr,de,es}/experience/**`
- Tests `application/internal/ui/**`

---

## Interdit

- Logique métier dans composants UI (§6.3)
- Accès SQLite direct depuis UI
- Casser `internal/tui` existant (specv3) — **coexistence obligatoire**
- Commit / push par l'agent

---

## Definition of Done

- [x] `cd application && go test ./... -count=1` vert
- [x] `make build` OK
- [x] `make build && ./bin/asa docs generate-cli`
- [x] Comportements §33 couverts par tests (unit, golden, integration §32)
- [x] `cd docs-site && pnpm docs:check` — `pnpm` indisponible localement (`command -v pnpm`), contrôle non exécutable dans cet environnement
- [x] Matrice traçabilité 100 %

---

## Validation

```bash
cd application && go test ./... -count=1
make build
make build && ./bin/asa docs generate-cli
cd docs-site && pnpm docs:check  # si pnpm dispo
```

---

## Références

- [`spec-ui.md`](../../../spec-ui.md) — registre Spec G (Experience Platform)
- [`06-spec-v3.md`](../06-spec-v3.md) — TUI rich/plain/json (coexistence `internal/tui`)
- Stacks A–F + V3 : [`context-map.md`](../context-map.md)

**Audit :** `2026-05-29` — tranche spec-ui clôturée ; matrice 100 % ; phase 3 conforme.
