# Spec — Asagiri Trust & Verification Engine

## 1. Vision

Le Trust & Verification Engine ajoute à Asagiri une couche explicite de validation, confiance et vérification déterministe.

Le but est d’empêcher les agents de produire des changements non vérifiés, incohérents ou dangereux.

Aujourd’hui, beaucoup de workflows IA suivent implicitement :

```text
prompt
  ↓
agent patch
  ↓
hope-driven validation
```

Évolution cible :

```text
intent
  ↓
flows/contracts/specs
  ↓
implementation
  ↓
trust checks
  ↓
confidence scoring
  ↓
verification gates
  ↓
review
  ↓
validated delivery
```

Le moteur Asagiri doit pouvoir mesurer :

* niveau de confiance ;
* qualité de validation ;
* cohérence architecture ;
* cohérence flow ;
* couverture observabilité ;
* couverture sécurité ;
* risque de blast radius ;
* stabilité potentielle.

Le système ne doit jamais présenter une confiance élevée sans évidence vérifiable.

---

## 2. Positionnement

Cette couche ne remplace pas :

* les tests ;
* la review humaine ;
* les validations CI ;
* les analyses statiques.

Elle orchestre et structure la confiance.

Le moteur doit devenir capable de répondre à :

```text
Pourquoi pense-t-on que ce changement est sûr ?
```

Et :

```text
Quels risques restent non couverts ?
```

---

## 3. Principes fondamentaux

### 3.1 Pas de confiance implicite

Toute confiance doit être liée à :

* tests ;
* flows ;
* contracts ;
* validations ;
* analyses ;
* reviews ;
* preuves runtime.

### 3.2 Confidence != correctness

Un score élevé n’est jamais une garantie.

Le moteur doit afficher :

* niveau de confiance ;
* limites ;
* zones non couvertes ;
* hypothèses restantes.

### 3.3 Les validations sont composables

La confiance finale doit être dérivée de plusieurs validations spécialisées.

### 3.4 Les risques doivent être visibles

Le moteur doit rendre explicites :

* blast radius ;
* dépendances critiques ;
* flows sensibles ;
* régressions possibles ;
* angles morts observabilité.

---

## 4. Verification Pipeline

Créer un pipeline de vérification structuré.

Pattern cible :

```text
Implementation
  ↓
Static validation
  ↓
Contract validation
  ↓
Flow validation
  ↓
Observability validation
  ↓
Security validation
  ↓
Architecture validation
  ↓
Cost validation
  ↓
Confidence scoring
  ↓
Review gates
```

---

## 5. Commandes CLI

### 5.1 `asa verify trust`

Usage :

```bash
asa verify trust onboarding-flow
```

Options :

```bash
asa verify trust onboarding-flow \
  --flow onboarding \
  --task task-003 \
  --branch onboarding-enterprise \
  --strict \
  --json \
  --ci
```

---

### 5.2 `asa trust gates`

Afficher les gates actives.

```bash
asa trust gates
```

---

### 5.3 `asa trust replay`

Rejouer une vérification.

```bash
asa trust replay trust-2026-05-27-001
```

---

## 6. Trust Report

Créer :

```text
.asagiri/trust/<id>/report.md
.asagiri/trust/<id>/report.json
```

Exemple :

```text
Trust Report
────────────

Architecture confidence:        82%
Implementation confidence:      67%
Flow integrity confidence:      91%
Observability confidence:       58%
Security confidence:            73%
Regression confidence:          64%

Warnings:
- invitation retry path not tested
- onboarding metrics incomplete
- no rate limit validation found
- async failure handling uncertain

Residual risk:
medium
```

---

## 7. Trust Dimensions

### 7.1 Architecture confidence

Mesure :

* cohérence contracts ;
* dépendances ;
* layering ;
* conventions ;
* blast radius ;
* dette introduite.

### 7.2 Implementation confidence

Mesure :

* validations passées ;
* qualité diff ;
* stabilité estimée ;
* complexité ;
* couverture tests.

### 7.3 Flow integrity confidence

Mesure :

* flows complets ;
* états présents ;
* transitions cohérentes ;
* permissions ;
* actions liées.

