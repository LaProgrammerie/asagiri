# Asagiri — écarts doc / code / spec

> **Registre maître phase finale :** [`spec-phase-finale.md`](spec-phase-finale.md) §1 (IDs **PF-***, statut **Fermé**).  
> **Clôture (`2026-05-29`) :** tous les GAP-* et PF-* synchronisés — voir [`handoff.md`](docs/ai/active/handoff.md) archive phase finale.

| ID | Zone | Problème | Sévérité | Statut |
| --- | --- | --- | --- | --- |
| GAP-001 | CLI `resume` | Sans `--execute`, affiche le prochain step. Avec `--execute`, enchaîne les steps restants (agents réels hors dry-run global). | Moyenne | **Clôturé** — PF-X-01 (`2026-05-29`) |
| GAP-002 | Cost | Tokenizers provider dans `cost/tokenizer.go` ; fallback heuristique `chars_per_token`. Montants = estimations, pas factures. | Faible | **Clôturé** — PF-X-02 (`2026-05-29`) |
| GAP-003 | RAG | `asa index` + `index search` : embeddings via `runtime.memory.embedder` ; `--keyword` / `--skip-embeddings` documentés. | Faible | **Clôturé** — PF-X-03 (`2026-05-29`) |
| GAP-004 | Docgen | Pages CLI avec `cobra.Example` ; régénération `asa docs generate-cli`. | Faible | **Clôturé** — PF-X-04 (`2026-05-29`) |

## Aucun écart bloquant identifié le 2026-05-17

- Flags documentés dans `docs-site/content/docs/en/cli/generated/` régénérés depuis Cobra (`docs generate-cli`) et cohérents avec `work.go` / `root.go` lors de la revue manuelle.
- Pas de flag documenté inexistant sur les commandes critiques (`work`, `estimate`, `sync`, `resume`).
- Comportements expérimentaux (MCP, Notion) marqués **Experimental** dans le site docs.

### Limitations connues (non-bugs)

- Les agents cloud envoient des données aux fournisseurs configurés — hors scope du sandbox Asagiri.
- `pricing.models` à `0` affiche un coût nul mais les plafonds de tokens restent actifs.
- `sync notion` nécessite `sources.notion.enabled` + token env valide.
