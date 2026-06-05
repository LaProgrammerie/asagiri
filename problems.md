# Asagiri — écarts doc / code / spec

> **Remediation_Register actif :** tranche
> [`audit-coherence-consolidation`](.kiro/specs/audit-coherence-consolidation/audit-report.md)
> — constats **AUD-001 … AUD-007**. Une entrée par constat ; statut dans
> `{ouvert, en cours, clôturé}`.
> **Automate de statut :** `ouvert → en cours → clôturé`, réouverture
> `clôturé → ouvert` (réapparition d'un constat → divergence signalée).

## Remediation_Register — audit-coherence-consolidation

> Source des constats : [`audit-report.md`](.kiro/specs/audit-coherence-consolidation/audit-report.md).
> Sévérité : `info` < `warn` < `error` < `blocking`. Clôture d'un constat
> uniquement quand son garde-fou (test/CI) passe au vert. Aucun constat
> `blocking` ne doit rester `ouvert` (ni absent) à la livraison.

| ID | Zone | Problème | Sévérité | Statut |
| --- | --- | --- | --- | --- |
| AUD-001 | docs-site / docgen | runs.mdx manquant ; régénérer la doc CLI | error | clôturé |
| AUD-002 | docs-site / docgen | ~50 pages sans lien fratrie Runs ; régénération | error | clôturé |
| AUD-003 | outillage / CI | golangci-lint inopérant ; pin v2 + Go CI | error | clôturé |
| AUD-004 | registre de dérive | problems.md périmé ; en faire le registre | warn | clôturé |
| AUD-005 | routing | noms d'agents codés en dur ; routing config-driven | warn | clôturé |
| AUD-006 | policy Ollama | rôles liés à une spec historique ; check cohérence | info | clôturé |
| AUD-007 | UX CLI | chemin guidé noyé ; mise en avant sans retrait | info | clôturé |

> **Clôture de la tranche (`2026-06-05`) :** les sept constats `AUD-001 … AUD-007`
> sont `clôturé`. Garde-fous prouvés au vert : Quality_Gate complet
> (`make build` ∧ `go test ./...` ∧ `go vet ./...` ∧ `golangci-lint run`, tous
> exit 0) et Regeneration_Without_Diff (`asa docs generate-cli` → tmp +
> `diff … --exclude=meta.json`, aucune divergence). **Zéro** constat de sévérité
> `blocking` au statut `ouvert`. Note : la réparation d'`AUD-003` a rendu
> `golangci-lint` opérant, ce qui a révélé puis fait corriger une dette lint
> pré-existante hors périmètre des constats (errcheck/ineffassign/staticcheck/
> unused), afin que le gate complet soit réellement vert à la livraison.

## Archive — phase finale (clôturée le `2026-05-29`)

> **Registre maître phase finale :** [`spec-phase-finale.md`](docs/ai/archives/specs/spec-phase-finale.md) §1 (IDs **PF-***, statut **Fermé**).  
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