### 7.4 Observability confidence

Mesure :

* logs ;
* métriques ;
* traces ;
* dashboards ;
* alertes.

### 7.5 Security confidence

Mesure :

* auth ;
* permissions ;
* secrets ;
* inputs ;
* destructive actions ;
* rate limiting.

### 7.6 Regression confidence

Mesure :

* blast radius ;
* couverture ;
* flows impactés ;
* diff size ;
* dépendances critiques.

---

## 8. Verification Checks

Créer :

```text
internal/trust/checks/
```

Checks possibles :

```text
static-analysis
contracts
flows
permissions
observability
analytics
architecture
security
performance
cost
backward-compatibility
migration-safety
```

---

## 9. Check Model

```go
type VerificationCheck struct {
    ID          string
    Name        string
    Type        CheckType
    Status      CheckStatus
    Confidence  float64
    Findings    []Finding
    Evidence    []Evidence
    Duration    time.Duration
}
```

---

## 10. Finding Model

```go
type Finding struct {
    Severity     Severity
    Category     string
    Message      string
    Evidence     []Evidence
    SuggestedFix string
}
```

---

## 11. Confidence Engine

Créer :

```text
internal/trust/confidence/
  scoring.go
  weighting.go
  aggregation.go
  normalization.go
```

Le moteur doit agréger :

* checks ;
* tests ;
* reviews ;
* investigations ;
* runtime evidence ;
* flow criticality ;
* blast radius.

---

## 12. Blast Radius Analysis

Le moteur doit estimer l’impact potentiel.

Sources :

* graphes dépendances ;
* flows ;
* routes ;
* événements ;
* contrats ;
* modules critiques.

Exemple :

```text
Blast Radius
────────────
Flows impacted: 4
Critical APIs: 2
Shared modules: 6
Migration risk: medium
Public contract risk: high
```

---

## 13. Flow Integrity Validation

Créer validation dédiée flows.

Checks possibles :

* états loading/error/success ;
* transitions invalides ;
* actions orphelines ;
* permissions manquantes ;
* analytics absents ;
* observabilité absente ;
* métriques manquantes.

---

## 14. Contract Validation

Valider :

* OpenAPI ;
* events ;
* permissions ;
* analytics contracts ;
* observability contracts.

Le moteur doit détecter :

* divergence ;
* routes manquantes ;
* payloads incohérents ;
* breaking changes ;
* événements non émis.

---

## 15. Observability Validation

Le moteur doit vérifier :

* existence métriques ;
* traces ;
* logs ;
* dashboards ;
* alertes critiques ;
* coverage flows.

Exemple :

```text
Observability Validation
────────────────────────
✓ onboarding.start traced
✓ onboarding.complete traced
⚠ invite_member failure metric missing
⚠ no dashboard contract for onboarding funnel
```

---

## 16. Security Validation

Créer checks :

* auth required ;
* permission checks ;
* dangerous actions ;
* destructive flows ;
* token expiration ;
* secret exposure ;
* rate limiting ;
* replay protection.

---

## 17. Cost Validation

Le moteur doit détecter :

* explosion appels ;
* coûts infra ;
* queues excessives ;
* storage amplification ;
* token overuse ;
* observability cost explosion.

---

## 18. Runtime Verification Events

Ajouter événements :

```text
verification.started
verification.completed
trust.low_confidence
security.issue_detected
flow.integrity_failed
contract.breaking_change_detected
```

---

## 19. Verification Gates

Créer système de gates.

Exemple :

```yaml
verification:
  gates:
    production:
      min_confidence:
        architecture: 0.8
        implementation: 0.75
        security: 0.85
      required_checks:
        - contracts
        - flows
        - observability
        - security
```

---

## 20. Review Orchestration

Le moteur doit pouvoir déclencher automatiquement :

* review architecture ;
* review sécurité ;
* review observabilité ;
* review performance ;
* review produit.

Selon :

* criticité flow ;
* blast radius ;
* confiance faible ;
* changements sensibles.

---

## 21. Replayable Verification

Chaque vérification doit être rejouable.

Créer :

```text
.asagiri/trust/<id>/replay.yaml
```

