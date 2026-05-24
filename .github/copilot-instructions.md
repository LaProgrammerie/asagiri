# GitHub Copilot instructions (repository)

**Adapter:** same sources of truth as Kiro / Cursor without duplicating the full manual.

## Sources of truth

1. **Entry:** `AGENTS.md`
2. **Context map:** `docs/ai/context-map.md`
3. **Detail:** `docs/ai/*.md` (notably `02-architecture.md`, `03-standards.md`)
4. **Execution:** `docs/ai/active/handoff.md`

## Expected behaviour

- Respect invariants in `AGENTS.md` and the `.kiro/specs/` ↔ `current-spec.md` ↔ `handoff.md` alignment described in `context-map.md`.
- For coding: `handoff.md` is the **contract** (scope, allowed/denied files, DoD) — do not widen it without updating spec and handoff.
- Generic workflows (review, release, debug, plan): **`~/.kiro/skills/`** on the machine.
- **Handoff / spec → execution** workflow specific to this repo: `.kiro/skills/create-handoff/` if needed.
- Do not suggest committing secrets or sensitive files that are not gitignored.

**Stack (template) :** Docker + Castor (docker-starter), app PHP sous `application/` par défaut, Node possible dans le builder ou service dédié. Infra déployable : Yoimachi — `docs/ai/02-architecture.md` + `infra/yoimachi/`. Tests / commandes : `docs/ai/03-standards.md`.
