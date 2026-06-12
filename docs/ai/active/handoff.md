# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.
> **Statut :** **monetization-distribution-v1 M0–M4 spec documentaire livrée** (`2026-06-10`).
> **Moteur :** agent-orchestration-platform-v1 T29 livré — inchangé, 100 % OSS (ADR-037).
> **Prochaine implémentation :** configurer `HOMEBREW_TAP_GITHUB_TOKEN` + première release tag (M1.4) → M2.1 `asagiri-packs`.
> **Docs :** [`getting-started-public.md`](getting-started-public.md) · [`distribution-oss.md`](distribution-oss.md)

## Monetization & Distribution V1 — M2/M3/M4 (spec documentaire) ✅

### Livré

- **M2 Pro Local** — `m2-pro-local.md` + ADR-038 (`asagiri-pack-v1`, `asagiri-packs`, signature, marketplace future)
- **M3 Team Cloud** — `m3-team-cloud.md` + ADR-039 (registry, sync, runs, analytics, backend)
- **M4 Enterprise** — `m4-enterprise.md` + ADR-040 (SSO, RBAC, audit, on-prem, SLA)
- **Roadmap** — `design.md` + `tasks.md` M1→M4 continus

### Verdict spec M2–M4

**GO doc** — prêt pour implémentation incrémentale sans toucher au moteur.

---

## Monetization & Distribution V1 — M1.0 (canon distribution OSS)

### Livré

- **`docs/ai/active/distribution-oss.md`** — licence Apache 2.0, versioning, canaux install, release GitHub, artefacts, checklist, hors scope
- **Homebrew** — `brew install asagiri` → binaire `asagiri` (`bin.install "asa" => "asagiri"`, évite `/usr/bin/asa` macOS)
- **Install script** — `scripts/install.sh` (curl archive + SHA256) ; procédure `docs/ai/active/release-process.md`
- **GoReleaser** — existant documenté (`.goreleaser.yaml`, `release.yml`) — pas de réimplémentation
- **Tasks M1** — découpage M1.0–M1.4 dans `.kiro/specs/monetization-distribution-v1/tasks.md`

### Hors scope M1.0

- Aucun changement Go / `.goreleaser.yaml`
- Pas de tag `v*` public ni première release
- Pas de licence enforcement

### Verdict M1.0

**M1.1** — livré ; tap corrigé → `homebrew-tap` ; PAT `HOMEBREW_TAP_GITHUB_TOKEN` doit avoir `contents:write` sur `LaProgrammerie/homebrew-tap`.

Spec : `.kiro/specs/monetization-distribution-v1/tasks.md` · Canon : `distribution-oss.md`.

---

## Monetization & Distribution V1 — M1.2 (onboarding public)

### Livré

- **`docs/ai/active/getting-started-public.md`** — install, init, agents sync, dry-run, troubleshooting, What is/is not, No account
- **docs-site** `getting-started/installation.mdx` (en/fr/de/es) — ordre Homebrew → install.sh → manuel ; exemples `VER`/`OS`/`ARCH`
- **README** — sections positionnement + install unifiée
- **`distribution-oss.md` / `release-process.md`** — convention archives corrigée

### Verdict M1.2

**GO doc** — prêt pour utilisateur externe ; brew/script à valider sur première release publique.

---

## Monetization & Distribution V1 — M0 (spec produit)

### Livré

- **Spec Kiro** — `.kiro/specs/monetization-distribution-v1/` (`requirements`, `design`, `tasks`)
- **ADR-037** — `docs/decisions/037-monetization-distribution.md`
- **Éditions** — OSS, Pro Local, Team Cloud, Enterprise
- **Matrice** — capacités × éditions ; liste **inviolables** (jamais paywallés)
- **Positionnement** — Terraform orchestration IA ; local-first control plane
- **Contraintes** — monétisation additive ; cloud optionnel ; pas d’enforcement licence dans `cmd/asa`

### Hors scope M0

- Aucun changement Go / CLI / UI
- Pas de licence enforcement
- Pas de modification des features existantes

### Verdict M0

**GO spec** — prêt pour planification **M1 distribution OSS**.

Spec : `.kiro/specs/monetization-distribution-v1/tasks.md` · ADR : `docs/decisions/037-monetization-distribution.md`.

---

## Hygiène release — doctor architecture (`2026-06-10`)

### Livré

- **`doctor.BuildArchitecture`** — rapport `doctor-architecture-v1`
- **Croisement read-only** — tasks, execution graphs, knowledge, trust reports, agent ledger
- **CLI** — `asa doctor architecture [--json]`
- **Extraction** — commande dans `internal/doctorcli` (wiring `cli/doctor_wiring.go`)
- **Tests** — unit (`internal/doctor/architecture_test.go`) + CLI JSON
- **Docs CLI** — `doctor-architecture.mdx` (docgen)

### Invariants

- Lecture seule — pas de mutation workflow/gates/trust/ledger
- `--json` → stdout JSON pur ; stderr vide en succès

### Verdict

**GO** — diagnostic architecture croisé exploitable en CLI.

