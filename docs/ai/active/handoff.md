# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche active :** **specv3** — Cost, Performance & Token Optimization — **livrée** (`2026-05-29`).  
> **Précédent :** spec-my-F + phase finale PF-* — livrées.

## Objectif

Livrer intégralement [`specv3.md`](../archives/specs/specv3.md) (§17 critères, phases 1–7, §16 CLI) sous branding **Asagiri** / **`asa`**.

---

## Matrice traçabilité specv3 (après livraison)

| ID / § | Livrable | Statut |
|--------|----------|--------|
| §3–6 | `cost/` estimation, pricing, budgets | [x] |
| §7 | `DurationModel` + historique SQLite | [x] |
| §8 | `contextopt/` reduce, pack, savings | [x] |
| §9 | `investigation/` + CLI `investigate` | [x] |
| §10 | `mcp serve`, tools `asagiri.*`, scope | [x] |
| §11 | `routing/` + intégration estimate | [x] |
| §12 | `work` pre-flight, flags V3 | [x] |
| §13 | `tui/` rich/plain/json | [x] |
| §14 | `run_metrics` / `step_metrics` | [x] |
| §15 | Rapport Cost & Performance | [x] |
| §16 | `estimate`, `context`, `cost`, `mcp` | [x] |
| §17 | Critères d'acceptation | [x] |
| Doc canon | `06-spec-v3.md`, ADR-010 | [x] |
| Doc site 4 locales | cost-performance, workflows, config | [x] existant |

**Couverture :** 14/14 (100 %).

---

## Périmètre autorisé (specv3)

- `application/internal/cost/**`, `contextopt/**`, `investigation/**`, `routing/**`, `telemetry/**`, `tui/**`, `mcp/**`, `pipeline/**`
- `application/internal/cli/{estimate,work,context,cost_cmd,mcp_cmd,v3_display}.go`
- `application/internal/report/**`, `store/sqlite/metrics.go`
- `docs/ai/06-spec-v3.md`, `02-architecture.md`, `05-decisions.md` (si besoin)
- `docs-site/content/docs/{en,fr,de,es}/**` (cost-performance, concepts, config, CLI généré)

---

## Lots livrés (`2026-05-29`)

### Lot 1 — Estimation & work pre-flight

- `RunV3PreFlight` / `RunV3Execute` ; estimation affichée **avant** exécution `work`
- `asa estimate`, `work --estimate-only`

### Lot 2 — Métriques & cost report

- Persistance `step_metrics` planifiés ; `cost report` (local/cloud %, savings)

### Lot 3 — Context optimization

- `ComputeOptimize` ; `context --optimize` ; `--no-context-reduction`

### Lot 4 — Routing & budgets

- `BuildOpts` PreferLocal/NoCloud ; confirmation budget

### Lot 5 — Rapport & doc

- `report.CostPerformance` + workflow ; `06-spec-v3.md` ; tests §17

---

## Definition of Done

- [x] `cd application && go test ./... -count=1` vert
- [x] `make build` OK
- [x] Comportements §17 couverts par tests (pipeline, contextopt, report, tui)
- [x] `make build && ./bin/asa docs generate-cli`
- [ ] `cd docs-site && pnpm docs:check` — exécuter si `pnpm` disponible localement

---

## Validation

```bash
cd application && go test ./... -count=1
make build && ./bin/asa docs generate-cli
./bin/asa work "develop agentflow-test" --dry-run --estimate-only
./bin/asa estimate agentflow-test --dry-run
./bin/asa context agentflow-test --optimize --dry-run
./bin/asa cost report --since 7d
```

---

## Hors scope

- Tokenizers provider-exacts hors PF-X-02 (approximation + tiktoken option conservée)
- Commit / push par l'agent

---

## Références

- [`specv3.md`](../archives/specs/specv3.md), [`06-spec-v3.md`](../06-spec-v3.md)
- Stacks A–F : [`06-spec-my-f.md`](../06-spec-my-f.md)

**Audit :** `2026-05-29` — specv3 alignée code ; prochaine spec au choix produit.