Exemple :

```yaml
checks:
  - contracts
  - flows
  - observability
repo_commit: abc123
flow: onboarding
branch: onboarding-enterprise
commands:
  - composer test
  - composer phpstan
```

---

## 22. Intégration avec `asa work`

Ajouter :

```bash
asa work "fix onboarding" --strict-trust
```

Comportement :

1. implementation ;
2. verification pipeline ;
3. confidence scoring ;
4. review gates ;
5. validation finale.

---

## 23. Intégration CI

Ajouter mode CI :

```bash
asa verify trust --ci --json
```

Sortie machine-readable.

---

## 24. Architecture Go

Créer :

```text
internal/trust/
  engine.go
  report.go
  findings.go
  gates.go
  confidence/
  checks/
  replay/
  scoring/
```

Interfaces :

```go
type TrustEngine interface {
    Verify(ctx context.Context, req VerificationRequest) (VerificationResult, error)
}

type VerificationCheckRunner interface {
    Run(ctx context.Context, scope VerificationScope) (VerificationCheck, error)
}

type ConfidenceAggregator interface {
    Aggregate(ctx context.Context, checks []VerificationCheck) (ConfidenceReport, error)
}
```

---

## 25. UX terminal cible

```text
Asagiri Trust Engine
════════════════════
Flow: onboarding
Branch: onboarding-enterprise

Checks
──────
✓ Static analysis
✓ Contracts
✓ Flow integrity
⚠ Observability
⚠ Security
✓ Backward compatibility

Confidence
──────────
Architecture:    0.82
Implementation:  0.67
Security:        0.73
Observability:   0.58
Regression:      0.64

Warnings
────────
- invite_member has no retry validation
- onboarding funnel metrics incomplete
- rate limiting not validated

Residual risk: medium

Gate status: BLOCKED
Reason: security confidence below required threshold
```

---

## 26. Critères d’acceptation

Cette évolution est acceptable si :

* `asa verify trust` fonctionne ;
* plusieurs checks spécialisés existent ;
* un trust report est généré ;
* les scores sont explicables ;
* les findings référencent des évidences ;
* les gates peuvent bloquer un workflow ;
* les vérifications sont rejouables ;
* les flows/contracts peuvent être validés ;
* les risques sont visibles ;
* les événements runtime sont émis ;
* les résultats sont exportables JSON ;
* le runtime reste local-first.

---

## 27. Découpage d’implémentation recommandé

### Phase 1 — Trust model

* modèles ;
* report ;
* findings ;
* confidence report.

### Phase 2 — Basic checks

* static analysis ;
* contracts ;
* flows ;
* tests.

### Phase 3 — Confidence engine

* aggregation ;
* scoring ;
* weighting.

### Phase 4 — Blast radius

* dependency graph ;
* impacted flows ;
* critical modules.

### Phase 5 — Gates

* blocking rules ;
* CI mode ;
* policies.

### Phase 6 — Runtime integration

* runtime events ;
* memory ;
* reports ;
* replay.

### Phase 7 — Specialized checks

* observability ;
* analytics ;
* security ;
* performance ;
* cost.

---

## 28. Risques

### Faux sentiment de sécurité

Mitigation :

* scores explicables ;
* limites visibles ;
* residual risk ;
* confiance != vérité.

### Trop de bruit

Mitigation :

* scoring ;
* grouping ;
* severity ;
* prioritization.

### Vérifications trop lentes

Mitigation :

* parallel execution ;
* cache ;
* quick/full modes.

### Over-engineering

Mitigation :

* checks modulaires ;
* runtime optional ;
* policies configurables.

---

## 29. Résumé

Cette évolution ajoute à Asagiri une couche de confiance déterministe et explicable.

Avant :

```text
implementation
  ↓
maybe tested
  ↓
merge
```

Après :

```text
implementation
  ↓
verification pipeline
  ↓
confidence scoring
  ↓
trust gates
  ↓
review
  ↓
validated delivery
```

Principe clé :

> Un agent ne doit jamais être considéré fiable par défaut. La confiance doit être construite à partir d’évidences vérifiables, de validations spécialisées et de risques explicitement visibles.
