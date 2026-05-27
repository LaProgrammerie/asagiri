# Current spec — Asagiri Executable Product Layer (spec-my-A)

**Phase :** `spec-my-A` — **livrée**  
**Date :** 2026-05-27

## Spec active

- **Mission :** [`spec-my-A.md`](../../../spec-my-A.md) (§1–26)
- **Handoff :** [`handoff.md`](handoff.md)
- **Suite optionnelle :** [`spec-phase-finale.md`](../../../spec-phase-finale.md) (embeddings sémantiques Ollama + publish npm)

## Résumé livré

1. **Executable Product Layer** — `asa prototype`, `asa flows`, `asa contracts`, `asa spec generate-from-product`, compatibilité `asa work`
2. **Business Intent** — `business.yaml`, `asa flows review`, `asa architecture derive`, tasks flow-first
3. **Runtime persistant** — daemon/worker, sessions, branches, hooks, memory, skills, API HTTP + Unix socket, SDK Go/TS
4. **Analysis layer** — `asa analysis build` → 7 graphes sous `.asagiri/analysis/<product>/`
5. **Investigation** — rapport, `context-pack.md`, graph, `asa investigate impact`, intégration work/verify

Branding : **Asagiri** / **`asa`** / `github.com/LaProgrammerie/asagiri`.

## Documentation

- **Site docs** (EN / FR / DE / ES) : `cli/runtime`, `cli/runtime-serve`, `cli/investigate`, `cli/analysis`, `concepts/runtime`, `concepts/investigation`, `concepts/analysis-layer`, `reference/typescript-sdk`, bloc `runtime` dans `configuration/config-file`, CLI `generated/` regénéré
- **Canon IA** : [`06-spec-my-a.md`](../06-spec-my-a.md), handoff audité, [`context-map.md`](../context-map.md)
- ADR : 018 (spec-my-A), 019 (API + analysis)

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
