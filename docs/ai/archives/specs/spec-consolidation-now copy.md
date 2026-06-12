# Mission — Consolidation restante (exécution immédiate)

Objectif:
Finaliser les éléments de consolidation encore ouverts après `spec-postv123`, sans élargir le scope produit.

Contraintes:
- pas de nouvelle feature produit;
- pas de refactor massif non justifié par fiabilité/coût/qualité;
- livrer des artefacts vérifiables (docs + tests + métriques).

---

## Backlog à traiter maintenant

### 1) Workflow/intent reliability (<50% → cible >=80%)

À faire:
- augmenter la couverture de tests sur `workflow` et `intent` (cas nominal + erreurs + reprise);
- ajouter tests de non-régression sur transitions de state machine;
- documenter les limites connues restantes.

Critères d’acceptation:
- score interne workflow/intent >=80% (mesure définie dans le rapport);
- tests critiques workflow/intent verts en local et CI.

---

### 2) TODO priorisée de consolidation

À faire:
- extraire une liste unique de dette/action depuis `docs/consolidation/*`;
- prioriser P0/P1/P2 avec impact, risque, effort;
- affecter chaque item à un lot d’exécution concret.

Critères d’acceptation:
- fichier de backlog priorisé publié;
- chaque item a owner, statut et DoD minimale.

---

### 3) Plan de rationalisation technique

À faire:
- lister les simplifications à fort ROI (suppression duplication, points de couplage, surface API interne);
- proposer une séquence en lots de petite taille;
- poser les garde-fous de non-régression.

Critères d’acceptation:
- plan validé en 3 vagues max (stabilisation, simplification, durcissement);
- risques et rollback documentés par vague.

---

### 4) Robustesse des workflows agents

À faire:
- exécuter une matrice de scénarios mini (local-only, cloud-only, hybride, fallback, reprise après échec);
- capturer comportements non déterministes et points fragiles;
- corriger les écarts P0/P1.

Critères d’acceptation:
- matrice de scénarios exécutée et tracée;
- aucun scénario P0 ne casse le flux de bout en bout.

---

### 5) Qualité et tests critiques manquants

À faire:
- compléter les tests manquants (MCP, recovery, config, routing);
- fiabiliser les tests fragiles/flaky;
- définir une baseline de couverture utile par domaine.

Critères d’acceptation:
- liste de tests critiques manquants fermée ou planifiée;
- suite de tests stable sur 3 runs consécutifs.

---

### 6) Performance/coût: gains mesurés

À faire:
- rejouer benchmark workflow de référence;
- mesurer temps, tokens estimés, taille contexte;
- implémenter quick wins restants (réduction contexte/cache) sans changer le comportement métier.

Critères d’acceptation:
- rapport avant/après publié;
- amélioration mesurée sur au moins un axe (latence, coût estimé, contexte).

---

### 7) Open source readiness (74/100 → cible >=85/100)

À faire:
- fermer les gaps de docs d’onboarding/contribution encore ouverts;
- vérifier structure repo et reproductibilité d’installation;
- durcir checklist de publication.

Critères d’acceptation:
- score readiness >=85/100;
- checklist OSS exécutable sans ambiguïté.

---

### 8) Explainability & trust (sorties utilisateur)

À faire:
- standardiser les sorties expliquant modèle/coût/contexte/décision;
- aligner `estimate` et `work` sur le même format de justification;
- réduire le bruit et garder les signaux utiles.

Critères d’acceptation:
- sorties homogènes sur les commandes clés;
- validation manuelle sur scénarios de référence.

---

## Livrables obligatoires

- rapport consolidation final;
- backlog priorisé P0/P1/P2;
- plan de rationalisation;
- rapport benchmark avant/après;
- rapport qualité (tests ajoutés + risques restants);
- score OSS readiness mis à jour;
- score confiance/fiabilité mis à jour.

---

## Definition of Done (tranche)

- [ ] Workflow/intent >=80% avec preuves de tests.
- [ ] Backlog priorisé consolidé et actionnable.
- [ ] Plan de rationalisation approuvé et séquencé.
- [ ] Matrice de robustesse agents exécutée sans blocant P0.
- [ ] Tests critiques manquants couverts ou planifiés avec échéance.
- [ ] Gains performance/coût mesurés et documentés.
- [ ] OSS readiness >=85/100.
- [ ] Explainability homogène sur les commandes clés.
