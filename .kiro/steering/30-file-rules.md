# Steering split by file (project)

## Important

Do **not** rely on a single `fileMatch: "**/*"` for the whole repo: that removes the benefit of targeted context. Split into **separate files** with **narrow** globs.

## Pattern

- One `.kiro/steering/` file = one or a few coherent zones (`fileMatch`) + short rules.
- Rename `31-`, `32-`… to match your real layout (`src/`, `app/`, `packages/`, etc.).

## Example files shipped

| File | Zones (example) |
|------|-----------------|
| `31-stack.md` | `**/*.{ts,tsx}` (replace with your stack) |
| `32-boundaries.md` | `**/api/**`, `**/migrations/**`, `infra/**` |

Delete or rename these examples if your paths differ.
