# project-onboarding — Tasks

## Lot 1 — Core CLI

- [x] **OB-1.1** Scaffold `internal/onboarding/` + detector registry
- [x] **OB-1.2** Implement Go detector (`go.mod`, default validation)
- [x] **OB-1.3** Implement Castor/PHP detector (`castor.php`, `castor qa:*`)
- [x] **OB-1.4** Implement Node detector (`package.json` scripts)
- [x] **OB-1.5** Config writer with merge policy + backup
- [x] **OB-1.6** Wizard step machine (interactive + `--yes --non-interactive`)
- [x] **OB-1.7** CLI `asa onboard` with flags (`--dry-run`, `--check-only`, `--stack`, `--step`)
- [x] **OB-1.8** CLI `asa ready` + `--json`/`--plain`/`--strict`
- [x] **OB-1.9** Extend `asa doctor --full` (gitignore, agents, docs, kiro spec, macOS asa conflict)
- [x] **OB-1.10** Docs bootstrap (feature skeleton, 01-product, handoff stub, guards)
- [x] **OB-1.11** Unit + integration tests
- [x] **OB-1.12** Update `06-spec-onboarding.md`, ADR-028, docs-site stub (en)

## Lot 2 — TUI

- [x] **OB-2.1** `GetReadinessQuery` / onboarding commands on bus
- [x] **OB-2.2** Screen `onboarding/` + `asa onboard --ui`
- [x] **OB-2.3** Mission Control readiness banner + palette entries
- [x] **OB-2.4** Golden + integration tests UI

## Lot 3 — Polish

- [x] **OB-3.1** `--resume` state persistence
- [x] **OB-3.2** `asa docs generate-cli` for onboard/ready
- [x] **OB-3.3** docs-site 4 locales (onboarding guide)
- [x] **OB-3.4** Example `examples/onboarding/` walkthrough

## Validation (each lot)

```bash
cd application && go test ./internal/onboarding/... ./internal/cli/... -count=1
make build
./bin/asa onboard --help
./bin/asa ready --json
```

**Statut global :** livré `2026-05-29`
