# Current spec — Asagiri Trust & Verification Engine (spec-my-B)

**Phase :** `spec-my-B` — **livrée** (`2026-05-29`)  
**Audit clôture :** [`handoff.md`](handoff.md) — lots 1–6 [x], DoD global §26 [x], validation 2026-05-29

## Spec active (clôturée)

- **Mission :** [`spec-my-B.md`](../../../spec-my-B.md) (§1–29, critères §26)
- **Canon :** [`06-spec-my-b.md`](../06-spec-my-b.md), ADR-020, ADR-021
- **Handoff :** [`handoff.md`](handoff.md) — tous lots [x]

## Résumé spec-my-B

Couche **Trust & Verification** local-first :

```text
implementation → trust checks → confidence (6 dimensions) → gates → review
```

| Lot | Focus | Statut |
|-----|--------|--------|
| **1** | Modèles, TrustEngine, rapports `.asagiri/trust/<id>/` | livré |
| **2** | Pipeline Verify, checks de base, confidence | livré |
| **3** | Checks avancés, blast radius | livré |
| **4** | CLI, gates YAML, UX terminal | livré |
| **5** | Runtime events, replay, `--strict-trust` | livré |
| **6** | Canon + site EN/FR/DE/ES | livré |

**CLI :** `asa verify trust`, `asa trust gates`, `asa trust replay`, `asa work --strict-trust`  
**Docs site :** `concepts/trust-engine`, `cli/verify-trust`, `cli/trust-gates`, `cli/trust-replay`, config `verification` ; EN `cli/generated/trust`, `verify-trust`, `trust-gates`, `trust-replay`

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.

## Suite possible

- [`spec-phase-finale.md`](../../../spec-phase-finale.md) (embeddings Ollama + npm) — hors scope spec-my-B

## Prérequis livré — spec-my-A

**Phase :** `spec-my-A` — **livrée** (`2026-05-27`)

1. **Executable Product Layer** — `asa prototype`, `asa flows`, `asa contracts`, `asa spec generate-from-product`
2. **Business Intent** — `business.yaml`, `asa flows review`, `asa architecture derive`
3. **Runtime persistant** — daemon/worker, sessions, API, SDK Go/TS
4. **Analysis layer** — `asa analysis build` → graphes sous `.asagiri/analysis/<product>/`
5. **Investigation** — rapport, `context-pack.md`, `asa investigate impact`

- **Spec :** [`spec-my-A.md`](../../../spec-my-A.md)
- **Canon :** [`06-spec-my-a.md`](../06-spec-my-a.md), ADR-018, ADR-019

## Previous phases

Voir ci-dessous (spec-better-flow, spec-release, spec-rename, spec-postv123).

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