Référence : `docs/ai/active/architecture-audit.md`.

---

## Agent Orchestration Platform V1 — T29 (export bundle)

### Livré

- **`agentledger.Export`** — rapport CLI `agent-run-export-v1`
- **Bundle** — `ledger-entry.json`, `inspect.json`, `replay-preview.json`, `artifacts/`, `manifest.json` (hashes + tailles)
- **CLI** — `asa agents export <run_id> [--output <dir>] [--include-prompt] [--json]`
- **Lecture seule** sur ledger/logs ; écriture uniquement dans le dossier d’export
- **Tests** — unit + CLI (introuvable, sans/avec prompt, manifest)
- **Docs CLI** — `agents-export.mdx` regénéré

### Verdict T29

**GO** — export bundle exploitable en CLI.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T29.

---

## Agent Orchestration Platform V1 — T28 (run diff)

### Livré

- **`agentledger.Diff`** — rapport `agent-run-diff-v1`
- **Comparaison** — métadonnées, hashes, contract_valid, artefacts (existence/taille/timestamp)
- **CLI** — `asa agents diff <left_run_id> <right_run_id> [--json]`
- **Lecture seule** — pas de relance provider ; workflow/gates/trust inchangés
- **Tests** — unit + CLI (identiques, hashes, artefacts manquants, introuvable)
- **Docs CLI** — `agents-diff.mdx` regénéré

### Verdict T28

**GO** — comparaison de runs exploitable en CLI.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T28.

---

## Agent Orchestration Platform V1 — T27 (replay preview)

### Livré

- **`agentledger.ReplayPreview`** — rapport `agent-run-replay-preview-v1`
- **Reconstruction read-only** — hashes ledger + artefacts logs avec contenu JSON inline
- **Prompt optionnel** — `--include-prompt` pour `prompt.md` complet
- **CLI** — `asa agents run <run_id> --preview [--json] [--include-prompt]`
- **Invariants** — pas de relance provider ; workflow/gates/trust/scoring inchangés
- **Tests** — unit + CLI (absent, complet, manquants, include-prompt)
- **Docs CLI** — `agents-run.mdx` regénéré

### Verdict T27

**GO** — replay preview exploitable en CLI.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T27.

---

## Agent Orchestration Platform V1 — T26 (run inspection)

### Livré

- **`agentledger.Inspect`** — rapport `agent-run-v1` par `run_id`
- **Métadonnées** — task/feature/agent/role/provider, durée, exit_code, hashes, contract_valid, log_dir
- **Artefacts** — `prompt.md`, `invocation.json`, `context.json`, `contract.json` (chemin, existence, taille, timestamp)
- **CLI** — `asa agents run <run_id> [--json]`
- **Lecture seule** — workflow/gates/trust/scoring/analytics inchangés
- **Tests** — unit inspect + CLI (absent, présent, logs manquants)
- **Docs CLI** — `agents-run.mdx` regénéré

### Verdict T26

**GO** — inspection run exploitable en CLI.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T26.

---

## Agent Orchestration Platform V1 — T25 (agent analytics)

### Livré

- **`internal/agentanalytics`** — `Build`, rapport `agents-stats-v1`
- **Agrégations** — global, `by_agent`, `by_provider`
- **Métriques** — total_runs, success/failure, avg/p95 duration_ms, contract_valid_ratio, last_run_at
- **CLI** — `asa agents stats [--json] [--agent <id>] [--provider <provider>]`
- **Lecture seule** — consomme `agentledger.List` ; workflow/gates/trust/scoring inchangés
- **Tests** — unit analytics + CLI JSON
- **Docs CLI** — `agents-stats.mdx` regénéré

### Verdict T25

**GO** — analytics ledger exploitables en CLI.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T25.

---

## Agent Orchestration Platform V1 — T24 (agent run ledger)

### Livré

- **`internal/agentledger`** — `Entry`, `Record`, `List`, rapport `agents-runs-v1`
- **Persistance** — `.asagiri/logs/agents/ledger.jsonl` (append-only JSONL)
- **Champs** — task/run/feature, agent/role/provider, durée, exit_code, prompt/context/output hash, contract_valid, log_dir
- **Hook** — enregistrement après `devOneTask` ; métadonnées via `devresolve` (ContextHash, LogDir)
- **CLI** — `asa agents runs [--json]`, `asa agents runs --task <id>`
- **Tests** — unit + CLI + workflow orchestration
- **Docs CLI** — `agents-runs.mdx` regénéré

### Invariants T24

- Gates / trust / scoring **inchangés**
- `--json` → stdout JSON pur
- Lecture seule via CLI (pas de mutation ledger hors exécution agent)

### Verdict T24

**GO** — traçabilité exécutions agents exploitable.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T24.

---

## Agent Orchestration Platform V1 — T23 (release hardening)

### Livré

