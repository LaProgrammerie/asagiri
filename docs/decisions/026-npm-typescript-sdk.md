# ADR-026 — npm TypeScript SDK distribution (PF-A-02)

**Date :** 2026-05-29  
**Status :** accepted  
**Spec :** [`spec-phase-finale.md`](../../spec-phase-finale.md) PF-A-02

## Context

The runtime REST API is consumed from TypeScript via `sdk/typescript/`, but consumers need a published package and automated releases decoupled from the Go binary version.

## Decision

1. Publish **`@laprogrammerie/asagiri`** from `sdk/typescript/` with semver independent of `asa` releases.
2. CI workflow **`.github/workflows/sdk-npm-publish.yml`** on tags `sdk-v*` and `workflow_dispatch`; requires repository secret **`NPM_TOKEN`**.
3. Package ships compiled **`dist/`** only (`prepublishOnly`: build + test).
4. Consumer docs under **`docs-site/.../reference/typescript-sdk`** (EN, FR, DE, ES).

## Consequences

- Go releases (`v*`) and SDK releases (`sdk-v*`) are separate tags.
- Forks without `NPM_TOKEN` can still build the SDK locally.

## Related

- [`sdk/typescript/README.md`](../../sdk/typescript/README.md)
- ADR-019 (runtime API foundation)
