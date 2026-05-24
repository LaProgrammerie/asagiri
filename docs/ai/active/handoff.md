# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation. Derived from tasks / design under `.kiro/specs/`.  
> Any action outside **allowed scope** or outside **files to change** is **out of contract** — update the spec and this file before expanding.  
> See `docs/ai/context-map.md`.

## Immediate objective

*(One sentence: single deliverable for this session / iteration.)*

## Allowed scope

- **In:** *(what must be done now)*
- **Explicitly out:** *(what must not be touched or “improved” on initiative)*

## Files to change

*(As precise as possible: paths or globs; “create” vs “edit”.)*

- …

## Files not to change

*(Lock what would derail architecture, public contract, or other teams.)*

- …

## Implementation plan

Split into **small**, **ordered**, **verifiable** steps (one step = observable outcome: file compiles, test green, manual behaviour reproducible).

1. …
2. …
3. …

## Tests to add / run

- **Commands:** *(e.g. `pnpm test`, file target — see `docs/ai/03-standards.md`)*
- **Cases to cover:** …
- **Out of test scope:** …

## Risks / invariants

- **Invariants:** *(do not break: API, schema, perf, security)*
- **Risks:** …

## Definition of Done

**Always** include at least (adapt wording to the repo):

- [ ] Planned code / change **implemented** within handoff scope
- [ ] Relevant **tests** passing (or explicit justification if untested area)
- [ ] **Docs / spec** synced if impact is not code-only (`docs/ai/*`, `.kiro/specs/`, `current-spec.md` as applicable)

*(Add task-specific criteria below.)*

- [ ] …

## Ambiguities / to resolve

*(Optional but better than guessing: one line per open point; if none, write “None”.)*

- …

## References

- `docs/ai/active/current-spec.md`
- Kiro spec: `.kiro/specs/<feature>/`
- Architecture (useful sections): `docs/ai/02-architecture.md`