- **Audit read-only** T13–T22 — git clean, docs alignées, docgen nodiff
- **Smoke CLI** — `agents list/show/sync`, `agents external/sync`, `doctor --json` ; repo temp cycle create/write/skip/conflict/force
- **Invariants vérifiés** — JSON stdout pur, hints stderr, pas de scan `$HOME`, pas d'écriture sans `--write`
- **Logs orchestrés** — `.asagiri/logs/<task>/agents/<agent>/` (`context.json`, `prompt.md`, `contract.json`) via tests workflow
- **Test angle mort** — `agentexternal.ExternalDrift` (doctor `external_drift` / `external_missing`)

### Verdict T23

**GO** — plateforme orchestration V1 release-ready.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T23.

## Agent Orchestration Platform V1 — T22 (external sync write opt-in)

### Livré

- **CLI** — `asa agents external sync [--json] [--write] [--force] [--agent <id>]`
- **`internal/agentexternal.Sync`** — dry-run par défaut ; rapport `agents-external-sync-v1`
- **Allowlist** — écriture uniquement vers `spec.external.path` ou `config.agents.external_path`
- **Refus** — chemin absent, provider unsupported, conflit hash sans `--force`, hors allowlist
- **Format Markdown** — frontmatter Asagiri, id/version/hash, prompt orchestré (`agentcontext`), output contract
- **Registry** — mise à jour `spec.external.last_synced_hash` après write
- **Doctor** — drift `external_drift` / `external_missing` → hint sync write
- **Tests** — unit + CLI (dry-run, write, conflict, force, JSON pur)
- **Docs CLI** — `agents-external-sync.mdx` regénéré

### Invariants T22

- Pas de scan `$HOME` ; pas d'installation implicite Cursor/Kiro/Codex
- Pas d'exécution agent ; workflow / gates / trust **inchangés**
- `--json` → stdout JSON pur ; hints sur stderr

### Verdict T22

**GO** — sync provider externe opt-in opérationnelle.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T22.

---

## Agent Orchestration Platform V1 — T21 (external audit read-only)

### Livré

- **`agentspec.external`** — `kind`, `path`, `last_synced_hash` ; hash sémantique sur kind/path
- **`config.agents.external_path`** — chemin explicite optionnel (pas de scan `$HOME`)
- **`internal/agentexternal`** — `Audit`, `ExternalTarget`, `Report` (`agents-external-v1`)
- **CLI** — `asa agents external [--json]` (read-only)
- **Détection sûre** — CLI `PATH`, chemins `spec.external.path` / `external_path`, stat fichier + `writable` sans écriture
- **Politique** — sync provider **opt-in** ; jamais automatique (chemins utilisateur, pas de mutation outil sans flag futur)
- **Tests** — unit `agentexternal` + CLI JSON/texte + hash external
- **Docs CLI** — `agents-external.mdx` regénéré

### Invariants T21

- Pas de write Cursor/Kiro/Codex ; pas d'installation provider
- Workflow / gates / trust / scoring **inchangés**
- `--json` → stdout JSON pur

### Verdict T21

**GO** — audit externe read-only prêt pour sync provider opt-in ultérieur.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T21.

---

## Agent Orchestration Platform V1 — T20 (agents sync + onboarding)

### Livré

- **`internal/agentsync`** — `Sync`, actions `create|update|skip|conflict`, `FormatJSON`/`FormatText`
- **CLI** — `asa agents sync [--json] [--check] [--write] [--force] [--agent <id>]`
- **Mode safe** — dry-run par défaut ; écriture `--write` ; conflit hash sans `--force` (exit non-zero)
- **Source** — templates embarqués → `.asagiri/agents/*.yaml` uniquement
- **Doctor / onboarding** — `checkAgentRegistry`, hints drift → `asa agents sync --write` ; wizard step `agents`
- **Tests** — unitaires `agentsync` + CLI + doctor drift cleared
- **Docs CLI** — `agents-sync.mdx` regénéré

### Invariants T20

- Pas de push/pull Cursor/Kiro/Codex ; pas d'installation provider
- Workflow / gates / trust / scoring **inchangés**
- `--json` → stdout JSON pur ; hints humains sur stderr si besoin

---

### Livré

- **`internal/agentslist`** — `Build`, `Show`, `FormatJSON`, `FormatText`
- **CLI** — `asa agents list [--json]`, `asa agents show <id> [--json]` (read-only)
- **Champs** — id, role, version, content_hash, stored_hash, source (disk|embedded), output_format, provider_targets, provider_support
- **Hash** — `SemanticHash` stable (tests reorder YAML + golden snapshot)
- **Docs CLI** — regénération `agents-list.mdx`, `agents-show.mdx`

### Invariants T19

- Pas de sync, pas d'installation provider
- Workflow / gates / trust **inchangés**
- `asa agents watch` (TUI) conservé

### Verdict T20

**GO** — `asa agents sync` + onboarding (explicitement hors scope immédiat avant T19 vert).

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T20.

---

## Agent Orchestration Platform V1 — T18 (Output contract validation)

### Livré

- **`internal/agentcontract`** — `ValidateOutput`, `WriteLog`, `ContractValidationResult`
- **Formats** — `text`, `asagiri-v1`, `gate-yaml`, `gate-json` ; erreurs typées
- **Hook dev orchestré** — après `agent.Run`, écrit `contract.json` ; ne bloque pas le workflow
- **`devresolve.Result`** — expose `Spec` + `AgentID` quand orchestré
- **Tests** — unitaires par format + golden `contract_valid.json` ; intégration `devOneTask`

