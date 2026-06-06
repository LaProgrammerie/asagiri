# Implementation Plan: task-validation-gates (Tranche A)

> Feature : **task-validation-gates**
> Nature : couche gouvernance workflow — **opt-in**, inline dans `DevFeature`.
> **État :** **Tranche A livrée** (`2026-06-06`, ADR-031).

## Overview

Plan incrémental pour livrer la Tranche A uniquement. Ordre :

1. Modèle config + defaults + example
2. Verdict parse + classify
3. Gate runner (dry-run + live)
4. Hook DevFeature + retry + state machine
5. Trace payload / logs / report
6. Tests + doc canon + ADR-031

**Interdit** sans MAJ spec : smart mode, nouvelle commande CLI, execution graph, UI, refonte review/verify.

## Tasks

- [x] **1. Config `work.governance` (R1)**
  - [x] 1.1 Ajouter `WorkGovernanceConfig` + champ `WorkConfig.Governance` dans `config.go`
  - [x] 1.2 Defaults dans `applyDefaults` / helpers : `enabled=false`, `mode=off`, `max_retries=2`, `warn_is_advisory=true`
  - [x] 1.3 `Config.GovernanceAgent()` + `WorkGovernanceConfig.IsActive()`
  - [x] 1.4 Documenter bloc (commenté/disabled) dans `.asagiri/config.yaml.example`
  - [x] 1.5 Tests : `config/governance_config_test.go` — defaults, IsActive, agent resolution (reviewer, architect explicite)
  - _Requirements: R1, R6_

- [x] **2. Modèle verdict + parse (R3)**
  - [x] 2.1 Créer `workflow/governance_parse.go` — structs `GovernanceVerdict`, `GovernanceFinding`, `ClassifyVerdict`
  - [x] 2.2 Parser stdout (YAML/JSON), normalisation status, clamp confidence
  - [x] 2.3 Parse error → verdict fail + `parse_error`
  - [x] 2.4 Tests : `governance_parse_test.go` — pass/warn/fail, fail_on filter, alias PASS/WARN/FAIL, invalid → fail, JSON
  - _Requirements: R3.3–R3.5_

- [x] **3. Extension payload tâche (R5)**
  - [x] 3.1 Ajouter `Governance *TaskGovernance` (ou équivalent) dans `pkg/asagiri/types.go`
  - [x] 3.2 Helpers `persistGovernanceVerdict`, `governanceRetries`, `incrementGovernanceRetries` dans workflow
  - [x] 3.3 Sérialisation round-trip payload JSON/YAML tâche (`pkg/asagiri/governance_test.go`)
  - [x] 3.4 Test : historique appendé sur pass/warn/fail
  - _Requirements: R5.1_

- [x] **4. Gate runner (R3, R4, dry-run)**
  - [x] 4.1 Créer `workflow/governance.go` — `runGovernanceGate`, prompt builder, `gitDiffForGovernance` helper
  - [x] 4.2 Dry-run : verdict simulé pass sans appel agent
  - [x] 4.3 Live : `GovernanceAgent()`, `ensureAgent`, `Run`, parse, classify
  - [x] 4.4 Écrire `.asagiri/logs/<taskID>/governance.log` + `governance.json`
  - [x] 4.5 Tests unitaires runner avec agent mock / dry-run service
  - _Requirements: R3, R4.5, NFR1_

- [x] **5. Intégration DevFeature (R2, R4) — A.2**
  - [x] 5.1 Extraire `devOneTask(...)` depuis boucle `DevFeature` (refactor minimal)
  - [x] 5.2 Après `StatusImplemented`, appeler gate si `IsActive()` (`devTaskWithGovernanceRetries`)
  - [x] 5.3 PASS/WARN → continue ; WARN trace avec `warn_is_advisory`
  - [x] 5.4 FAIL → si `retries_used < max_retries` : consommer une relance, `implemented` → `running`, reboucler dev ;
    sinon `failed` + erreur step dev (passages governance max = `max_retries + 1`)
  - [x] 5.5 State machine : autoriser `implemented → running` (force) — retry governance uniquement
  - [x] 5.6 Tests workflow : governance off = régression nulle ; warn continue ; fail retry 0/1/2 ;
    history `retry` 0..n ; dry-run pass
  - _Requirements: R2, R4, R6.3_

- [x] **6. Report trace (R5.3)**
  - [x] 6.1 Section Governance dans report run (table ou liste par tâche)
  - [x] 6.2 Test : report contient entrée warn/fail quand gate exécutée
  - _Requirements: R5.3_

- [x] **7. Documentation & ADR**
  - [x] 7.1 `docs/ai/05-decisions.md` — **ADR-031** (governance gates Tranche A)
  - [x] 7.2 Mention courte dans `docs/ai/02-architecture.md` (section workflow)
  - [x] 7.3 Sync `current-spec.md` + `handoff.md` post-livraison Tranche A
  - _Requirements: Acceptance_

- [x] **8. Quality Gate (Acceptance)**
  - [x] 8.1 `make build && go vet ./... && go test ./...` exit 0
  - [x] 8.2 Vérifier configs existantes sans `work.governance` inchangées (test régression dédié)
  - _Requirements: NFR3, NFR4, AC7_

## Definition of Done (Tranche A)

- [x] `work.governance` config + defaults documentés
- [x] Modes `off` et `per-task` fonctionnels ; autres modes ignorés (off + warn log via `EnabledButInactive`)
- [x] Gate post-dev par tâche, validateur read-only, verdict structuré
- [x] FAIL → relances dev : `retries_used < max_retries` (max `max_retries + 1` passages governance) ;
  WARN advisory avec trace
- [x] Dry-run → PASS simulé
- [x] Tests parse + config + workflow state + retry
- [x] Aucune nouvelle Unitary_Command ; review/verify inchangés
- [x] ADR-031 enregistrée

## Files touched (expected)

| Area | Files |
|------|-------|
| Config | `application/internal/config/config.go`, `governance_config.go`, `.asagiri/config.yaml.example` |
| Workflow | `application/internal/workflow/governance.go`, `governance_parse.go`, `dev_task.go`, `workflow.go`, `state_machine.go` |
| Types | `application/pkg/asagiri/types.go`, `governance_test.go` |
| Report | `application/internal/report/report.go`, `governance_report_test.go` |
| Tests | `governance_*_test.go`, `governance_config_test.go`, `state_machine_test.go` |
| Docs | `docs/ai/05-decisions.md`, `02-architecture.md`, `active/*` |

## Explicitly NOT in this plan

- `asa governance` command
- `mode: smart | per-step | milestone`
- execution graph / trust integration
- onboarding wizard step / UI screens
- default agent `architect`
- migration configs utilisateur
