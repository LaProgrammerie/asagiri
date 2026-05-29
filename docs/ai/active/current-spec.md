# Current spec — spec-ui FULL FEATURE

**Phase :** [`spec-ui.md`](../../../spec-ui.md) — Asagiri Experience Platform — **FULL FEATURE livré** (`2026-05-29`, audit reviewer)  
**Handoff :** [`handoff.md`](handoff.md) — lots 7A–7D clôturés ; réserves P1/P2 documentées dans l'audit

## Spec active

- **Registre :** [`spec-ui.md`](../../../spec-ui.md) — Spec G — Asagiri Experience Platform (§1–36)
- **Canon :** [`06-spec-ui.md`](../06-spec-ui.md)
- **Handoff :** [`handoff.md`](handoff.md)

## Lot 7D livré

| Domaine | Contenu |
|---------|---------|
| **Mission Control §11** | Actions recommandées (runtime/trust/queue), coût jour/mois, flows critiques |
| **Prototype §19** | Pipeline TUI : `prototype create`, `flows extract`, `contracts extract`, `spec generate-from-product` |
| **Explain §21** | Questions typées, `FocusContext` → QueryBus, drill-down depuis explorers |
| **Souris §10.2** | Double-clic, hover, menu contextuel, sélection, resize |
| **Tests §32** | Intégration TTY `integration_test.go` |
| **Docs §31** | `experience/` 4 locales + `06-spec-ui.md` |

## Validation

```bash
cd application && go test ./... -count=1
make build
make build && ./bin/asa docs generate-cli
cd docs-site && pnpm docs:check  # si pnpm dispo
```

## Précédent

- **V1 lots 1–6** — shell + navigation + golden tests — livré `2026-05-29`
- **specv3** — [`06-spec-v3.md`](../06-spec-v3.md) — livrée `2026-05-29`

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.
