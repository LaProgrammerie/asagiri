# Current spec — spec-my-F livrée

**Phase :** [`spec-my-F.md`](../../../spec-my-F.md) — **livrée** (`2026-05-29`)  
**Handoff :** [`handoff.md`](handoff.md) — tranche spec-my-F, matrice F-* 100 %

## Spec active

- **Registre :** [`spec-my-F.md`](../../../spec-my-F.md) — Replay & Deterministic Execution
- **Canon :** [`06-spec-my-f.md`](../06-spec-my-f.md)
- **Handoff :** [`handoff.md`](handoff.md)

## Résumé — livrables F

| Bloc | Contenu |
|------|---------|
| **Replay package** | `.asagiri/replays/<id>/`, `replay.yaml`, capture graph/trust/investigation/handoffs |
| **CLI** | `asa replay create|run|compare|explain|snapshot` |
| **Modes** | full, simulation, offline, audit, compare ; `--strict` |
| **Compare** | Coût, trust diff, divergences graph/artefacts |
| **Config** | Bloc `replay:` dans `config.yaml` |
| **Docs** | Site 4 locales + docgen `en/cli/generated/replay*` |

## Stack A–F (référence)

| Spec | Statut |
|------|--------|
| A + PF-A | Livré |
| B | Livré |
| C + PF-C P1 | Livré |
| D + D-FULL | Livré |
| E | Livré |
| **F** | **Livré** |

Canon : [`06-spec-my-a.md`](../06-spec-my-a.md) … [`06-spec-my-f.md`](../06-spec-my-f.md).

## Reliquats ouverts (hors F)

- **PF-C-06** — inférence dépendances V2 (P2)
- **PF-X-02** — tokenizers cost exacts (P3)
- Voir [`problems.md`](../../../problems.md) (GAP ↔ PF-X)

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.
