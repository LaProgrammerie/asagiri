# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche :** **spec-my-F** — livrée (`2026-05-29`).  
> **Clôture reliquats :** phase finale PF-* — **livrée** (`2026-05-29`, lot 7 doc).  
> **Précédent :** phase finale P1 + stack FULL A–E ; [`spec-my-E`](#tranche-spec-my-e--livrée) livrée `2026-05-29`.

## Objectif

Livrer le **Replay & Deterministic Execution System** ([`spec-my-F.md`](../../../spec-my-F.md)) : packages `.asagiri/replays/`, moteur `internal/replay/`, CLI `asa replay`, modes offline/simulation, comparaison et divergences, intégrations investigation / trust / execution graph / coordination.

---

## Stack A–F (état code `2026-05-29`)

| Spec | Statut | Notes |
|------|--------|-------|
| **A** + PF-A | Livré | Product, runtime, embedder, SDK npm |
| **B** | Livré | Trust engine ; `trust replay` ≠ `asa replay` |
| **C** + PF-C | Livré | Execution graph, checkpoints, trust runner, inférence V2 |
| **D** + D-FULL | Livré | Coordination, worktrees, `NodeExecutor` |
| **E** | Livré | Knowledge graph |
| **F** | **Livré** | Replay packages, compare, snapshots, redaction |

**Code :** `application/internal/replay/`, CLI `replay_*`, `.asagiri/replays/`.

---

## Tranche spec-my-F — livrée

### Prérequis livrés

- [`spec-my-A.md`](../../../spec-my-A.md) … [`spec-my-E.md`](../../../spec-my-E.md) — stacks A–E + phase finale P1
- Canon : [`06-spec-my-f.md`](../06-spec-my-f.md)

### Périmètre autorisé (spec-my-F)

- `application/internal/replay/**`
- `application/internal/cli/replay_cmd.go`, `replay_cmd_test.go`
- `application/internal/config/config.go` (bloc `ReplayConfig` uniquement si nécessaire)
- `.asagiri/replays/`, `application/internal/replay/testdata/**`
- `docs/ai/**`, `docs-site/content/docs/{en,fr,de,es}/**`
- `.asagiri/config.yaml.example` (bloc `replay:`)

### Lots — Definition of Done (spec-my-F §29)

#### Lot 1 — Replay package

- [x] Format `replay.yaml`, capture artefacts, restore minimal, `go test ./internal/replay/...`

#### Lot 2 — Runtime capture

- [x] Runtime events, graph state, checkpoints copiés sous `graph/`, `runtime/`

#### Lot 3 — Comparison engine

- [x] `replay compare`, trust diff, `DetectDivergences`, `replay explain`

#### Lot 4 — Offline / simulation modes

- [x] `--offline`, `--simulation`, `--strict`, policies config

#### Lot 5 — Integrations

- [x] Investigation (`investigations/`), trust (`trust/`), execution graph, coordination handoffs

#### Lot 6 — Provenance, snapshots, UX, documentation

- [x] Provenance index, `replay snapshot`, UX terminal §26, canon + site 4 locales, docgen CLI

### Matrice traçabilité spec-my-F

| ID | Lot | Statut |
|----|-----|--------|
| F-PKG-1 | 1 | [x] |
| F-CAP-1 | 1 | [x] |
| F-MANIFEST-1 | 1 | [x] |
| F-RT-1 | 2 | [x] |
| F-GRAPH-CP-1 | 2 | [x] |
| F-CMP-1 | 3 | [x] |
| F-DIV-1 | 3 | [x] |
| F-EXPLAIN-1 | 3 | [x] |
| F-OFFLINE-1 | 4 | [x] |
| F-SIM-1 | 4 | [x] |
| F-STRICT-1 | 4 | [x] |
| F-INT-INV-1 | 5 | [x] |
| F-INT-TRUST-1 | 5 | [x] |
| F-INT-EG-1 | 5 | [x] |
| F-INT-COORD-1 | 5 | [x] |
| F-PROV-1 | 6 | [x] |
| F-SNAP-1 | 6 | [x] |
| F-UX-1 | 6 | [x] |
| F-DOC-1 | 6 | [x] |
| F-JSON-1 | 6 | [x] |

**Couverture F-* :** 20/20 (100 %).

### DoD global spec-my-F (§28)

- [x] Tous critères §28 (create, run, offline, compare, divergences, trust/graph/handoffs, redaction, tests)

### Validation globale spec-my-F

```bash
cd application && go test ./internal/replay/... -count=1
make build && ./bin/asa docs generate-cli
./bin/asa replay create --from-graph graph-2026-05-29-test0001 --include-runtime --include-events
./bin/asa replay run replay-2026-05-29-<id> --offline --dry-run
./bin/asa replay compare replay-a replay-b --json
./bin/asa replay explain replay-a replay-b
./bin/asa replay snapshot replay-2026-05-29-<id> --name smoke
```

---

## Hors scope (spec-my-F)

- Déterminisme parfait des LLM ; replay d’APIs externes live
- Commit / push par l’agent
- Nouvelles specs post-F sans mise à jour registre

---

## Archive — phase finale (clôture reliquats)

| ID | Livrable | Statut |
|----|----------|--------|
| PF-A-01 | Embedder + `memory doctor` | [x] |
| PF-A-02 | SDK npm | [x] |
| PF-C-01 … PF-C-05 | Graph durcissement | [x] |
| PF-C-06 | Inférence dépendances V2 | [x] |
| PF-X-01 | `resume --execute` boucle complète | [x] |
| PF-X-02 | Tokenizers cost | [x] |
| PF-X-03 | Index RAG sémantique | [x] |
| PF-X-04 | Docgen Cobra | [x] |
| D-WT-1 … D-RUN-1 | D-FULL | [x] |

**Couverture PF-* :** 12/12 (100 %).

Validation archive :

```bash
cd application && go test ./internal/memory/... ./internal/executiongraph/... ./internal/cost/... -count=1
./bin/asa memory doctor
./bin/asa memory reindex
./bin/asa index
./bin/asa index search "checkpoint resume"
./bin/asa resume <run-id> --execute --max-steps 20
./bin/asa graph run minimal-product --flow workspace-onboarding --checkpoint-every node --dry-run
make build && ./bin/asa docs generate-cli
```

---

## Tranche spec-my-E — livrée

Lots 1–6 et matrice E-* : **livrés** `2026-05-29`. Détail : [`06-spec-my-e.md`](../06-spec-my-e.md).

```bash
cd application && go test ./internal/knowledge/... -count=1
./bin/asa knowledge build --include-flows --include-code
```

---

## Références

- [`spec-my-F.md`](../../../spec-my-F.md), [`06-spec-my-f.md`](../06-spec-my-f.md)
- [`spec-phase-finale.md`](../../../spec-phase-finale.md), ADR-025/026
- [`06-spec-my-e.md`](../06-spec-my-e.md), ADR-024
- [`problems.md`](../../../problems.md)

**Audit :** `2026-05-29` — spec-my-F alignée code ; regénérer CLI EN : `make build && ./bin/asa docs generate-cli`.
