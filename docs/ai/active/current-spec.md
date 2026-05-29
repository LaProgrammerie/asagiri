# Current spec — spec-ui livrée

**Phase :** [`spec-ui.md`](../../../spec-ui.md) — Asagiri Experience Platform — **livrée** (`2026-05-29`)  
**Handoff :** [`handoff.md`](handoff.md) — tranche spec-ui, matrice 100 %

## Spec active

- **Registre :** [`spec-ui.md`](../../../spec-ui.md) — Spec G — Asagiri Experience Platform (§1–36)
- **Canon :** [`06-spec-ui.md`](../06-spec-ui.md)
- **Handoff :** [`handoff.md`](handoff.md)

## Résumé — lots spec-ui

| Lot | Contenu |
|-----|---------|
| **1 — Foundation** | Bubble Tea app, router, layout, theme, CommandBus/QueryBus, config `ui:` étendu, `internal/ui/` Charm |
| **2 — Mission Control** | `asa` / `asa mission`, `asa dashboard`, widgets V1, event feed, live updates |
| **3 — Palette & nav** | Command Palette Ctrl+P, navigation raccourcis, safety UX confirmations |
| **4 — Explorers** | Graph, Flow, Knowledge, Trust, Explain, event feed intégré |
| **5 — Theatre & replay** | Agent Theatre (`asa agents watch`), Replay Explorer, Prototype Mode split view |
| **6 — Polish & doc** | Mouse resize panels (basic), no-animation mode, high-contrast usage, responsive compact/wide, écran d’aide accessibilité, canon `06-spec-ui.md`, docs-site `experience/` EN/FR/DE/ES, critères §33 |

## Principes clés

- **`asa`** (TTY) → Mission Control ; commandes CLI directes conservées
- TUI = **client du moteur** — aucune logique métier dans `internal/ui`
- Coexistence avec `internal/tui` (specv3 rich/plain/json)
- Toute action TUI expose son équivalent CLI (§3.1)

## Livraison

- **Spec-ui :** livré de bout en bout, y compris docs canon + docs-site 4 locales + clôture critères §33
- **Validation attendue dans handoff :** `go test ./... -count=1`, `make build`, `make build && ./bin/asa docs generate-cli`, `pnpm docs:check` (si disponible)

## Précédent livré

- **specv3** — Cost, Performance & Token Optimization — [`06-spec-v3.md`](../06-spec-v3.md) — livrée `2026-05-29`
- Stacks A–F + PF — voir [`context-map.md`](../context-map.md)

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.
