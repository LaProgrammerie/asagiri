# task-validation-gates — Design (Tranche A)

## Overview

Ajouter une **Governance_Gate** inline dans la boucle `DevFeature`, déclenchée
après chaque transition tâche → `implemented` lorsque `work.governance.enabled`
et `mode: per-task`.

Pas de nouvelle commande CLI. Pas de nouveau rôle coordination. Réutilisation
de l'infrastructure agent + logs + payload canonique existants.

```
DevFeature (par tâche)
  … agent.Run (implémentation)
  → transition implemented
  → [si governance per-task]
       runGovernanceGate(task)
       pass/warn → continue
       fail → relancer dev si relances restantes (< max_retries consommées) ou failed
  → tâche suivante
```

### Sémantique `max_retries` (canonique A.2)

| Champ / règle | Signification |
|---------------|---------------|
| `max_retries` | Relances autorisées **après le premier FAIL** |
| `governance.retries` | Relances **déjà consommées** |
| `history[].retry` | Numéro de tentative governance (0 = première évaluation) |
| Passages max | `max_retries + 1` |
| Condition retry | `retries_used < max_retries` → relance ; sinon `failed` |

Exemples :

- **`max_retries=0`** : attempt 0 FAIL → `failed` (1 passage).
- **`max_retries=1`** : attempt 0 FAIL → retry → attempt 1 FAIL → `failed` (2 passages).
- **`max_retries=2`** : attempt 0 FAIL → retry → attempt 1 FAIL → retry → attempt 2 FAIL → `failed` (3 passages).

**Ne pas confondre** : `max_retries` n'est pas le nombre total d'échecs avant blocage ni le nombre
de passages governance seul sans relance intermédiaire.

## Config model

### Types (`application/internal/config/config.go`)

```go
type WorkConfig struct {
    // … champs existants …
    Governance WorkGovernanceConfig `yaml:"governance"`
}

type WorkGovernanceConfig struct {
    Enabled         bool     `yaml:"enabled"`
    Mode            string   `yaml:"mode"` // off | per-task (Tranche A)
    Agent           string   `yaml:"agent"`
    FailOn          []string `yaml:"fail_on"`
    WarnIsAdvisory  *bool    `yaml:"warn_is_advisory"`
    MaxRetries      *int     `yaml:"max_retries"` // nil → défaut 2 ; 0 explicite = aucune relance
}

func (c *Config) GovernanceAgent() string {
    if c == nil {
        return DefaultAgentReviewer
    }
    if a := strings.TrimSpace(c.Work.Governance.Agent); a != "" {
        return a
    }
    return c.WorkReviewerAgent()
}

func (c WorkGovernanceConfig) IsActive() bool {
    return c.Enabled && strings.EqualFold(strings.TrimSpace(c.Mode), "per-task")
}
```

### Defaults (`applyDefaults` / `applyV3Defaults`)

```go
if g.MaxRetries == nil {
    v := 2
    g.MaxRetries = &v
}
// WarnIsAdvisory: *bool ; si nil → true
// max_retries: 0 explicite en YAML est conservé (aucune relance après le 1er FAIL)
```

### Example (`.asagiri/config.yaml.example`)

Bloc commenté ou disabled :

```yaml
work:
  governance:
    enabled: false
    mode: per-task
    agent: reviewer
    fail_on:
      - spec_drift
      - architecture_violation
      - unexpected_design_change
    warn_is_advisory: true
    max_retries: 2
```

## Package layout (Tranche A)

```
application/internal/workflow/
  governance.go          # RunGate, build prompt, apply verdict
  governance_parse.go    # Parse + normalize YAML/JSON verdict
  governance_test.go
  workflow.go            # hook dans DevFeature
```

Pas de `internal/governance/` séparé en Tranche A (éviter surcouche).

## Verdict model

