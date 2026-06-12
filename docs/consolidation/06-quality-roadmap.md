# Tests & qualité

**Date :** 2026-05-17 · **Référence :** spec-postv123 §6

## Couverture (après consolidation)

| Package | Couverture | Cible |
|---------|------------|-------|
| `workflow` | ~24 % | >50 % |
| `intent` | ~35 % | >50 % |
| `pipeline` | **~52 %** | >50 % ✓ |
| `cost` | **~58 %** | >50 % ✓ |
| `routing` | **~78 %** | >50 % ✓ |
| `contextopt` | ~17 % | >40 % |
| `telemetry` | faible | tests via sqlite store |
| `cli` | ~47 % | maintenir |

`go test -race ./...` : **vert**.

## Tests ajoutés (cette mission)

- `pipeline/v3_test.go` — estimate-only
- `routing/router_test.go`
- `intent/executor_test.go`
- `contextopt/collector_pipeline_test.go`
- `cost/estimator_golden_test.go`
- `redact`, `workflow` status, `config` MCP validate

## Manquants critiques

1. Tests intégration `Executor` + `workflow.Service` (dev dry-run bout en bout)
2. Golden JSON plan `work --plan-only`
3. Tests MCP deny path / truncation
4. Tests corruption SQLite recovery
5. Tests `contextopt.Reduce` + compressor sur gros fichiers

## Roadmap qualité (90 jours)

| Sprint | Objectif |
|--------|----------|
| S1 | workflow + intent >50 % ; golden plans |
| S2 | contextopt >40 % ; benchmark CI artifact |
| S3 | integration tag Notion + MCP opt-in |
