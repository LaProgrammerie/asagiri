# Kiro hooks (examples)

Kiro can react to events (files, agent, tools, **spec tasks**, etc.). Exact syntax depends on your version.

## Target uses (to wire up)

| Intent | Idea |
|--------|------|
| Spec / tasks | After task updates → refresh `docs/ai/active/handoff.md` (and `current-spec.md` if needed) |
| Design | End of design phase → consistency check with `docs/ai/02-architecture.md` |
| Sensitive task | Before execution (migrations, prod, secrets) → inject checklist / optional block |

See `hooks.config.example.yaml` for an **illustrative** shape; replace with the official Kiro format.
