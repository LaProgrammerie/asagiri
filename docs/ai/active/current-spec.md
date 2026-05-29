# Current spec — Asagiri Multi-Agent Coordination System (spec-my-D)

**Phase :** `spec-my-D` — **en cours**  
**Handoff :** [`handoff.md`](handoff.md) — lots 1–6 [ ], matrice D-* [ ]

## Spec active

- **Mission :** [`spec-my-D.md`](../../../spec-my-D.md) (§1–22, critères §20)
- **Handoff :** [`handoff.md`](handoff.md)

## Résumé vision

Couche **Multi-Agent Coordination** local-first — gouvernance explicite des agents au-dessus du graphe d'exécution :

```text
intent → execution graph → agent assignment → isolated execution
  → cross-agent validation → trust gates → merge / reject
```

| Lot | Focus | Statut |
|-----|--------|--------|
| **1** | Foundation §3–5, §18 — package `coordination/`, rôles, isolation, interfaces, config `coordination:` | [ ] |
| **2** | Assignment & pipeline §6–7, §12 — assignation multi-critères, pipelines rejouables, context reduction | [ ] |
| **3** | Handoffs & validation §8–9, §11 — `.asagiri/handoffs/`, cross-validation, policies | [ ] |
| **4** | Budget, retry, conflict, merge §13–16 | [ ] |
| **5** | Runtime §10, §17, worktrees §5 — events `agent.*`, coordination graph, runner + trust | [ ] |
| **6** | UX terminal §19, acceptance §20, doc canon + site EN/FR/DE/ES | [ ] |

**Artefacts cibles :** `.asagiri/handoffs/<handoff-id>/`  
**Config :** bloc `coordination:` dans `.asagiri/config.yaml`  
**Packages :** `application/internal/coordination/` (§18)

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.

## Prérequis livrés

### spec-my-C — Execution Graph Planner

**Phase :** `spec-my-C` — **livrée** (`2026-05-29`)

Planification orchestrée par graphe de dépendances : planner, scheduler, runner, checkpoints, CLI `plan graph` / `graph run|status|resume|visualize`.

- **Spec :** [`spec-my-C.md`](../../../spec-my-C.md)
- **Canon :** [`06-spec-my-c.md`](../06-spec-my-c.md), ADR-022
- **Artefacts :** `.asagiri/graphs/<graph-id>/`

### spec-my-B — Trust & Verification Engine

**Phase :** `spec-my-B` — **livrée** (`2026-05-29`)

Couche trust local-first : checks → confidence (6 dimensions) → gates → review.

- **Spec :** [`spec-my-B.md`](../../../spec-my-B.md)
- **Canon :** [`06-spec-my-b.md`](../06-spec-my-b.md), ADR-020, ADR-021
- **CLI :** `asa verify trust`, `asa trust gates`, `asa trust replay`, `asa work --strict-trust`

### spec-my-A — Product & Runtime Layer

**Phase :** `spec-my-A` — **livrée** (`2026-05-27`)

1. **Executable Product Layer** — `asa prototype`, `asa flows`, `asa contracts`, `asa spec generate-from-product`
2. **Business Intent** — `business.yaml`, `asa flows review`, `asa architecture derive`
3. **Runtime persistant** — daemon/worker, sessions, API, SDK Go/TS
4. **Analysis layer** — `asa analysis build` → graphes sous `.asagiri/analysis/<product>/`
5. **Investigation** — rapport, `context-pack.md`, `asa investigate impact`

- **Spec :** [`spec-my-A.md`](../../../spec-my-A.md)
- **Canon :** [`06-spec-my-a.md`](../06-spec-my-a.md), ADR-018, ADR-019

## Phase ultérieure — reliquats transverses

- [`spec-phase-finale.md`](../../../spec-phase-finale.md) — registre unique des écarts V1 / stubs après A, B, C, D (embeddings, npm, durcissements PF-*, reliquats CLI PF-X-*)

## Previous phases

Voir ci-dessous (spec-better-flow, spec-release, spec-rename, spec-postv123).

---

# Previous phase — spec-my-C (archive résumé)

**Livrée** `2026-05-29` — détail : [`06-spec-my-c.md`](../06-spec-my-c.md), handoff archivé dans historique git.

---

# Previous phase — spec-better-flow

Fusionnée dans spec-my-A (§23).

---

# Previous phase — release distribution (spec-release)

**Date :** 2026-05-17 — [`spec-release.md`](../../../spec-release.md), ADR-015

---

# Previous phase — Asagiri rebrand (spec-rename)

**Date :** 2026-05-20 — [`spec-rename.md`](../../../spec-rename.md), ADR-016

---

# Previous phase — consolidation (post-V3)

**Date :** 2026-05-17 — [`spec-postv123.md`](../../../spec-postv123.md)
