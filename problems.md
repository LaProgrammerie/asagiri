# AgentFlow — écarts doc / code / spec

| ID | Zone | Problème | Sévérité | Statut |
| --- | --- | --- | --- | --- |
| GAP-001 | CLI `resume` | Hors `--dry-run`, `agentflow resume <run-id>` affiche seulement le prochain step (`plan`, `dev`, …) et **n’exécute pas** les agents. L’enchaînement automatique existe uniquement via `resume --execute` avec `--dry-run` (`ResumeRunDryExecute` retourne une erreur sinon). | Moyenne | Documenté (Experimental) — implémentation future |
| GAP-002 | Cost | Pas de tokenizers provider-exacts : estimation par heuristique `chars_per_token` uniquement (`cost/token_counter.go`). La roadmap P0 le prévoit ; la doc ne doit pas présenter les montants comme factures. | Faible | Limitation connue |
| GAP-003 | RAG | `agentflow index` indexe des chunks SQLite ; pas d’embeddings vectoriels ni retrieval sémantique malgré `embedding_model` dans l’exemple config. | Faible | Roadmap P2 — Experimental en doc |
| GAP-004 | Docgen | Les pages CLI générées utilisent un exemple minimal `agentflow <cmd>` sans args obligatoires (ex. `work` sans instruction) tant que `cobra.Command.Example` n’est pas renseigné. | Faible | Amélioration docgen |

## Aucun écart bloquant identifié le 2026-05-17

- Flags documentés dans `docs-site/content/docs/cli/generated/` régénérés depuis Cobra (`docs generate-cli`) et cohérents avec `work.go` / `root.go` lors de la revue manuelle.
- Pas de flag documenté inexistant sur les commandes critiques (`work`, `estimate`, `sync`, `resume`).
- Comportements « non stables » (MCP, Notion, resume auto, tokenizers, RAG vectoriel) marqués **Experimental** ou **Limitation** dans le site docs.

### Limitations connues (non-bugs)

- Les agents cloud envoient des données aux fournisseurs configurés — hors scope du sandbox AgentFlow.
- `pricing.models` à `0` affiche un coût nul mais les plafonds de tokens restent actifs.
- `sync notion` nécessite `sources.notion.enabled` + token env valide.
