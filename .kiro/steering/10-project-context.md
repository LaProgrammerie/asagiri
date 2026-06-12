# Kiro steering — project context

## Role

Point to the **canonical memory outside** Kiro and restate the spec → projections flow in `docs/ai/active/`.

## Required references

- Agent entry: `AGENTS.md`
- Context map: `docs/ai/context-map.md`
- Detailed canon: `docs/ai/`
- Active spec (summary): `docs/ai/active/current-spec.md`
- Handoff: `docs/ai/active/handoff.md`

## Expected flow

1. Produce or update artefacts under `.kiro/specs/<feature>/` (requirements, design, tasks).
2. Update `docs/ai/active/current-spec.md` (cross-tool summary).
3. Generate or refresh `docs/ai/active/handoff.md` for implementation (Cursor / Copilot).
4. If a structural decision: `docs/ai/05-decisions.md`.

*(Kiro hooks can automate part of this — see `.kiro/hooks/README.md`.)*
