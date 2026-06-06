# Requirements Document

> Feature : **task-validation-gates** (Tranche A)
> Nature : couche de **gouvernance workflow** légère — pas un nouveau rôle métier obligatoire.

## Introduction

Asagiri orchestre aujourd'hui un pipeline unitaire :

`spec → plan → enrich → dev → verify → review → report`

Les étapes **verify** (commandes locales) et **review** (agent) couvrent surtout la
**qualité technique** et la revue **tardive**. Elles ne protègent pas suffisamment
contre :

- la dérive fonctionnelle par rapport à la spec ;
- les violations d'architecture ou de patterns attendus ;
- les décisions de design non prévues ;
- la dette introduite progressivement sur plusieurs tâches.

Cette feature introduit une **gate de gouvernance read-only** exécutée **après
chaque tâche dev** (mode `per-task`), avec verdict structuré `PASS | WARN | FAIL`.
Elle réutilise un **agent logique existant** (default : `reviewer`) — pas de
nouveau rôle Asagiri obligatoire.

**Tranche A uniquement.** Les modes `smart`, `per-step`, `milestone`, l'intégration
execution graph / trust graph, une commande CLI top-level dédiée, l'UI dashboard et
toute refonte de `review` / `verify` sont **hors scope**.

## Décisions produit (figées Tranche A)

| # | Décision | Valeur |
|---|----------|--------|
| D1 | Placement config | `work.governance` |
| D2 | WARN | **Advisory par défaut** — le workflow continue, trace obligatoire |
| D3 | Agent MVP | **`reviewer`** par défaut ; `architect` autorisé si configuré explicitement |

## Glossary

- **Governance_Gate** : contrôle read-only post-dev par tâche ; ne produit pas de code.
- **Governance_Verdict** : `pass`, `warn`, ou `fail` (normalisé en minuscules).
- **Governance_Agent** : clé sous `config.agents` invoquée pour la gate (default `reviewer`).
- **Task_Payload** : JSON canonique de la tâche (`.asagiri/tasks/<feature>/<id>.json`).
- **Governance_Trace** : enregistrement structuré du verdict dans logs, payload tâche et report run.
- **Retry_Budget** : nombre de **relances autorisées après le premier FAIL** governance (`max_retries`).
  Passages governance max par tâche = `max_retries + 1` (1 tentative initiale + relances).
