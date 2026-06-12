# Current spec — Monetization & Distribution V1 (spec produit)

**Phase :** **monetization-distribution-v1 — M0–M4 spec documentaire** (`2026-06-10`) — roadmap complète, moteur inchangé
**Handoff :** [`handoff.md`](handoff.md)
**Canon distribution :** [`distribution-oss.md`](distribution-oss.md)
**Spec Kiro active :** [`.kiro/specs/monetization-distribution-v1/`](../../.kiro/specs/monetization-distribution-v1/)
**ADR :** [`docs/decisions/037-monetization-distribution.md`](../../decisions/037-monetization-distribution.md)
**Spec moteur livrée :** [`.kiro/specs/agent-orchestration-platform-v1/`](../../.kiro/specs/agent-orchestration-platform-v1/) — T29 Agent Run Export
**Spec précédente :** [`.kiro/specs/trust-engine-v1/`](../../.kiro/specs/trust-engine-v1/) — T12 Trust Gate work

## Positionnement

- **Terraform pour l’orchestration IA de dev** — specs déclaratives, état `.asagiri/`, modules (packs) réutilisables.
- **Local-first AI development control plane** — `asa` exécute en local ; cloud optionnel pour sync/collab.

## Prochaine implémentation

| Phase | Contenu | Statut |
|-------|---------|--------|
| M0 | Spec éditions, matrice inviolables, ADR-037 | ✅ doc |
| M1.0 | Canon `distribution-oss.md` | ✅ doc |
| **M1.1** | **Automation release** — **prochaine implémentation** | ⬜ |
| M1.2–M1.4 | Docs publiques, CHANGELOG, tag `v*` | ⬜ |
| M2.0 | Spec packs `m2-pro-local.md`, ADR-038 | ✅ doc |
| M2.1–M2.4 | `asagiri-packs`, catalogue, commercial | ⬜ |
| M3.0 | Spec cloud `m3-team-cloud.md`, ADR-039 | ✅ doc |
| M3.1–M3.5 | API, sync, dashboard | ⬜ |
| M4.0 | Spec enterprise `m4-enterprise.md`, ADR-040 | ✅ doc |
| M4.1–M4.5 | SSO, audit, on-prem | ⬜ |

Détail : `.kiro/specs/monetization-distribution-v1/tasks.md` · M1 : [`distribution-oss.md`](distribution-oss.md).

---

## Référence — Agent Orchestration Platform V1 (T29 livré)

T13–T29 **livrés** (`2026-06-09`). Le moteur reste **100 % OSS** — voir matrice ADR-037.

| Phase | Contenu | Statut |
|-------|---------|--------|
| T13 | AgentSpec + registry `.asagiri/agents/*.yaml` + `internal/agentspec` | ✅ |
| T14 | ExecutionContext + logs `agentcontext` | ✅ |
| T15 | Provider adapters render-only `agentadapter` | ✅ |
| T16 | Prompt resolve + hook workflow dev (`devresolve`) | ✅ |
| T17 | Doctor drift detection (registry ↔ config ↔ adapter) | ✅ |
| T18 | Output contract validation (`agentcontract`) | ✅ |
| T19 | Hash/versioning + `asa agents list` | ✅ |
| T20 | `asa agents sync` + onboarding | ✅ |
| T21 | External audit read-only (`asa agents external`) | ✅ |
| T22 | External sync write opt-in (`asa agents external sync --write`) | ✅ |
| T23 | Release hardening — smoke CLI + tests drift doctor | ✅ |
| T24 | Agent Run Ledger (`asa agents runs`) | ✅ |
| T25 | Agent Analytics (`asa agents stats`) | ✅ |
| T26 | Agent Run Inspection (`asa agents run`) | ✅ |
| T27 | Agent Run Replay Preview (`asa agents run --preview`) | ✅ |
| T28 | Agent Run Diff (`asa agents diff`) | ✅ |
| T29 | Agent Run Export (`asa agents export`) | ✅ |

## API agentledger export (T29)

```bash
asa agents export <run_id> [--output <dir>] [--include-prompt] [--json]
```

Package : `agentledger.Export(repoRoot, runID, opts)`.

JSON CLI `agent-run-export-v1` : `output_dir`, `manifest_path`, `files[]` (path, size_bytes, sha256).

Bundle : ledger-entry, inspect, replay-preview (sans prompt inline par défaut), artefacts copiés, manifest.

Écriture uniquement dans `--output` (défaut `.asagiri/exports/agents/<run_id>`).

## API agentledger diff (T28)

```bash
asa agents diff <left_run_id> <right_run_id> [--json]
```

Package : `agentledger.Diff(repoRoot, leftRunID, rightRunID)`.

JSON `agent-run-diff-v1` : `identical`, `fields[]`, `artifacts[]` (existence, tailles, timestamps).

Lecture seule — pas de relance provider.

## API agentledger replay preview (T27)

```bash
asa agents run <run_id> --preview [--json]
asa agents run <run_id> --preview --include-prompt [--json]
```

Package : `agentledger.ReplayPreview(repoRoot, runID, opts)`.

