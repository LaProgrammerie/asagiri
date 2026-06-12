# ADR-036 — Agent Orchestration Platform V1

**Date :** 2026-06-08  
**Statut :** accepté (T13 foundation)  
**Spec :** `.kiro/specs/agent-orchestration-platform-v1/`

## Contexte

Les prompts et contrats agents sont dispersés (workflow builders, fichiers Cursor/Kiro externes, chaînes fixes). `config.agents` ne transporte que commande/provider.

## Décision

1. Introduire **AgentSpec** canonique (YAML plat) sous `.asagiri/agents/<id>.yaml`.
2. Package **`internal/agentspec`** : load, validate, hash sémantique stable.
3. **Templates embarqués** (dev, reviewer, enricher, governance) si registry absent.
4. `config.agents` reste la couche **transport** ; pas de hook workflow en T13.

## Schéma T13 (minimal)

Champs : `id`, `version`, `role`, `provider_targets`, `system_prompt`, `instructions`, `constraints`, `output_contract`, `metadata`.

Hash : SHA-256 du payload sémantique JSON (ignore `metadata.updated_at`, `content_hash`, etc.).

Read-only list/show (T19) ; sync registry disque (T20).

## Conséquences

- T14+ injecte ExecutionContext sans casser gates/trust.
- **`asa agents sync` (T20)** — matérialise `.asagiri/agents/*.yaml` depuis templates embarqués ; dry-run par défaut, `--write` / `--force` ; **pas** de sync provider externe en V1.
- Mapping `work.default_*` → ids registry documenté en T14.

## Sync registry (T20)

```bash
asa agents sync              # dry-run (check)
asa agents sync --write      # crée/met à jour fichiers absents ou identiques
asa agents sync --write --force   # écrase conflits hash
asa agents sync --json       # stdout JSON pur (agents-sync-v1)
```

Actions par agent : `create`, `update`, `skip`, `conflict`. Conflit = hash disque ≠ template sans `--force` → exit non-zero.

**Hors scope V1 :** push/pull `spec.external.path`, `last_synced_hash` provider, installation profils Cursor/Kiro.

## External audit read-only (T21)

```bash
asa agents external              # rapport texte
asa agents external --json       # stdout JSON pur (agents-external-v1)
```

Champs `ExternalTarget` : `provider`, `support_level`, `detected_path`, `writable`, `installed_hash`, `desired_hash`, `status`.

Sources de chemin autorisées : `spec.external.path`, `config.agents.external_path`. Expansion `~/` uniquement pour ces chemins explicites.

**Pourquoi opt-in :** les profils Cursor/Kiro/Codex vivent hors du dépôt ; une sync automatique risquerait d'écraser la config utilisateur ou de dépendre de chemins fragiles. T21 audite ; T22 écrit **uniquement** avec `--write` vers chemins explicites.

## External sync write opt-in (T22)

```bash
asa agents external sync              # dry-run (check)
asa agents external sync --write      # écrit le profil Markdown provider
asa agents external sync --write --force --agent dev --json
```

Actions : `create`, `update`, `skip`, `conflict`, `reject`. Conflit = fichier externe modifié hors Asagiri sans `--force`.

Contenu écrit : frontmatter `asagiri: true`, `agent_id`, `agent_version`, `content_hash`, sections output contract + prompt orchestré (`agentcontext`).

Post-write : `spec.external.last_synced_hash` mis à jour dans `.asagiri/agents/<id>.yaml`.

Doctor : `external_drift`, `external_missing` → `asa agents external sync --write --agent <id>`.

## Release hardening (T23)

Audit read-only T13–T22 : build + tests + docgen nodiff ; smoke CLI documenté dans `tasks.md` § T23.

Invariants revalidés : JSON stdout pur, hints stderr, allowlist chemins externes, pas de scan `$HOME`, logs orchestrés sous `.asagiri/logs/<task>/agents/<agent>/`.

## Agent Run Ledger (T24)

```bash
asa agents runs [--json]
asa agents runs --task <task_id>
```

Append-only : `.asagiri/logs/agents/ledger.jsonl` — une ligne JSON par exécution agent (hook `devOneTask`).

Champs : task/run/feature, agent/role/provider, durée, exit_code, `prompt_hash`, `context_hash`, `output_hash`, `contract_valid`, `log_dir`.

Gates / trust / scoring **inchangés**.

## Agent Analytics (T25)

```bash
asa agents stats [--json]
asa agents stats --agent <agent_id>
asa agents stats --provider <provider>
```

Package `agentanalytics` — agrégation read-only du ledger : global, par agent, par provider.

Métriques : `total_runs`, `success_count` / `failure_count` (exit_code), `avg_duration_ms`, `p95_duration_ms`, `contract_valid_ratio`, `last_run_at`.

Hors scope : replay, coûts tokens, dashboard TUI.

## Agent Run Inspection (T26)

```bash
asa agents run <run_id> [--json]
```

`agentledger.Inspect` — rapport `agent-run-v1` : métadonnées ledger + artefacts logs (`prompt.md`, `invocation.json`, `context.json`, `contract.json`).

Affiche chemins, existence, tailles et timestamps. Lecture seule.

Hors scope : replay provider réel, relance d'exécution, coûts tokens, dashboard TUI.

## Agent Run Replay Preview (T27)

```bash
asa agents run <run_id> --preview [--json]
asa agents run <run_id> --preview --include-prompt
```

`agentledger.ReplayPreview` — rapport `agent-run-replay-preview-v1` : reconstruction read-only des entrées exploitables (JSON inline, prompt optionnel).

Hors scope : relance provider, comparaison entre runs (T28), mutation ledger.

## Agent Run Diff (T28)

```bash
asa agents diff <left_run_id> <right_run_id> [--json]
```

`agentledger.Diff` — rapport `agent-run-diff-v1` : métadonnées, hashes, contract_valid, artefacts (existence/taille/timestamp).

Hors scope : relance provider, mutation ledger.

## Agent Run Export Bundle (T29)

```bash
asa agents export <run_id> [--output <dir>] [--include-prompt] [--json]
```

`agentledger.Export` — bundle read-only : ledger-entry, inspect, replay-preview, artefacts, manifest (hashes/tailles).

Écriture uniquement dans le dossier d’export. Hors scope : relance provider, mutation ledger source.
