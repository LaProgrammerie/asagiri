# Decisions (lightweight ADRs)

**Policy:** any decision that **durably** affects architecture, code standards, team workflows, or product invariants **must** be recorded here (even briefly). If it is not in this log, it is not considered decided for the repo.

**Sync:** if a durable decision changes architecture or standards, the affected files in `docs/ai/` (`02-architecture.md`, `03-standards.md`, `04-workflows.md`, etc.) **must be updated in the same change** (same PR / same batch of commits), not “when we get around to it”.

Suggested format per entry:

| ID | Date | Decision | Context | Consequences |
|----|------|----------|---------|--------------|
| ADR-001 | 2026-05-17 | Stack Go ; module racine ; code sous `application/` ; Docker local via Compose ; orchestration `Makefile` ; pas Castor ni Yoimachi dans le template | Adaptation du template AI Engineering depuis PHP/docker-starter | Doc, squelette `agentflow`, CI Go ; déploiement/IaC hors scope jusqu’à nouvelle ADR |
| ADR-002 | 2026-05-17 | État AgentFlow en **SQLite** via `modernc.org/sqlite` (`CGO_ENABLED=0`) ; schéma V1 : `schema_version`, `runs`, `tasks` ; migrations SQL embarquées (`embed`) | Tranche `agentflow-init` — fondation traçable avant worktrees/agents | `state.sqlite` sous `.agentflow/` (gitignored) ; pas de Postgres pour l’état local V1 |
| ADR-003 | 2026-05-17 | Agents via **subprocess** (`exec.Command`, pas de shell) ; mode **`--dry-run` / `AGENTFLOW_DRY_RUN=1`** pour CI et tests sans binaires kiro/cursor/codex | V1 workflow agentique reproductible | Config `agents.*` dans `config.yaml` ; logs factices en dry-run |
| ADR-004 | 2026-05-17 | **Worktree Git par tâche** sous `.agentflow/worktrees/` ; branche `agentflow/<feature>/<task-id>` | Isolation des modifications agent | `clean` avec filtres `--merged` / `--failed` |
| ADR-005 | 2026-05-17 | **Validation externe** : commandes depuis `validation.commands` (config) ou payload tâche ; défauts Go si `go.mod` | Principe §2.4 spec — l’agent ne valide pas seul | Pas de `composer` dans les défauts du template Go |
| ADR-006 | 2026-05-17 | **Modèle tâche canonique** dans `pkg/agentflow` ; persistance YAML+JSON sous `.agentflow/tasks/` | spec §8 | Statuts §8.2 distincts des statuts `runs` |
| ADR-007 | 2026-05-17 | **RAG local minimal** : `chunks.sqlite` + recherche LIKE ; pas d’embeddings en V1 | spec §10.3 | `agentflow index` ; enrich heuristique si index absent |
| ADR-008 | 2026-05-17 | **State machine** explicite dans `workflow/state_machine.go` ; `--force` ; `resume --execute` dry-run | spec §12 | Reprise hors dry-run = diagnostic uniquement |
| ADR-009 | 2026-05-17 | **Intent Layer** (`internal/intent`) : resolver hybride + planner composant les primitives V1 ; sources via `internal/source` ; repo = vérité exécutable | specv2 §3–6 | Notion jamais exécuté sans sync local ; `NOTION_TOKEN` via env |
| ADR-010 | 2026-05-17 | **V3 cost/perf** : packages `cost`, `contextopt`, `investigation`, `telemetry`, `tui` ; pipeline `RunV3Pipeline` ; prix **uniquement** via `pricing.models` (pas de hardcode) ; MCP stdio désactivé par défaut ; TUI isolée (`internal/tui`) | specv3 | Migration SQLite `run_metrics` / `step_metrics` ; `work --estimate-only` |
| ADR-011 | 2026-05-17 | **Consolidation post-V3** : spec active `spec-postv123.md` ; collecte contexte via candidats investigation (`CollectForPipeline`) ; LICENSE **Apache 2.0** LaProgrammerie ; explainability dans sorties estimate/work ; package `redact` | Ouverture OSS, réduction I/O et dette | Audits `docs/consolidation/` ; pas de breaking CLI V1–V3 |
| ADR-012 | 2026-05-17 | **Docs publiques** : `docs-site/` (Fumadocs + Next static export) ; référence CLI générée via `agentflow docs generate-cli` ; canon agents reste `docs/ai/` ; `basePath` `/hyper-fast-builder` uniquement si `GITHUB_PAGES=true` (legacy) | `spec-doc.md` ; OSS international (EN) | ADR publics dans `docs/decisions/` ; hébergement : **ADR-014** |
| ADR-014 | 2026-05-17 | **Hébergement docs** : Cloudflare Pages (direct upload Wrangler) ; CI `.github/workflows/docs-cloudflare-pages.yml` (prod `main`, preview PR) + `.github/workflows/docs-check.yml` (build sans secrets) ; **pnpm** dans `docs-site/` ; workflow GitHub Pages `docs.yml` **retiré** | `spec-deploy-doc.md` ; custom domain / previews PR | Secrets repo : `CLOUDFLARE_API_TOKEN`, `CLOUDFLARE_ACCOUNT_ID`, `CLOUDFLARE_PAGES_PROJECT` ; pas de `GITHUB_PAGES` en CI Cloudflare |
| ADR-013 | 2026-05-17 | **Docs i18n** : Fumadocs `parser: dir` — `content/docs/{en,fr,de,es}/` ; URLs `/docs` (en) et `/docs/{fr,de,es}/…` ; sélecteur de langue (navbar Fumadocs) ; CLI Cobra générée **en anglais** sous `en/cli/generated/` ; FR/DE/ES : pages manuelles traduites + stubs narratifs pour commandes clés | `spec-doc.md` ; demande OSS multilingue | `TestPublicDocsNoPlaceholders` parcourt les 4 locales ; pas de `--locale` docgen en V1 |

## Log

*(Add structural tradeoffs as you go.)*
