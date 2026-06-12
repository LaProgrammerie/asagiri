# ADR-025 — Pluggable memory embeddings (PF-A-01)

**Date :** 2026-05-29  
**Status :** accepted  
**Spec :** [`spec-phase-finale.md`](../ai/archives/specs/spec-phase-finale.md) PF-A-01

## Context

Runtime memory used deterministic bag-of-words vectors (`internal/embedutil`). Semantic retrieval for investigations and `asa memory list --query` needs pluggable embedders without mandating cloud APIs.

## Decision

1. Add **`application/internal/memory/embedder/`** with `Embedder` interface and implementations: **`hash`** (legacy), **`ollama`** (`POST /api/embeddings`), **`cloud`** (OpenAI-compatible, opt-in only).
2. Configure under **`runtime.memory`** in `.asagiri/config.yaml` (`embedder`, `ollama`, `cloud.enabled`).
3. Default embedder **`hash`** when unset (CI-safe); example config documents **`ollama`** for local-first setups.
4. **`asa memory reindex`** recomputes all stored vectors; retrieval uses cosine similarity on stored JSON embeddings.
5. Cloud embedder returns an error when `cloud.enabled: false`, even if `OPENAI_API_KEY` is set.

## Consequences

- Hash behaviour preserved for non-regression tests.
- Ollama required only when `embedder: ollama` is selected.
- Separate from ADR-020 (trust engine).

## Related

- [`spec-my-A.md`](../ai/archives/specs/spec-my-A.md) §24.10
- Config: [`docs-site` runtime.memory](../site) (EN/FR/DE/ES config-file)