- **Governance_Retries_Used** : compteur `governance.retries` dans le payload — relances **déjà consommées**
  (incrémenté uniquement lorsqu'une relance dev est décidée).
- **Governance_Attempt** : numéro de tentative governance dans `history[].retry` (0 = première évaluation).
- **Unitary_Command** : commandes CLI existantes (`asa dev`, `asa verify`, …) — aucune nouvelle commande top-level Tranche A.

## Functional requirements

### R1 — Configuration `work.governance`

**R1.1** Le schéma YAML expose un bloc `work.governance` avec au minimum :

```yaml
work:
  governance:
    enabled: false
    mode: off          # Tranche A : off | per-task uniquement
    agent: reviewer    # agent logique ; vide → work.default_reviewer
    fail_on:
      - spec_drift
      - architecture_violation
      - unexpected_design_change
    warn_is_advisory: true
    max_retries: 2
```

**R1.2** Defaults à l'apply :

- `enabled: false` (opt-in global, compat configs existantes)
- `mode: off`
- `agent` résolu vers `work.default_reviewer` (aujourd'hui `reviewer` logique)
- `warn_is_advisory: true`
- `max_retries: 2`
- `fail_on` : liste vide acceptée (= toute finding `fail` bloque)

**R1.3** Si `enabled: false` ou `mode: off`, le pipeline est **strictement identique**
à l'existant (aucun appel agent, aucun changement d'état).

**R1.4** Si `agent` référence une clé absente de `config.agents`, erreur guidée
(Guided_Remediation) — pas de panic.

### R2 — Modes Tranche A

**R2.1** `mode: per-task` : après chaque tâche passant à `implemented` dans
`DevFeature`, exécuter la Governance_Gate **avant** de passer à la tâche suivante
ou de clôturer l'étape `dev`.

**R2.2** `mode: off` : gate ignorée (même si `enabled: true` sans mode valide → traiter
comme off + log warn config).

### R3 — Validateur read-only

**R3.1** Entrées minimales du prompt :

- extrait spec feature (requirements + tasks.md pertinents) ;
- tâche courante (titre, acceptance, scope) ;
- diff git depuis la branche/worktree de la tâche (ou `git diff` worktree vs base) ;
- optionnel Tranche A : chemin `docs/ai/02-architecture.md` s'il existe (pas de knowledge graph).

**R3.2** Le validateur **ne doit pas** écrire de fichiers ni lancer d'outils d'édition ;
mode agent existant en lecture seule (prompt contractuel).

**R3.3** Sortie agent **obligatoirement** parseable en structure :

```yaml
governance:
  status: pass | warn | fail
  confidence: 0.0-1.0
  notes:
    - string
  findings:
    - code: spec_drift | architecture_violation | unexpected_design_change | other
      severity: warn | fail
      message: string
      actions:
        - string
```

**R3.4** Mapping verdict :

- tout finding `severity: fail` dont `code ∈ fail_on` (ou `fail_on` vide) → `status: fail`
- sinon findings `warn` seuls → `status: warn`
- aucun finding bloquant → `status: pass`

**R3.5** Parse invalide / stdout vide → `status: fail`, note `governance_parse_error`
(retryable selon budget).

### R4 — Comportement workflow

**R4.1** `PASS` : tâche reste `implemented` ; append Governance_Trace ; continue.

**R4.2** `WARN` avec `warn_is_advisory: true` (default) : tâche reste `implemented` ;
append trace avec statut `warn` ; **continue** vers tâche suivante / fin dev.

**R4.3** `FAIL` — sémantique retry (Tranche A.2, canonique) :

- `max_retries` = nombre de **relances autorisées après le premier FAIL** governance.
- `governance.retries` = relances **déjà consommées** (incrémenté seulement si une relance est accordée).
- `history[].retry` = numéro de **tentative governance** (0 = première évaluation).
- Condition : si `retries_used < max_retries` → consommer une relance, transition tâche
  `implemented` → `running`, relancer dev **sur la même tâche** (même worktree si encore valide) ;
  sinon → transition → `failed`, step `dev` en échec, message actionnable listant findings + actions.
- **WARN advisory** et **PASS** ne consomment jamais de relance.
- Parse error / agent error = FAIL retryable (même budget).

Exemples :

| `max_retries` | Séquence |
|---------------|----------|
| `0` | attempt 0 FAIL → `failed` immédiatement (1 passage governance) |
| `1` | attempt 0 FAIL → retry 1 → attempt 1 FAIL → `failed` (2 passages) |
| `2` | attempt 0 FAIL → retry 1 → attempt 1 FAIL → retry 2 → attempt 2 FAIL → `failed` (3 passages) |

Anti-boucle : au plus `max_retries + 1` évaluations governance par tâche ; `max_retries` est fini et fixe.

**R4.4** La gate s'exécute **dans** `DevFeature` (pas de nouvelle Unitary_Command).
`ResumeRunExecute` hérite du comportement via `dev` inchangé côté CLI.

**R4.5** Dry-run (`service.dryRun == true`) : **verdict simulé `pass`**, trace
`governance_dry_run: true`, **aucun** appel agent externe.

### R5 — Traçabilité

**R5.1** Persister dans Task_Payload :

```json
"governance": {
  "history": [
    {
      "at": "RFC3339",
      "status": "pass|warn|fail",
      "confidence": 0.91,
      "notes": [],
      "findings": [],
      "retry": 0,
      "dry_run": false
    }
  ],
  "retries": 0
}
```

- `retries` : relances **consommées** (pas le nombre d'échecs ni le numéro de tentative en cours).
- `history[].retry` : index de **tentative governance** (0-based) pour cette entrée d'historique.

**R5.2** Écrire un log fichier `.asagiri/logs/<taskID>/governance.log` (stdout agent + verdict parsé).

**R5.3** Enrichir le report run existant (section gouvernance par tâche) sans refonte report.

### R6 — Compatibilité

**R6.1** Compatible tous providers via factory existante (`agentfactory.NewFromConfig`).

**R6.2** Aucun changement obligatoire des configs existantes.

**R6.3** `review` et `verify` restent inchangés dans leur sémantique et leur placement pipeline.

## Non-functional requirements

**NFR1** Faible coût tokens : prompt compact ; pas de relecture repo entier.

**NFR2** Déterminisme dry-run : même entrée → même trace simulée.

**NFR3** Pas de `panic` aux frontières ; erreurs retournées (`03-standards.md`).

**NFR4** Tests unitaires : parse verdict, defaults config, boucle retry, dry-run pass.

## Out of scope (Tranche A — interdit)

- Modes `smart`, `per-step`, `milestone`
- Nœud execution graph / trust graph
- Commande `asa governance` (top-level)
- Agent `architect` comme default obligatoire
- UI / dashboard / TUI
- Refonte `review`, `verify`, coordination, routing
- Migration automatique des configs utilisateur
- Nouveau package `internal/audit`

## Acceptance criteria (Tranche A)

1. Config example documente `work.governance` (disabled by default).
2. `enabled: false` → zéro régression tests workflow existants.
3. `per-task` + dry-run → PASS simulé, pipeline complet inchangé côté états finaux.
4. FAIL simulé (test) → jusqu'à `max_retries + 1` passages governance, puis `failed`
   (ex. `max_retries=2` → 3 FAIL consécutifs avant blocage).
5. WARN → tâche `implemented`, historique payload + log présents.
6. Agent configurable `architect` fonctionne si présent dans `agents:`.
7. Quality_Gate vert (`build`, `vet`, `test`).

## References

- `application/internal/workflow/workflow.go` — `DevFeature`, transitions tâche
- `application/internal/workflow/state_machine.go` — machine d'états
- `application/internal/config/config.go` — `WorkConfig`
- `docs/ai/03-standards.md` — erreurs, pas de panic
- ADR à créer en implémentation : **ADR-031** (governance gates Tranche A)