```go
package workflow

type GovernanceFinding struct {
    Code     string   `json:"code" yaml:"code"`
    Severity string   `json:"severity" yaml:"severity"` // warn | fail
    Message  string   `json:"message" yaml:"message"`
    Actions  []string `json:"actions,omitempty" yaml:"actions,omitempty"`
}

type GovernanceVerdict struct {
    Status     string              `json:"status" yaml:"status"` // pass | warn | fail
    Confidence float64             `json:"confidence" yaml:"confidence"`
    Notes      []string            `json:"notes,omitempty" yaml:"notes,omitempty"`
    Findings   []GovernanceFinding `json:"findings,omitempty" yaml:"findings,omitempty"`
    DryRun     bool                `json:"dry_run,omitempty"`
    ParseError string              `json:"parse_error,omitempty"`
}

type GovernanceHistoryEntry struct {
    At         string              `json:"at"`
    Status     string              `json:"status"`
    Confidence float64             `json:"confidence"`
    Notes      []string            `json:"notes,omitempty"`
    Findings   []GovernanceFinding `json:"findings,omitempty"`
    Retry      int                 `json:"retry"`
    DryRun     bool                `json:"dry_run,omitempty"`
}

type GovernanceState struct {
    History []GovernanceHistoryEntry `json:"history,omitempty"`
    Retries int                      `json:"retries"` // relances consommées
}
```

`GovernanceHistoryEntry.Retry` = numéro de tentative governance (0 = première évaluation).

Stockage dans `asagiri.Task` — extension du struct canonique :

```go
// pkg/asagiri/types.go
type Task struct {
    // …
    Governance *TaskGovernance `json:"governance,omitempty" yaml:"governance,omitempty"`
}
```

Alternative minimale : champ opaque dans payload JSON sans modifier struct public —
**préférer extension struct** pour tests et report typés.

## Parser contract

1. Extraire bloc `governance:` depuis stdout agent (`agent.ParseResult` pattern existant).
2. Accepter YAML ou JSON équivalent.
3. Normaliser `status` → lowercase ; mapper alias `PASS`/`WARN`/`FAIL`.
4. Valider `confidence` ∈ [0,1] ; clamp si hors bornes + note.
5. Appliquer règles `fail_on` :

```go
func ClassifyVerdict(v GovernanceVerdict, failOn []string) string {
    blocking := failOnSet(failOn) // empty set = all fail severities block
    for _, f := range v.Findings {
        if f.Severity != "fail" {
            continue
        }
        if len(blocking) == 0 || blocking[f.Code] {
            return "fail"
        }
    }
    for _, f := range v.Findings {
        if f.Severity == "warn" {
            return "warn"
        }
    }
    return coalesceStatus(v.Status, "pass")
}
```

6. Parse error → verdict `fail` + `ParseError` rempli.

## Gate execution (`runGovernanceGate`)

```go
func (s *Service) runGovernanceGate(
    ctx context.Context,
    feature string,
    task sqlite.Task,
    worktreePath string,
    retry int,
) (GovernanceVerdict, error)
```

### Dry-run branch

```go
if s.dryRun {
    return GovernanceVerdict{
        Status: "pass",
        Confidence: 1,
        Notes: []string{"governance dry-run: simulated pass"},
        DryRun: true,
    }, nil
}
```

### Live branch

1. `agentName := s.cfg.GovernanceAgent()`
2. `ensureAgent(agentName)`
3. Construire prompt (template constante) avec :
   - `readSpecExcerpt(feature, taskID)`
   - `formatTaskCanonical(task)`
   - `gitDiff(worktreePath)` — `git diff` depuis worktree ; fallback `git diff HEAD` si clean
   - optional `readFile(docs/ai/02-architecture.md)` tronqué (max N chars)
4. Instruction stricte : répondre **uniquement** avec le bloc YAML `governance:`.
5. `a.Run(ctx, RunRequest{…})` — pas de write tools
6. `ParseGovernanceVerdict(stdout)`
7. `ClassifyVerdict`
8. Persister trace : payload + `governance.log`

## DevFeature integration

Point d'injection (après L442 `transitionTask(..., StatusImplemented)`):

```go
if s.cfg.Work.Governance.IsActive() {
    verdict, err := s.runGovernanceGate(ctx, feature, task, worktreePath, governanceRetries(task))
    if err != nil {
        // update step failed, return
    }
    if err := s.recordGovernanceVerdict(task, verdict); err != nil { … }

    switch verdict.Status {
    case "pass", "warn":
        // warn advisory: continue (R4.2)
        continue
    case "fail":
        used := governanceRetries(task) // relances déjà consommées
        max := s.maxGovernanceRetries()
        if used < max {
            incrementGovernanceRetries(task) // consomme une relance
            _ = s.transitionTask(task, asagiri.StatusRunning, true)
            // re-run dev loop body for same task (devTaskWithGovernanceRetries)
        } else {
            _ = s.transitionTask(task, asagiri.StatusFailed, true)
            _ = s.updateStep(run, "dev", sqlite.StatusFailed, formatGovernanceFailure(verdict))
            return run.ID, fmt.Errorf("governance gate failed after %d retries (max %d)", used, max)
        }
    }
}
```

