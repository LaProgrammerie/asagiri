# Project workflows

Operational doc for the team: more detail than the root [`README.md`](../../README.md).  
**Generic** procedures (review, release, debug, plan, refactor) stay in **`~/.kiro/skills/`** so they are not duplicated in every repo.

---

## Day to day: spec → handoff → implementation

1. **Specify (Kiro)**  
   Work in `.kiro/specs/<feature>/` (requirements → design → tasks per your process).

2. **Summarize**  
   Update `docs/ai/active/current-spec.md` whenever scope or acceptance criteria change in a **material** way.

3. **Frame execution**  
   Update `docs/ai/active/handoff.md` before a Cursor session or implementation PR (scope, files, plan, tests, DoD).  
   Use the `.kiro/skills/create-handoff/` skill if you want a consistent handoff.

4. **Implement**  
   Follow the **handoff first**, then `03-standards.md` and useful sections of `02-architecture.md`.

5. **Close the loop**  
   If code reveals a spec gap: fix the **Kiro spec and projections** (`current-spec`, `handoff`) first, not only the code.

---

## What to update when (operational reminder)

| Change type | Action |
|-------------|--------|
| Idea / business need that shifts product target | `01-product.md` |
| New module, dependency, or moved boundary | `02-architecture.md` + `05-decisions.md` if durable |
| New command, linter, or test policy | `03-standards.md` |
| New team convention (branches, release, review) | This file (`04-workflows.md`) + `05-decisions.md` if structural |
| Tasks or design of the active spec | `.kiro/specs/...` → then `current-spec.md` → then `handoff.md` if needed |
| Only the next coding session changes | `handoff.md` (and spec if scope was wrong) |

Simple rule: **if another human (or agent) could get scope wrong tomorrow, update the canonical file today.**

---

## Local development

*(Fill in: install, dev server, env vars — point to `03-standards.md` for commands.)*

## Review and merge

*(Branches, PRs, acceptance criteria.)*  
Structured review: `code-review` skill in `~/.kiro/skills/`.

## Release

*(Versioning, changelog, deploy.)*  
Checklist: `release-checklist` skill in `~/.kiro/skills/`.

## Incidents / production debug

*(Short runbooks.)*  
Diagnostics: `debugging` skill in `~/.kiro/skills/`.

---

## Maintaining the “context system”

**Tooling** changes (new Cursor rules, new Kiro steering file, hooks enabled):

1. Edit files under `.cursor/` or `.kiro/` as needed.
2. If it changes a **durable** team rule: add a line to `05-decisions.md` + update `context-map.md` or this file if the human workflow changes.

For “what each folder is for”, see [`context-map.md`](context-map.md) and the [README](../../README.md).