JSON `agent-run-replay-preview-v1` : métadonnées ledger + `artifacts[]` avec contenu JSON inline ; `prompt.md` seulement avec `IncludePrompt: true`.

Lecture seule — pas de relance provider.

## API agentledger inspect (T26)

```bash
asa agents run <run_id> [--json]
```

Package : `agentledger.Inspect(repoRoot, runID)`.

JSON `agent-run-v1` : métadonnées ledger + `artifacts[]` (`prompt.md`, `invocation.json`, `context.json`, `contract.json`) avec path, exists, size_bytes, modified_at.

Lecture seule — pas de replay provider ni relance d'exécution.

## API agentanalytics (T25)

```bash
asa agents stats [--json]
asa agents stats --agent <agent_id> [--json]
asa agents stats --provider <provider> [--json]
```

Package : `agentanalytics.Build(repoRoot, opts)`.

JSON `agents-stats-v1` : `global`, `by_agent[]`, `by_provider[]` avec total_runs, success/failure, avg/p95 duration_ms, contract_valid_ratio, last_run_at.

Lecture seule sur `.asagiri/logs/agents/ledger.jsonl` — aucune modification workflow/gates/trust.

## API agentledger (T24)

```bash
asa agents runs [--json]
asa agents runs --task <task_id> [--json]
```

Package : `agentledger.List(repoRoot, opts)`.

JSON `agents-runs-v1` : `ledger_path`, `count`, `entries[]` avec hashes, `contract_valid`, `log_dir`.

Persistance append-only : `.asagiri/logs/agents/ledger.jsonl` (écrit lors de `devOneTask`).

## API agentexternal sync (T22)

```bash
asa agents external sync [--json] [--write] [--force] [--agent <id>]
```

Package : `agentexternal.Sync(repoRoot, cfg, opts)`.

JSON `agents-external-sync-v1` : `mode` (`check`|`write`), `items[]` avec `action` (`create`|`update`|`skip`|`conflict`|`reject`), `content_hash`, `spec_updated`.

Dry-run par défaut ; `--write` pousse le Markdown provider ; conflit hash → exit non-zero sans `--force`.

Allowlist stricte : `spec.external.path`, `config.agents.external_path` — jamais de scan `$HOME`.

## API agentexternal audit (T21)

```bash
asa agents external [--json]
```

Package : `agentexternal.Audit(repoRoot, cfg)`.

JSON `agents-external-v1` : `read_only`, `policy`, `targets[]` avec `provider`, `support_level`, `detected_path`, `writable`, `installed_hash`, `desired_hash`, `status`.

Chemins résolus uniquement depuis `spec.external.path` ou `config.agents.external_path` — **pas** de scan agressif de `$HOME`.

Sync provider reste **opt-in** : T21 n'écrit jamais dans Cursor/Kiro/Codex.

## API agentsync (T20)

```bash
asa agents sync [--json] [--check] [--write] [--force] [--agent <id>]
```

Package : `agentsync.Sync(repoRoot, opts)`.

JSON `agents-sync-v1` : `mode` (`check`|`write`), `items[]` avec `action` (`create`|`update`|`skip`|`conflict`), `hint`.

Dry-run par défaut ; `--write` matérialise sous `.asagiri/agents/` depuis templates embarqués ; conflit hash → exit non-zero sans `--force`.

Onboarding/doctor : warn registry absent + hint `asa agents sync --write`.

## API agentspec (T13)

```go
agentspec.NewLoader(repoRoot)
loader.Load(id) / LoadAll() / List()
agentspec.SemanticHash(spec)
```

## API agentcontext (T14)

```go
agentcontext.Build(Input{Spec, Feature, TaskID, RunID, Phase, UserTaskPrompt, ...})
agentcontext.RenderPrompt(ctx)
agentcontext.ContextHash(ctx)
agentcontext.WriteLogs(repoRoot, ctx, prompt)
agentcontext.AgentLogDir(repoRoot, taskID, agentID)
```

Logs : `.asagiri/logs/<task-id>/agents/<agent-id>/{context.json,prompt.md}`

## API agentadapter (T15)

```go
agentadapter.NewFactory().Render(inv)
agentadapter.RenderFromConfig(cfg, agentKey, spec, ctx)
agentadapter.SupportMatrix(spec)
agentadapter.Explain(providerType)
```

Support : `native_profile` (claude-code) | `inline_prompt` (cursor, kiro, codex, exec…) | `unsupported`

## API devresolve (T16)

```go
devresolve.LegacyDevPrompt(taskID)
devresolve.Resolve(Params{RepoRoot, Config, AgentKey, RunID, Feature, TaskID, ContextFiles})
```

Logs orchestrés : `.asagiri/logs/<task-id>/agents/<agent-id>/{context.json,prompt.md,invocation.json,resolve.json}`

Fallback : spec absent/invalide sous `.asagiri/agents/` → prompt legacy inchangé + `dev-orchestration.warn`.

## API doctor agents (T17)

```bash
asa doctor [--json] [--strict]
```