**Refactor minimal** : extraire `devOneTask(ctx, run, feature, task, agentName, force)` pour
permettre retry sans duplication.

## State machine impact

Tranche A **n'ajoute pas** de statuts publics obligatoires (`governance_failed` optionnel —
**non retenu** pour minimiser la surface).

Transitions utilisées :

| Verdict | Transition |
|---------|------------|
| pass/warn | reste `implemented` |
| fail (retry) | `implemented` → `running` (relance governance uniquement) |
| fail (épuisé) | → `failed` |

`implemented → running` : **uniquement** pour relancer dev après un FAIL governance (Tranche A.2).
Pas de statut `governance_failed` en Tranche A.

## Report enrichment

Dans `report.Writer` ou section existante du report run :

```markdown
## Governance

| Task | Status | Confidence | Notes |
|------|--------|------------|-------|
| feat-001 | warn | 0.72 | API contract drift (advisory) |
```

Lecture depuis payload tâches du run — pas de nouvelle table SQLite Tranche A.

## Prompt template (sketch)

```
Tu es un validateur de gouvernance. Tu ne produis AUCUN code.
Analyse la tâche implémentée vs la spec et le diff.
Réponds UNIQUEMENT avec un bloc YAML:

governance:
  status: pass|warn|fail
  confidence: 0.0-1.0
  notes: [...]
  findings:
    - code: spec_drift|architecture_violation|unexpected_design_change|other
      severity: warn|fail
      message: ...
      actions: [...]

--- SPEC ---
{{specExcerpt}}

--- TASK ---
{{taskYAML}}

--- DIFF ---
{{gitDiff}}
```

## Error handling

| Cas | Comportement |
|-----|--------------|
| Agent absent | Erreur Guided_Remediation avant Run |
| Agent run error | step dev failed, tâche failed |
| Parse error | traité comme fail (retryable) |
| git diff error | note dans prompt « diff unavailable » ; gate continue |

## Testing strategy

| Test | Fichier | Objet |
|------|---------|-------|
| Parse YAML/JSON valide | `governance_parse_test.go` | ClassifyVerdict, alias PASS |
| Parse invalide → fail | idem | ParseError |
| Defaults config | `config/governance_test.go` | enabled false, max_retries |
| WARN advisory continue | `governance_test.go` | mock agent warn → implemented |
| FAIL retry then fail | `governance_retry_test.go` | `max_retries` 0/1/2, history retry 0..n, anti-boucle |
| Dry-run pass | `governance_test.go` | no agent call |
| IsActive off | workflow test | no hook |

Mocks : injecter interface `GovernanceRunner` ou stub `agent.Agent` via factory test hook —
**préférer** test parse + test workflow avec agent dry-run existant.

## Security / cost

- Prompt limité (troncature spec/diff/archi).
- Pas d'exécution de code agent côté validateur.
- Gate désactivée par default.

## Risks & mitigations

| Risque | Mitigation Tranche A |
|--------|---------------------|
| Faux positifs bloquants | WARN advisory ; max_retries ; opt-in |
| Parse fragile LLM | fail → retry ; template strict |
| Boucle retry infinie | `max_retries` fini ; au plus `max_retries + 1` gates ; comparaison `used < max` avant consommation |
| Transition state machine | test explicite implemented→running |
| Latence / coût | opt-in ; une gate par tâche seulement si activé |

## Future (hors Tranche A — ne pas implémenter)

- `mode: smart` (risk, files_changed, domains)
- `per-step`, `milestone`
- Nœud `NodeTypeGovernance` dans execution graph
- Intégration trust gates
- `asa governance` CLI
- UI dashboard

## ADR

Enregistrer **ADR-031** dans `docs/ai/05-decisions.md` à la livraison :
Governance gates Tranche A, config `work.governance`, agent default reviewer.
