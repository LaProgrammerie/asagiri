# TODO priorisée (post-audit)

## P0 — avant publication OSS

1. Couverture `workflow` et `intent` > 50 %
2. GitHub Actions : `go test -race`, golangci-lint
3. Appliquer `redact` sur sorties agent en erreur
4. Audit historique git secrets

## P1 — confiance produit

5. Cache context pack (feature + task + commit)
6. Golden tests plan JSON (`work --plan-only`)
7. Propagation `context.Context` cancellation → agents
8. Tests intégration executor + workflow dry-run

## P2 — polish

9. Badges README + diagramme architecture (mermaid)
10. Tokenizers provider-exacts
11. MCP `get_run_status` branché SQLite
12. TUI bubbletea interactif

## P3 — backlog

13. Embeddings RAG
14. Séparer template PHP du binaire AgentFlow
15. Multi-repo orchestration