Champs JSON : `agent_registry`, `agent_specs`, `agent_drift`, `last_orchestrated`.

Section texte : **Agent specs**, **Agent drift**, warnings actionnables + hint `asa agents sync --write`.

Registry absent → WARN (templates embarqués) ; résolution via `asa agents sync --write`.

## API doctor architecture (hygiène release)

```bash
asa doctor architecture [--json]
```

Package : `doctor.BuildArchitecture(repoRoot)`.

JSON `doctor-architecture-v1` : `repository`, `summary`, `sources`, `findings[]`, `recommendations[]`.

Croisement read-only tasks, execution graphs, knowledge store, trust reports et agent ledger — aucune mutation workflow/gates/trust/ledger.

## API agentcontract (T18)

```go
agentcontract.ValidateOutput(spec agentspec.Spec, rawOutput string) ContractValidationResult
agentcontract.WriteLog(repoRoot, taskID, agentID, result)
```

Formats : `text`, `asagiri-v1`, `gate-yaml`, `gate-json`. Codes erreur : `missing_required_field`, `invalid_json`, `invalid_yaml`, `unknown_format`, `empty_output`.

JSON stable : `valid`, `format`, `errors[]`, `warnings[]`, `extracted_summary` (optionnel).

Hook **dev orchestré uniquement** (après `agent.Run`) → `.asagiri/logs/<task>/agents/<agent>/contract.json`. Non bloquant ; fallback legacy sans spec → pas de `contract.json`. Gates/trust/scoring **inchangés**.

## API agentslist (T19)

```bash
asa agents list [--json]
asa agents show <id> [--json]
```

Package : `agentslist.Build(repoRoot, cfg)`, `agentslist.Show(repoRoot, id, cfg)`.

JSON `agents-list-v1` : `registry`, `agents[]` avec `id`, `role`, `version`, `content_hash`, `stored_hash`, `source` (`disk`|`embedded`), `output_format`, `provider_targets`, `provider_support` (config + matrice cibles).

Read-only : pas de sync, pas d'installation provider.

## Trust Engine V1 (référence — T12 livré)

## CLI trust dédié

```bash
asa trust task <task-id> [--json] [--explain]
asa trust feature <feature> [--json] [--explain]
asa trust run <run-id> [--json] [--explain]
```

## CLI quotidien

```bash
asa next [--feature <feature>] [--no-trust]
asa status [--no-trust]
asa work / asa continue          # ligne trust après exécution
asa doctor [--full] [--json] [--strict] [--save]   # --save → .asagiri/reports/doctor/latest.json ; confirmation save sur stderr
asa doctor architecture [--json]                   # croisement read-only tasks/graphs/knowledge/trust/ledger
asa doctor diff [--json] [--from] [--to]           # diff latest vs history (≥2 --save)
asa trust task|feature|run … [--save]              # --save → .asagiri/reports/trust/… ; avec --json : stdout JSON pur
asa trust diff task|feature|run … [--json]           # diff snapshots trust enregistrés
```

Cohabite avec `asa verify trust <flow>` (spec-my-B) et `asa trust gates|replay`.

## Livré

| Phase | Contenu | Statut |
|-------|---------|--------|
| T1 | `internal/worktrust` — BuildTaskReport, scorer | ✅ |
| T2 | CLI task + feature | ✅ |
| T3 | CLI run, plan gate | ✅ |
| T4 | UX texte (sections), `--explain`, harmonisation `asa next` | ✅ |
| T5 | Trust daily UX (next/status/work, `--no-trust`) | ✅ |
| T6 | `internal/doctor`, `asa doctor` enrichi, `--json` | ✅ |
| T7 | Doc UX mental-model + daily-workflow (4 locales), liens guided-path | ✅ |
| T8 | `internal/reportsink`, `--save` trust/doctor, écriture atomique JSON | ✅ |
| T9 | Operator & CI UX : streams JSON/stderr, daily-workflow CI, `02-architecture`, `SilenceErrors` | ✅ |
| T10 | Alignement recommendations : `RecommendForTask`, trust ↔ `asa next`, tests parité | ✅ |
| T11 | History `reportsink`, `reportdiff`, `asa doctor diff`, `asa trust diff` | ✅ |
| T12 | `work.gates.trust`, `WorkTrustReportToGateResult`, hook verify → review guard | ✅ |

## Config trust gate (opt-in)

```yaml
work:
  gates:
    trust:
      enabled: false
      mode: per-task
      min_score: 70
      block_verdicts: [blocked]
      warn_verdicts: [risky]
```

## API library

```go
doctor.Build(startDir, opts)
doctor.FormatText / FormatJSON
worktrust.BuildTaskReport / BuildFeatureReport / BuildRunReport
intent.RecommendForTask / RecommendNextFromTasks
reportdiff.DiffTrustTask / DiffDoctor
reportsink.ListHistory / DiffPairPaths
worktrust.FormatDailyNextBlock / FormatDailyStatusBlock / FormatDailyPostWorkLine
```

Branding : **Asagiri** / **`asa`**.
