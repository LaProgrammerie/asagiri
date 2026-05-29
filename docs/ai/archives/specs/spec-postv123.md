# Mission — Consolidation finale, fiabilisation et préparation open source d’AgentFlow

Objectif :
Réaliser une phase complète de consolidation technique et produit avant ouverture publique du repository.

Le but n’est PAS d’ajouter rapidement des features.
Le but est de rendre le système :

- cohérent ;
- fiable ;
- mesurable ;
- maintenable ;
- performant ;
- reproductible ;
- crédible techniquement ;
- prêt pour une adoption open source.

Le projet doit rester aligné avec sa vision centrale :

> Orchestrer des workflows de développement agentique de manière déterministe, observable, locale-first et cost-aware, en optimisant les coûts, les tokens, le contexte et la fiabilité.

Les modèles premium doivent être réservés au raisonnement à forte valeur ajoutée.
Les tâches lourdes de scanning, extraction, analyse et préparation doivent être réalisées localement autant que possible.

Le système doit pouvoir exploiter intelligemment :
- Kiro CLI + subagents (Sonnet / Opus)
- Cursor Agent en mode Auto
- modèles locaux Ollama
- modèles cloud fast
- modèles cloud heavy

sans dérive architecturelle ni perte de contrôle.

---

## Objectifs de consolidation

### 1. Audit architecture & drift

Vérifier :

- cohérence des responsabilités ;
- absence de duplication ;
- respect des interfaces ;
- absence de logique métier dispersée ;
- cohérence du state management ;
- cohérence des primitives et de la façade intentionnelle ;
- alignement avec les specs ;
- alignement avec la philosophie local-first / cost-aware.

Identifier :
- hacks ;
- dette technique ;
- couplage excessif ;
- sur-ingénierie ;
- features inutiles ;
- abstractions prématurées ;
- dépendances inutiles.

Produire :
- liste des écarts ;
- risques ;
- recommandations ;
- décisions d’architecture proposées.

---

### 2. Stabilisation API & primitives

Valider :
- cohérence des commandes CLI ;
- cohérence des structures ;
- stabilité des interfaces internes ;
- séparation claire entre :
  - moteur
  - providers
  - agents
  - TUI
  - estimation
  - investigation
  - orchestration haut niveau.

Vérifier :
- naming ;
- conventions ;
- ergonomie développeur ;
- extensibilité raisonnable.

---

### 3. Fiabilité & sécurité

Vérifier :
- sandboxing ;
- scope fichiers ;
- gestion secrets ;
- isolation worktrees ;
- protections MCP ;
- timeout ;
- cancellation ;
- retries ;
- reprise après crash ;
- corruption état SQLite ;
- race conditions ;
- logs sensibles.

Ajouter :
- garde-fous manquants ;
- validations ;
- protections défaut.

---

### 4. Performance & optimisation coût/tokens

Objectif :
réduire au maximum les appels cloud inutiles.

Auditer :
- volume de contexte envoyé ;
- étapes qui pourraient être locales ;
- coût potentiel ;
- duplication de contexte ;
- scans inutiles ;
- prompts trop longs ;
- absence de cache ;
- absence de compression.

Vérifier :
- pipeline investigation → réduction → packing ;
- estimation tokens/coût ;
- cache contexte ;
- stratégie de routing.

Mesurer :
- temps ;
- tokens ;
- coût estimé ;
- réduction de contexte ;
- étapes locales vs cloud.

Produire :
- benchmark actuel ;
- pistes optimisation ;
- gains potentiels.

---

### 5. Robustesse workflows agents

Tester :
- Kiro seul ;
- Cursor seul ;
- Ollama seul ;
- workflows hybrides ;
- fallback local → cloud ;
- reprise après échec ;
- tâches ambiguës ;
- tâches volumineuses ;
- gros repo ;
- prompts contradictoires ;
- contexte incomplet.

Identifier :
- hallucinations ;
- comportements dangereux ;
- drift ;
- mauvaises décisions ;
- prompts fragiles ;
- points de non-déterminisme.

---

### 6. Tests & qualité

Objectif :
avoir une base crédible open source.

Ajouter ou vérifier :
- unit tests ;
- integration tests ;
- golden tests ;
- tests worktree ;
- tests estimation ;
- tests routing ;
- tests MCP ;
- tests failure/recovery ;
- tests config ;
- tests state machine.

Produire :
- couverture utile ;
- tests critiques manquants ;
- roadmap qualité.

---

### 7. UX CLI & TUI

Objectif :
rendre le terminal :
- lisible ;
- élégant ;
- utile ;
- rassurant ;
- professionnel.

Auditer :
- bruit ;
- logs inutiles ;
- manque d’informations ;
- clarté des états ;
- progression ;
- cohérence visuelle.

Améliorer :
- progress bars ;
- live metrics ;
- status ;
- erreurs ;
- résumés ;
- explainability.

Le terminal doit inspirer :
- confiance ;
- contrôle ;
- visibilité ;
- maîtrise.

---

### 8. Open source readiness

Préparer :
- README ;
- architecture docs ;
- roadmap ;
- examples ;
- quickstart ;
- CONTRIBUTING ;
- LICENSE Apache 2.0 ;
- benchmarks ;
- philosophy ;
- diagrams ;
- demo workflows.

Vérifier :
- absence secrets ;
- qualité docs ;
- onboarding ;
- cohérence branding ;
- structure repo ;
- installabilité ;
- reproductibilité.

---

### 9. Explainability & trust

Le système doit expliquer :
- pourquoi ce modèle ;
- pourquoi ce coût ;
- pourquoi ce contexte ;
- pourquoi cette tâche ;
- pourquoi cette estimation ;
- pourquoi cette escalade.

L’utilisateur doit comprendre :
- ce que fait AgentFlow ;
- pourquoi il le fait ;
- combien cela coûte ;
- ce qui est local ;
- ce qui est cloud.

---

### 10. Livrables attendus

Produire :
- audit consolidation ;
- audit performance/coût ;
- audit sécurité ;
- audit drift architecture ;
- benchmark workflow ;
- plan de rationalisation ;
- liste TODO priorisée ;
- recommandations V1/V2/V3 ;
- score readiness open source ;
- score confiance/fiabilité ;
- propositions simplification.

Le projet doit privilégier :
- simplicité robuste ;
- déterminisme ;
- observabilité ;
- coût maîtrisé ;
- reproductibilité ;
- maintenabilité ;
- confiance utilisateur.