### Invariants T18

- Gates / trust / scoring / `ParseResult` / `WriteLogs` **inchangés**
- Validation **log-only** ; fallback legacy sans `contract.json`
- Pas de hook gates agents (reporté)

### Verdict T19

**GO** — `internal/agentslist` + `asa agents list/show`.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T19.

---

## Agent Orchestration Platform V1 — T17 (Doctor drift detection)

### Livré

- **`internal/doctor/agent_specs.go`** — collecte registry, specs, drift, dernier contexte orchestré
- **JSON** — `agent_registry`, `agent_specs`, `agent_drift`, `last_orchestrated`
- **Texte** — sections `Agent specs`, `Agent drift`, next actions (`asa agents sync` hint)
- **`AgentInfo` enrichi** — `spec_version`, `content_hash`, `drift`, `prompt_source`, `output_format`
- **Checks** — spec manquante/invalide, hash mismatch, provider unsupported, commande absente
- **Registry absent** → info OK (templates embarqués) ; WARN si agent config sans spec disque

### Invariants T17

- Read-only : pas de workflow, gates, sync provider
- Pas de commande `asa agents sync` implémentée

### Verdict T18

**GO** — `internal/agentcontract` + persistance `contract.json` (dev orchestré).

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T18.

---

## Agent Orchestration Platform V1 — T16 (Dev prompt resolve)

### Livré

- **`internal/devresolve`** — `Resolve`, `LegacyDevPrompt`, `LoadDiskOnly` via `agentspec`
- **`devOneTask`** — prompt final via AgentSpec + `agentcontext` + `agentadapter` ; fallback legacy exact
- **Logs** — `context.json`, `prompt.md`, `invocation.json`, `resolve.json` ; warning `dev-orchestration.warn` si fallback
- **Tests** — orchestré / fallback / logs / snapshot `prompt.golden`

### Invariants T16

- Gates / trust / scoring / `ParseResult` **inchangés**
- Pas de mutation provider ; pas de `asa agents sync` ; pas d'exécution adapter supplémentaire
- Hook **dev uniquement** — gates agents reportés

### Verdict T17

**GO** — extension `doctor` drift registry ↔ config ↔ filesystem.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T17.

---

## Agent Orchestration Platform V1 — T15 (Provider adapters)

### Livré

- **`internal/agentadapter`** — render-only : `Adapter`, `RenderedInvocation`, `SupportLevel`
- Adapters **cursor**, **kiro**, **claude-code**, **codex** + fallback **inline** (exec/ollama/gemini)
- **`RenderFromConfig`** — `provider.type` depuis config + prompt orchestré sur stdin
- **`SupportMatrix`** / **`Explain`** — matrice support par provider
- Pas d'exécution subprocess ; `internal/agent` inchangé

### Invariants T15

- Workflow / gates / trust **inchangés**
- Pas de `asa agents sync` ; pas d'installation profils provider

---

## Agent Orchestration Platform V1 — T14 (ExecutionContext)

### Livré

- **`internal/agentcontext`** — `types.go`, `builder.go`, `render.go`, `hash.go`
- **`Build(Input)`** — AgentSpec + feature/task/run/phase + prompt métier
- **`RenderPrompt`** — mode orchestré, scope, gates/handoff, output_contract
- **`WriteLogs`** — `context.json` + `prompt.md` sous `.asagiri/logs/<task>/agents/<agent>/`
- **`ContextHash`** — empreinte stable du contexte d’exécution

### Invariants T14

- Workflow / exécution agents **inchangés** (pas de hook `devOneTask`)
- Gates/trust **inchangés**
- `internal/agent.WriteLogs` (legacy) conservé

### Verdict T15

**GO** — adapters `kirocli` / `cursorcli` / `codexcli` + factory ; pas de changement gates.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T15.

---

## Agent Orchestration Platform V1 — T13 (AgentSpec + registry)

### Livré

- **`internal/agentspec`** — `types.go`, `loader.go`, `validate.go`, `hash.go`
- **Modèle** — id, version, role, provider_targets, system_prompt, instructions, constraints, output_contract, metadata
- **Loader** — `.asagiri/agents/*.yaml` ; fallback templates embarqués (dev, reviewer, enricher, governance)
- **`SemanticHash`** — SHA-256 stable ; ignore timestamps metadata
- **Templates** — `internal/agentspec/templates/` + `.asagiri/agents/`
- **ADR-036** — `docs/decisions/036-agent-orchestration-platform.md`

### Invariants T13

- Workflow / exécution agents **inchangés** (aucun import `agentspec` hors package)
- Gates/trust **inchangés**
- Pas de `asa agents sync`

### Verdict T14

**GO** — `ExecutionContext` + logs ; hook `devOneTask` uniquement, gates intouchés.

Spec : `.kiro/specs/agent-orchestration-platform-v1/tasks.md` section T14.

