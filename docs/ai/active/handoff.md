# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `agentflow-spec-7-12` : en cours** (`2026-05-17`).

## Immediate objective

Aligner code, tests, config et `docs/ai/` sur **spec.md §7–12** sans casser la V1 CLI §6.1.

## Allowed scope (agentflow-spec-7-12)

- `application/internal/config`, `workflow`, `agent`, `store/sqlite`, `validation`, `policy`, `rag`
- `application/pkg/agentflow/types.go`
- `application/internal/cli/` (+ commande `index`)
- `.agentflow/config.yaml.example`
- `docs/ai/` (handoff, current-spec, 02, 03, 05)
- `README.md` (sections AgentFlow)

## Definition of Done — agentflow-spec-7-12

- [x] Config §7 : validation.commands, policies, agents étendus + défauts Go si `go.mod`
- [x] Modèle tâche §8 : `pkg/agentflow/types`, fichiers `.agentflow/tasks/<feature>/*.yaml` + JSON
- [x] Contrat agent §9 : `AgentContext` / `AgentResult`, logs `context.json` / `result.json`
- [x] RAG §10 : `internal/rag`, `agentflow index`, enrich avec retriever/heuristique
- [x] Interfaces §11 : `WorkflowEngine`, `TaskStore`, `WorktreeManager`, `Validator` (déclarées)
- [x] State machine §12 : transitions, `--force`, `resume` + `--execute` (dry-run)
- [x] `go test -race ./...` vert
- [ ] `golangci-lint` vert (nécessite toolchain ≥ go.mod sur la machine CI/dev)

## Hors scope

- Embeddings Ollama réels (index = chunks texte + LIKE ; `embeddings.sqlite` réservé)
- Reprise automatique `resume` hors dry-run
- Renommage `application/` → `agentflow/` (ADR-001)

## References

- [`spec.md`](../../spec.md) §7–12
- [`current-spec.md`](current-spec.md)
- [`05-decisions.md`](../05-decisions.md) ADR-005 à ADR-008
