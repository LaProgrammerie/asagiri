---
name: create-handoff
description: >-
  Updates handoff.md using the mandatory template from .kiro/specs/; does not
  invent scope; surfaces ambiguities explicitly.
---

# Project skill — handoff from Kiro spec

## Goal

Produce a **normative handoff**: only the **immediate executable**, all sections of `docs/ai/active/handoff.md` filled, without copying the full spec.

## Entry criteria (required)

- `.kiro/specs/<feature>/` identified with at least `tasks.md` (or clear tasks elsewhere + explicit agreement).
- `docs/ai/active/current-spec.md` read; aligned with spec or **updated first** if scope changed.
- `docs/ai/02-architecture.md` read for boundaries and sensitive files.

## When to use

- After `tasks.md` or `design.md` changes
- Before an implementation session (Cursor / Copilot)

## Expected inputs

- Feature name / path `.kiro/specs/<feature>/`
- Any extra user constraints (deadline, prohibitions)

## Exclusion rules

- **Do not** paste full requirements/design into the handoff.
- **Do not** widen scope vs agreed tasks; if widening is needed, add an **“out of current scope”** note and list spec changes required.
- **Do not** leave `…` or empty sections: if unknown, fill **Ambiguities / to resolve** (below).

## Steps

1. Read `tasks.md`, then `design.md` / `requirements.md` if needed to resolve ambiguity.
2. If `current-spec.md` does not reflect state: **update it** (**State** fields, summary, scope).
3. Fill **every section** of `docs/ai/active/handoff.md`: Immediate objective, Allowed scope, Files to change / not to change, Plan, Tests, Risks, DoD, References.
4. List **exact files** to change when identifiable from design/tasks; otherwise give **area** + **discovery rule** (e.g. “all handlers under `src/api/foo/`”).
5. At end of handoff if needed:

```markdown
## Ambiguities / to resolve

- …
```

## Handling ambiguity

- Do not guess business invariants or architecture choices.
- **One line per ambiguity:** closed question + options + recommendation if one option is clearly safer.
- If the handoff cannot be safe: add **Definition of Done** items blocked on resolving those points.

## Mandatory output format

- `docs/ai/active/handoff.md` **fully** filled per template (no removed sections).
- If only proposing text: deliver **paste-ready markdown** for `handoff.md` + reminder to update `current-spec.md` if relevant.