---

## Trust Engine V1 — T12 (Trust Gate work)

### Livré

- **`work.gates.trust`** — gate désactivée par défaut ; seuils `min_score`, `block_verdicts`, `warn_verdicts`
- **`worktrust.WorkTrustReportToGateResult`** — mapping vers `gates.Result` sans modifier le scorer
- **Workflow** — hook post-`verify_evidence` ; persistance `gates.history` + logs ; guard review
- **`gates.TrustGateBlocksReview`** — FAIL/WARN non-advisory bloque review ; task reste `verified`
- **Intent** — `TrustGateBlocksReview` sur snapshot / `asa next`
- **ADR-035** — `docs/decisions/035-trust-gate-work.md`

### Invariants T12

- `internal/trust` (spec-my-B) inchangé
- Scoring `worktrust` inchangé
- Trust advisory CLI distinct de l’enforcement gate

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T12.

---

## Trust Engine V1 — T11 (Report history & diff)

### Livré

- **`reportsink`** — archive optionnelle vers `history/` avant overwrite ; `latest.json` inchangé
- **`internal/reportdiff`** — deltas score, verdict, dimensions, next-action
- **CLI** — `asa doctor diff`, `asa trust diff task|feature|run` (texte + `--json`)
- **ADR-034** — `docs/decisions/034-report-history-diff.md`

### Invariants T11

- Écriture limitée aux artefacts `.asagiri/reports/` (pas de mutation workflow)
- Aucun changement scoring / gates / workflow
- Pas de relecture auto des snapshots pour planner/scorer

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T11.

---

## Trust Engine V1 — T10 (Align recommendations)

### Livré

- **`intent.RecommendForTask`** / **`RecommendNextFromTasks`** — même logique que `RecommendNext`, focus task ou liste SQLite
- **`worktrust/recommend.go`** — `buildRecommendation` et `NextActions` feature/run délèguent à intent
- **Tests** — `recommend_align_test.go` (intent + worktrust) : parité commande trust vs `asa next`

### Invariants T10

- Aucun changement scoring / gates / workflow / reportsink
- Pas de nouvelle commande CLI

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T10.

---

## Trust Engine V1 — T9 (Operator & CI UX)

### Livré

- **`doctor.go`** — `SilenceErrors` (pas de double message stderr en `--strict`)
- **docs-site** — `daily-workflow` (en, fr, de, es) : CI `--json --save`, `jq`, chemins `.asagiri/reports/`, rappel stdout/stderr
- **`02-architecture.md`** — section Trust & Diagnostic (worktrust, reportsink, doctor, ADR-033)
- **`current-spec.md`** — phase T9 livrée

### Invariants T9

- Aucun changement workflow / gates / scoring / chemins reportsink
- Pas de nouvelle commande

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T9.

---

## Trust Engine V1 — T8 (persistent reports)

### Livré

- **`--save`** sur `asa trust task|feature|run` et `asa doctor`
- Package **`internal/reportsink`** — écriture JSON atomique
- Chemins :
  - `.asagiri/reports/trust/tasks/<task-id>.json`
  - `.asagiri/reports/trust/features/<feature>.json`
  - `.asagiri/reports/trust/runs/<run-id>.json`
  - `.asagiri/reports/doctor/latest.json`
- Stdout : rapport (texte ou JSON selon `--json`)
- Stderr : `Rapport enregistré : <chemin>` quand `--save`
- `--json --save` : JSON **pur** sur stdout + fichier ; confirmation sur stderr (convention Unix / `jq`)

### Invariants T8

- Read-only workflow ; seule écriture = artefact rapport
- Pas de DB, pas de cache implicite, pas de relecture auto des rapports
- Aucun changement gates/scoring/workflow

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T8.

---

## Trust Engine V1 — T7 (UX documentation)

### Livré

- **docs-site** : [`concepts/mental-model`](/docs/concepts/mental-model) — pipeline, gates, `next` / `work` / `continue` / `trust` / `doctor`
- **docs-site** : [`workflows/daily-workflow`](/docs/workflows/daily-workflow) — scénarios (feature saine, human_review, verify_evidence, doctor warn-only)
- **Locales** : en, fr, de, es + liens depuis `guided-path`
- **CLI** : `asa --help` mentionne mental-model et daily-workflow ; doctor suggère `asa trust feature` si tasks à risque

### Invariants T7

- Aucun changement workflow / gates / scoring
- Pas de nouvelle commande

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T7.

---

## Trust Engine V1 — T6 (UX onboarding / doctor)

### Livré

- **`asa doctor [--full] [--json] [--strict]`** — rapport structuré read-only
- Sections : Repository, State, Gates, Agents, Trust, Checks, Avertissements, Prochaines actions
- **`ready`** : true si aucun FAIL bloquant ; WARN n'impacte pas l'exit code par défaut
- **`--strict`** : exit code 1 si des avertissements sont présents (CI strict)
- JSON : champs `ready`, `warnings`, `failures` (+ `checks` complet)
- Checks onboarding via `onboarding.RunDoctorChecks` (--full), sans duplication

