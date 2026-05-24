---
fileMatch:
  - "**/api/**"
  - "**/migrations/**"
  - "**/infra/**"
---

# Sensitive boundaries (example)

## API

- Stable contracts: backward-compatible evolution or explicit versioning.
- Auth / authz: do not bypass; align with `docs/ai/02-architecture.md` + `03-standards.md`.

## Migrations

- Reversibility or documented rollback plan; no destructive migration without an intermediate step if production data exists.

## Infra

- No secrets in plain text; configuration via the project’s intended mechanism.
