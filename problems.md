# Asagiri — écarts doc / code / spec

> **Registre maître phase finale :** [`spec-phase-finale.md`](spec-phase-finale.md) §1 (IDs **PF-***).  
> **P1 (`2026-05-29`) :** PF-A-01/02, PF-C-01…05 fermés en code — voir [`handoff.md`](docs/ai/active/handoff.md).  
> Ce fichier garde la traçabilité **GAP-*** ↔ **PF-X-*** (P2/P3 encore ouverts).

| ID | Zone | Problème | Sévérité | Statut |
| --- | --- | --- | --- | --- |
| GAP-001 | CLI `resume` | Sans `--execute`, affiche le prochain step seulement. Avec `--execute`, exécute un step (agents réels hors dry-run global). | Moyenne | **Clôturé** — PF-X-01 (`2026-05-29`) |
| GAP-002 | Cost | Pas de tokenizers provider-exacts : estimation par heuristique `chars_per_token` uniquement (`cost/token_counter.go`). La roadmap P0 le prévoit ; la doc ne doit pas présenter les montants comme factures. | Faible | → **PF-X-02** |
| GAP-003 | RAG | `asa index` = chunks keyword (LIKE) ; embeddings sémantiques via `memory reindex` + PF-A-01 embedder. | Faible | **Clôturé (doc)** — PF-X-03 (`2026-05-29`) |
| GAP-004 | Docgen | Pages CLI sans args obligatoires tant que `cobra.Example` absent. | Faible | **Clôturé** — PF-X-04 (`2026-05-29`) |

## Aucun écart bloquant identifié le 2026-05-17

- Flags documentés dans `docs-site/content/docs/cli/generated/` régénérés depuis Cobra (`docs generate-cli`) et cohérents avec `work.go` / `root.go` lors de la revue manuelle.
- Pas de flag documenté inexistant sur les commandes critiques (`work`, `estimate`, `sync`, `resume`).
- Comportements « non stables » (MCP, Notion, resume auto, tokenizers, RAG vectoriel) marqués **Experimental** ou **Limitation** dans le site docs.

### Limitations connues (non-bugs)

- Les agents cloud envoient des données aux fournisseurs configurés — hors scope du sandbox Asagiri.
- `pricing.models` à `0` affiche un coût nul mais les plafonds de tokens restent actifs.
- `sync notion` nécessite `sources.notion.enabled` + token env valide.