### Invariants T6

- Read-only : aucun changement workflow / gates / scoring
- Pas de Trust Gate, pas de cache `--save`

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T6.

---

## Trust Engine V1 — T5 (Trust daily UX)

### Livré

- **`asa next`** : bloc Trust compact (verdict, risque principal, commande) si task courante ; `--no-trust`
- **`asa status`** : section Trust pour feature active (verdict, tasks à risque, next action) ; `--no-trust`
- **`asa work` / `asa continue`** : une ligne trust après exécution réelle (pas dry-run / plan-only)
- Fichiers : `worktrust/daily.go`, `cli/trust_daily.go`, tests `daily_test.go`, `trust_daily_test.go`
- Docs CLI regénérées (`next.mdx`, `status.mdx`)

### Invariants T5

- Read-only : aucun changement gates / workflow / scoring
- Pas de Trust Gate, pas de cache `--save`
- JSON `asa trust * --json` inchangé

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T5.

---

## Trust Engine V1 — T4 (UX polish)

### Livré

- **Sections stables** : Summary, Gates, Risks, Next actions
- **Verdicts humains** (Fiable, Acceptable, À surveiller, Bloqué) — score masqué par défaut
- **`--explain`** : dimensions détaillées (task) / scores par task (feature/run)
- **Harmonisation** : recommandations agrégées → `asa next --feature <feature>` ; enrich avec `--agent`
- **Snapshots** : `worktrust/testdata/formatter_*.txt`

### Commandes

```bash
asa trust task <task-id> [--json] [--explain]
asa trust feature <feature> [--json] [--explain]
asa trust run <run-id> [--json] [--explain]
```

### Invariants T4

- Aucun changement workflow / gates / Trust Gate
- Pas de `--save` / cache
- JSON report inchangé (schema v1)

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T4.

---

## Trust Engine V1 — T3 (CLI run)

### Livré

- **`asa trust run <run-id> [--json]`** — charge run + tasks SQLite, plan gate log, agrégation blocked/risky
- **`RunTrustReport`** + **`BuildRunReport(repoRoot, cfg, store, runID)`**
- **`FormatRunReport`** — score, feature, tasks, plan gate, next actions
- Tests : `worktrust/run_test.go`, `cli/trust_work_cmd_test.go` (run texte/json, inconnu, sans task, agrégation)

### Commandes (arborescence)

```
asa verify trust <flow>     # spec-my-B — inchangé
asa trust gates             # config verification — inchangé
asa trust replay <id>       # spec-my-B — inchangé
asa trust task <task-id>    # work synthesis
asa trust feature <feature> # work synthesis
asa trust run <run-id>      # NEW T3 work synthesis
```

### Invariants T3

- Aucun changement workflow / gates runtime / PendingGate / Trust Gate bloquante
- Pas de cache `--save`
- Pas de refactor `internal/trust`

### Backlog post-T3 (optionnel)

- Config `trust_synthesis:` thresholds
- ADR-033, `docs/ai/02-architecture.md`
- Trust Gate work bloquante (spec future)

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T3.

---

## Trust Engine V1 — T2 (CLI task + feature)

### Livré

- **`asa trust task <task-id> [--json]`** — charge SQLite, `BuildTaskReport`, texte ou JSON
- **`asa trust feature <feature> [--json]`** — agrège tasks : score moyen, verdict worst-case, tasks at risk
- Fichiers : `cli/trust_work_cmd.go`, `worktrust/feature.go`, tests CLI + feature

### Commandes (arborescence)

```
asa verify trust <flow>     # spec-my-B — inchangé
asa trust gates             # config verification — inchangé
asa trust replay <id>       # spec-my-B — inchangé
asa trust task <task-id>    # NEW work synthesis
asa trust feature <feature> # NEW work synthesis
```

### Invariants T2

- Aucun changement workflow / gates runtime / PendingGate
- Pas de `asa trust run` (T3)
- Pas de cache `--save`

### Prochaine (T3)

~~`asa trust run <run-id>`~~ — **livré**.

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T3.

---

## Trust Engine V1 — T1 (core read-only)

### Livré

- Package **`internal/worktrust`** : modèle, reader, scorer, formatter, tests
- **`BuildTaskReport(repoRoot, cfg, task)`** — agrégation gates.history, validation JSON, PendingGate HR
- **`FormatTaskReport`** — sortie terminal (sans CLI)
- **9 tests** — no gates, enrich pass, verify_evidence+validation, HR pending blocked, gate fail, verify_failed/failed, formatter

### Fichiers T1

| Fichier | Rôle |
|---------|------|
| `application/internal/worktrust/types.go` | Modèle V1 |
| `application/internal/worktrust/reader.go` | Collecte + recommendation |
| `application/internal/worktrust/scorer.go` | Dimensions + verdict |
| `application/internal/worktrust/formatter.go` | Format terminal |
| `application/internal/worktrust/worktrust_test.go` | Tests |

### Invariants T1

- **Aucun** changement workflow, gates runtime, PendingGate, CLI
- Lecture seule — pas de mutation payload/SQLite
- Coexistence `internal/trust` (spec-my-B) inchangée

