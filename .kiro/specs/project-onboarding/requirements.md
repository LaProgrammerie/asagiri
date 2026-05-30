# project-onboarding — Requirements

## Problem

New or forked repos require manual `.asagiri/config.yaml` editing, placeholder `docs/ai/` files,
and ad-hoc prerequisite checks before `asa work` is useful. The Experience Platform appears to
promise guidance but only operates on an already-configured project.

## Goals

1. **Guided setup** — `asa onboard` wizard (CLI first, TUI lot 2).
2. **Stack-aware config** — auto-detect Go / PHP Castor / Node and propose validation commands.
3. **Readiness gate** — `asa ready` with structured report and next actions.
4. **Extended doctor** — agents, gitignore, docs placeholders, Kiro spec presence.
5. **Minimal doc bootstrap** — first Kiro feature + `01-product.md` / handoff stub without overwriting real content.

## Users

- Developers forking the AI Engineering template (PHP docker-starter, Go template, etc.).
- Teams adopting Asagiri on an existing Git repo.

## Functional requirements

### FR-1 Detection

- FR-1.1 Detect `go.mod` → Go validation defaults.
- FR-1.2 Detect `castor.php` → Castor QA commands.
- FR-1.3 Detect `application/package.json` or root `package.json` → npm test/qa scripts.
- FR-1.4 Support `--stack` override and cumulative hints (e.g. PHP + Node frontend).

### FR-2 Config generation

- FR-2.1 Merge into existing `config.yaml` with safe defaults policy (see spec §6.3).
- FR-2.2 Set `project.name`, `worktrees.branch_prefix`, `validation.commands`, `specs.*`.
- FR-2.3 Validate written config via `config.Load` + `config.Validate`.
- FR-2.4 `--dry-run` shows diff without write; backup before overwrite.

### FR-3 Readiness

- FR-3.1 `asa ready` returns score, checks[], next_actions[].
- FR-3.2 `--json`, `--plain`, `--ci`, `--strict` output modes.
- FR-3.3 Persist last report to `.asagiri/onboarding/report.json`.

### FR-4 Doctor extension

- FR-4.1 Optional `--full` flag for onboarding-specific checks.
- FR-4.2 Warn on macOS `/usr/bin/asa` shadowing Asagiri binary.

### FR-5 Document bootstrap

- FR-5.1 Create `.kiro/specs/<feature>/` skeleton from wizard answer.
- FR-5.2 Fill `docs/ai/01-product.md`, `current-spec.md`, handoff stub if placeholders only.
- FR-5.3 Update `docs/ai/03-standards.md` command table from chosen validation.
- FR-5.4 Never overwrite substantiated files without `--force-docs`.

### FR-6 Idempotency & resume

- FR-6.1 Re-run onboard skips unchanged sections or prompts to update.
- FR-6.2 `--resume` from `.asagiri/onboarding/state.json`.

### FR-7 UI (lot 2)

- FR-7.1 Mission Control banner when not ready.
- FR-7.2 Onboarding screen + palette entry; CLI parity.

## Non-goals

- Installing cursor-agent / codex / kiro binaries.
- Generating business requirements automatically.
- Non-Git projects.

## Acceptance

See spec-onboarding §12 and handoff Definition of Done.