### Prochaine (T2)

- `asa trust task <id> [--json]`, `asa trust feature <feature>`
- Extension `cli/trust_cmd.go` ou `trust_work_cmd.go`

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T2.

---

## Trust Engine V1 — Phase 0 (spec only)

### Objectif

Synthèse UX-first **read-only** post-work-gates : score, verdict, dimensions, risques, prochaine action CLI.

**Distinct** de `internal/trust` (spec-my-B — `asa verify trust`, checks product).

### Livrables Phase 0

- [x] Audit packages `gates`, `trust`, `report`, `workflow`, `pkg/asagiri`
- [x] Spec Kiro `.kiro/specs/trust-engine-v1/` (requirements, design, tasks)
- [x] Modèle `WorkTrustReport` + 6 dimensions + sources V1 figées
- [x] CLI cible : `asa trust task|feature|run` (cohabite `trust gates|replay`)
- [x] Découpage T1 / T2 / T3 ; **verdict GO** pour T1
- [x] `current-spec.md` synchronisé

### Fichiers autorisés Phase 0 (done)

| Action | Chemins |
|--------|---------|
| Créer | `.kiro/specs/trust-engine-v1/*` |
| Modifier | `docs/ai/active/current-spec.md`, `docs/ai/active/handoff.md` |

### Fichiers interdits Phase 0

- `application/**/*.go` (production ou tests)
- Workflow, gates runtime, executor, CLI impl

### Prochaine (T3)

1. ~~CLI `asa trust run <run-id>` + `--json`~~ — **livré**
2. Doc docs-site CLI — **livré** (generate-cli)

~~T2 CLI task/feature~~ — **livré**.

Spec : `.kiro/specs/trust-engine-v1/tasks.md` section T3.

### Hors scope V1 (rappel)

Trust Gate bloquante, checks spec-my-B, git global, KG, TUI, persistance `.asagiri/trust/` obligatoire.

---

## Livraison Snapshot Refresh Executor (Phase 3.3)

### Problème

`Executor.Execute` capturait `FeatureState` une seule fois au début du plan. Après une step mutante (`enrich`, `dev`, `verify`, …), `EvaluateCondition` utilisait un état périmé → enchaînements gate-aware nécessitaient souvent une **deuxième invocation** (`asa work` / `asa continue --yes`).

### Solution (minimal diff)

- **`RefreshFeatureTaskState`** dans `internal/intent/snapshot.go` — recharge depuis SQLite + guards `gates.*` pour la feature courante
- **`Executor`** : champs `RepoRoot`, `Store` ; après step réussie parmi `plan|enrich|dev|verify|review`, refresh de `fs` in-memory
- **Branché** dans `cli/work.go` et `ui/bus/handlers_commands.go` — **aucun changement flags CLI**

### Enchaînements corrigés

| Avant (1 invocation) | Après |
|----------------------|-------|
| enrich OK → dev skipped (`EnrichGateBlocksDev` stale) | enrich → dev |
| verify OK → review skipped (`VerifyEvidenceGateBlocksReview` stale) | verify → review |
| HR resume dev OK → verify skipped (`PendingGate` stale) | dev → verify |

### Invariants préservés

- Guards runtime workflow (`devOneTask`, `VerifyFeature`, `ReviewFeature`) restent source de vérité à l'exécution
- `continue --yes` refusé si HR **submit** pending (non-régression CLI)
- Refresh **après** step seulement — pas avant chaque condition
- Sans `Store`/`RepoRoot` sur l'executor : comportement legacy (pas de refresh)

### Tests

- `internal/intent/executor_snapshot_refresh_test.go` (package `intent_test`)
- `internal/intent/snapshot_refresh_test.go`

Spec : `.kiro/specs/work-gates-model/` Phase 3.3

## Livraison Verify Evidence Gate (Phase 5 — MVP complet)

### Config & runtime (5.1–5.4)

- `work.gates.verify_evidence` : **disabled par défaut**, mode `per-task`, agent → `GovernanceAgent()` fallback
- `fail_on` default : codes parse/validation evidence (voir `verify_evidence_gate_config.go`)
- Module `verify_evidence_gate.go` : prompt read-only, parse/classify, persist `gates.history` + logs `verify_evidence.{json,log}`
- `runVerification` retourne `([]validation.Result, error)` ; `persistValidationEvidence` → `.asagiri/logs/<task-id>/validation/results.json`

### Pipeline VerifyFeature (5.5)

Ordre : HR guard → `runVerification` → persist evidence → (validation KO → **`verify_failed`**, gate skipped) → `processVerifyEvidenceGate` → transition **`verified`**.

- Gate FAIL / WARN non-advisory / parse error → task reste **`implemented`**, step verify failed
- Gate inactive → comportement legacy (transition verified directe après validation OK)
- Reload task depuis store avant transition (préserve `gates.history`)

### Anti-bypass review & orchestration (5.6)

- **`gates.VerifyEvidenceGateSatisfied`** : dernière entrée `gate == "verify_evidence"` ; pass OK ; warn OK si `warn_is_advisory=true`
- **`gates.VerifyEvidenceGateBlocksReview`** : task **`verified`** + gate active + non satisfait
- **`ReviewFeature`** : HR pending prioritaire, puis erreur — `verify evidence gate required before review: run asa verify <feature> --task <id> --force`
- **`RecommendNext`** / planner `review_enabled` : verify `--force` si gate non satisfaite sur `verified`

### Reprise

```bash
asa verify <feature> --task <id> --force
```

### Report (5.7)

- Entrées `gates.history` avec `gate: verify_evidence` visibles dans report `## Gates` (via payload générique)
- Fallback log `.asagiri/logs/<task-id>/gates/verify_evidence.json` si payload sans history verify_evidence

### Invariants Verify Evidence Gate

- **Pas de `PendingGate`** pour verify_evidence
- **Pas de `asa gates submit verify_evidence`**
- **Pas de nouveau statut métier** dérivé de la gate
- Validation command fail → **`verify_failed`** (gate non exécutée)
- Gate fail → **`implemented`** (pas verify_failed)

Spec : `.kiro/specs/verify-evidence-gate/`

## Pipeline work gates (référence)

```
plan → enrich → dev (+ governance + human_review) → verify (+ verify_evidence) → review
```

| Gate | PendingGate | gates submit | Statut après FAIL |
|------|-------------|--------------|-------------------|
| plan | non | non | step plan failed |
| enrich | non | non | `planned` |
| governance | non | non | retry / policy |
| human_review | **oui** | `human_review` | pending |
| verify_evidence | non | non | `implemented` |

## Livraison workflow guards (Phase 3.2)

- **`VerifyFeature`** / **`ReviewFeature`** : refuse si `BlockingPendingForTask` (Human Review pending)
- **`devOneTask`** : garde enrich pour `planned`, `pending`, `enriched`

## Livraison Enrich Gate (Phase 4)

- Gate post-enrich / pre-dev ; FAIL → `planned` ; reprise `asa enrich … --force`
- Report : payload `gates.history` + fallback log `enrich.json`

## Livraison Phase 3.1 (gate-aware orchestration)

- Modèle **`PendingGate`** — **Human Review uniquement**
- **`RecommendNext`**, **`continue`**, **`work`** : gate-aware pour pending gates bloquantes

## Limitation levée (Phase 3.3)

~~Snapshot `PendingGate` figé au début de `Executor.Execute` — pas de rafraîchissement mid-run.~~ → **refresh ciblé post-step** (voir section Phase 3.3 ci-dessus).

## Livraison Phase 2 (clôturée)

- `gates.LogDocument`, `GateHistoryEntry` / `TaskGates`, report `## Gates` via `gatesMarkdown()`

Références : `.kiro/specs/work-gates-model/`, ADR-032.

## Invariants globaux (ne pas régresser)

- **`gates.Result`** type canonique ; **`gates.Verdict`** = `pass|warn|fail`
- Config **`work.gates.*`** canonique ; legacy `work.governance` compat (ADR-031)
- Logs gate : `.asagiri/logs/<scope-id>/gates/<gate>.{json,log}`

## Definition of Done — Snapshot Refresh Executor

- [x] Audit point de figement (`Executor.Execute` ligne `fs := featureState(...)`)
- [x] `RefreshFeatureTaskState` + factorisation `applyTaskGateFields`
- [x] Refresh post-step mutante sans changement contrat CLI
- [x] Tests enrich→dev, verify→review, HR resume→verify, gates off, submit pending
- [x] `go build ./...` + `go test ./... -count=1` verts
- [x] `current-spec.md`, `handoff.md`, `.kiro/specs/work-gates-model/tasks.md` synchronisés

## Definition of Done — Verify Evidence Gate MVP

- [x] Config opt-in ; gate active mode per-task uniquement
- [x] Gate post-validation OK, pre-commit verified
- [x] Validation KO → verify_failed ; gate skipped
- [x] Gate FAIL → implemented ; PASS/WARN advisory → verified
- [x] Reprise `asa verify … --force` documentée
- [x] Review anti-bypass + RecommendNext cohérent
- [x] Pas de PendingGate verify_evidence ; pas de `gates submit verify_evidence`
- [x] Report `## Gates` inclut verify_evidence (payload + fallback log)
- [x] Tests verts ; handoff + current-spec synchronisés

## Quality_Gate (commande reproductible)

```bash
cd application && go build ./... && go test ./... -count=1
```

## Références

- `.kiro/specs/trust-engine-v1/` (Phase 0 spec — **en cours implémentation T1**)
- `.kiro/specs/verify-evidence-gate/` (Phase 5 MVP)
- `.kiro/specs/enrich-gate/` (Phase 4)
- `.kiro/specs/human-review-gate/` (Phase 3 + 3.1)
- `.kiro/specs/work-gates-model/` (Phase 2)
- `application/internal/gates/verify_evidence_satisfied.go`, `internal/workflow/verify_evidence_gate.go`, `internal/report/gates_report.go`
