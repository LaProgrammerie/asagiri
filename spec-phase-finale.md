# Spec — Phase finale (reliquats transverses)

**Date :** 2026-05-27 (création) — **mise à jour :** 2026-05-29  
**Statut :** À implémenter  
**Prérequis :** [`spec-my-A.md`](spec-my-A.md), [`spec-my-B.md`](spec-my-B.md), [`spec-my-C.md`](spec-my-C.md) livrés (handoff §27 / DoD global cochés)  
**Objectif :** fermer **tous les écarts assumés, limitations V1 et durcissements** laissés ouverts après les specs A, B et C — **sans rouvrir** le périmètre fonctionnel déjà validé de chaque spec parente.

> Ce document est la **source unique** des reliquats « phase finale ». Les specs A/B/C restent la vérité fonctionnelle ; ici on ne fait que **compléter** ce qui était explicitement V1, stub, ou non câblé au moment de la livraison.

---

## 1. Registre des reliquats

| ID | Spec source | Sujet | Sévérité | § détail |
|----|-------------|--------|----------|----------|
| **PF-A-01** | spec-my-A §24.10 | Embeddings mémoire sémantiques (Ollama) | P1 | [§3](#3-spec-my-a--embeddings-sémantiques) |
| **PF-A-02** | spec-my-A §24.18 | SDK TypeScript sur npm + CI | P1 | [§4](#4-spec-my-a--sdk-npm) |
| **PF-C-01** | spec-my-C §5.3 | `--checkpoint-every` exposé CLI, non propagé au runner | P1 | [§5](#5-spec-my-c--execution-graph) |
| **PF-C-02** | spec-my-C §24 | `execution_graph.gates` YAML non injecté dans le planner | P1 | [§5](#5-spec-my-c--execution-graph) |
| **PF-C-03** | spec-my-C §24 | `execution_graph.enabled: false` non respecté (défaut force activé) | P1 | [§5](#5-spec-my-c--execution-graph) |
| **PF-C-04** | spec-my-C §17 | Évaluation trust dans le runner : score stub, pas `trust.Engine` complet | P1 | [§5](#5-spec-my-c--execution-graph) |
| **PF-C-05** | spec-my-C §5.5, §15 | `asa graph resume` sans checkpoint : reprise partielle / erreur peu claire | P1 | [§5](#5-spec-my-c--execution-graph) |
| **PF-C-06** | spec-my-C §10 | Inférence dépendances V2 : events, architecture projection, mémoire historique | P2 | [§5](#5-spec-my-c--execution-graph) |
| **PF-X-01** | legacy / CLI | `asa resume <run-id>` n’exécute pas la chaîne agents (hors `--dry-run`) | P2 | [§6](#6-transverse--cli-et-plateforme) |
| **PF-X-02** | cost | Pas de tokenizers provider-exacts (heuristique `chars_per_token`) | P3 | [§6](#6-transverse--cli-et-plateforme) |
| **PF-X-03** | RAG | `asa index` sans embeddings vectoriels / retrieval sémantique index | P2 | [§6](#6-transverse--cli-et-plateforme) |
| **PF-X-04** | docgen | Exemples CLI générés minimaux (`cobra.Example` manquant) | P3 | [§6](#6-transverse--cli-et-plateforme) |

**Légende sévérité :** P1 = contrat spec ou UX trompeuse ; P2 = valeur produit importante ; P3 = qualité / doc.

Traçabilité historique : [`problems.md`](problems.md) (GAP-001…004) → IDs **PF-X-01…04**.

---

## 2. Vision

```text
spec-my-A, spec-my-B, spec-my-C (livrés — DoD §27 / handoff)
  ↓
phase finale (ce document)
  ├─ PF-A-*  mémoire sémantique + npm SDK
  ├─ PF-C-*  durcissement execution graph + trust runtime
  └─ PF-X-*  reliquats CLI / cost / RAG / docgen
  ↓
Aucun « écart assumé » restant dans handoff / problems.md
  ↓
spec-my-D+ (nouvelles capacités — hors ce document)
```

---

## 3. Spec-my-A — Embeddings sémantiques

**Réf.** [`spec-my-A.md`](spec-my-A.md) §24.10 · **ID** PF-A-01

### État actuel

Vecteurs déterministes bag-of-words (`internal/embedutil`).

### Cible

1. **Embedder pluggable** ; défaut **Ollama** local-first ;
2. Fallback **`hash`** (comportement actuel) sans réseau ;
3. Option **cloud** uniquement sur opt-in (`enabled: false` par défaut) ;
4. `asa memory reindex` ; retrieval sémantique CLI + API.

### Périmètre technique

```text
application/internal/memory/embedder/
  embedder.go      # interface Embedder
  hash.go          # migration embedutil
  ollama.go        # API Ollama /api/embeddings
  cloud.go         # optionnel OpenAI-compatible
```

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    Dimensions() int
    Name() string
}
```

Config :

```yaml
runtime:
  memory:
    embedder: ollama   # hash | ollama | cloud
    ollama:
      base_url: http://127.0.0.1:11434
      model: nomic-embed-text
    cloud:
      enabled: false
      provider: openai
      model: text-embedding-3-small
      token_env: OPENAI_API_KEY
```

| Commande | Rôle |
|----------|------|
| `asa memory reindex` | Re-calculer tous les embeddings |
| `asa memory list --query "..."` | Similarité cosinus (valider sémantique) |
| `asa memory doctor` | Ollama joignable, dimensions, entrées orphelines |

API : `GET /v1/memory?query=…` ; `POST /v1/memory/reindex` (admin).

### Critères d’acceptation — PF-A-01

- [ ] `embedder: hash` — non-régression tests ;
- [ ] `embedder: ollama` — similarité **> 0.7** sur paires synonymes (golden) ;
- [ ] `asa memory reindex` sans crash sur corpus existant ;
- [ ] `cloud` refusé si `enabled: false` même avec clé API ;
- [ ] Doc site EN/FR/DE/ES : `runtime.memory.embedder` ;
- [ ] ADR dédié embeddings (ne pas réutiliser ADR-020 trust).

### Tests

Unit `hash` / `ollama` (mock HTTP) ; golden synonymes ; intégration tag `integration` si Ollama en CI (skip sinon).

---

## 4. Spec-my-A — SDK npm

**Réf.** [`spec-my-A.md`](spec-my-A.md) §24.18 · **ID** PF-A-02

### Cible

1. Publier `@laprogrammerie/asagiri` sur npm (ou registry org) ;
2. CI sur tag `sdk-v*` ;
3. Doc consommateur HTTP (+ Unix si faisable) ;
4. Semver npm indépendant du binaire `asa`.

### Package (`sdk/typescript/`)

| Fichier | Contenu |
|---------|---------|
| `package.json` | `repository`, `license`, `files`, `prepublishOnly` |
| `CHANGELOG.md` | Keep a Changelog |
| `README.md` | Install, HTTP, token, session |

### CI release

Workflow `.github/workflows/sdk-npm-publish.yml` sur `sdk-v*` ; secret `NPM_TOKEN`.

### Critères d’acceptation — PF-A-02

- [ ] `npm install @laprogrammerie/asagiri` depuis projet vierge ;
- [ ] README : connexion `asa runtime serve --port 8765` ;
- [ ] Tag `sdk-v0.1.0` publie sans intervention locale ;
- [ ] Doc site `reference/typescript-sdk` (4 locales) : install npm ;
- [ ] ADR dédié distribution npm.

---

## 5. Spec-my-C — Execution graph

**Réf.** [`spec-my-C.md`](spec-my-C.md) · [`06-spec-my-c.md`](docs/ai/06-spec-my-c.md) · **IDs** PF-C-01…06

Les critères §27 de spec-my-C sont **livrés** ; cette section durcit les **5 écarts P1** identifiés à la clôture (2026-05-29) plus l’inférence V2.

### PF-C-01 — `--checkpoint-every`

| | |
|--|--|
| **Actuel** | Flag documenté sur `asa graph run` (`node` \| `group`) ; runner ignore la cadence |
| **Cible** | Propager au `runner` : créer checkpoint après chaque nœud ou après chaque parallel group selon la valeur |
| **Fichiers** | `internal/cli/graph_cmd.go`, `internal/executiongraph/runner.go`, `checkpoints.go` |
| **Tests** | Intégration : `--checkpoint-every node` → N checkpoints = N nœuds exécutés |

### PF-C-02 — Config `execution_graph.gates`

| | |
|--|--|
| **Actuel** | Bloc YAML dans `config.yaml.example` ; planner/runner n’appliquent pas `human_approval_for`, `trust_required_for_high_risk` |
| **Cible** | Injecter `config.ExecutionGraphConfig.Gates` dans enrichment trust + nœuds `manual_approval` |
| **Fichiers** | `internal/config/`, `planner.go`, `trust_enrichment.go` |
| **Tests** | Fixture high-risk + `trust_required_for_high_risk: true` → nœud trust obligatoire |

### PF-C-03 — `execution_graph.enabled`

| | |
|--|--|
| **Actuel** | `enabled: false` ignoré ; commandes graph toujours actives |
| **Cible** | Si `enabled: false` : `plan graph` / `graph run` retournent erreur structurée (code + message doc) |
| **Fichiers** | `internal/config/`, commandes CLI graph |
| **Tests** | Config disabled → exit ≠ 0, message explicite |

### PF-C-04 — Trust engine dans le runner

| | |
|--|--|
| **Actuel** | Gate trust avec score/heuristique stub dans `runner` |
| **Cible** | Appeler `trust.Engine` (spec-my-B) pour nœuds `trust_verification` / gates post-implémentation ; respecter `--strict-trust` |
| **Fichiers** | `internal/executiongraph/runner.go`, `trust_enrichment.go`, wiring depuis CLI |
| **Tests** | Fixture minimal-product : gate bloquant → nœud `failed` / graphe `blocked` |

### PF-C-05 — `graph resume` robuste

| | |
|--|--|
| **Actuel** | Reprise partielle si aucun checkpoint ; message peu clair si status `planned` |
| **Cible** | Sans checkpoint : erreur explicite **ou** dry-run des nœuds `ready` uniquement (comportement documenté) ; avec checkpoint : reprendre nœuds restants et mettre à jour `events.jsonl` |
| **Fichiers** | `runner.go`, `graph_cmd.go`, doc `graph-resume` |
| **Tests** | `resume` sans checkpoint → erreur ; avec checkpoint fixture → nœuds `succeeded` conservés |

### PF-C-06 — Dependency inference V2 (P2)

Étendre `DependencyInferer` (§10 spec-my-C) au-delà du V1 (tasks, flows, contracts, fichiers) :

| Source | Règle exemple |
|--------|----------------|
| Events | changement event public → backward compat + trust |
| Architecture projection | module dépendant → arête `requires` |
| Mémoire historique | échec récent sur flow → nœud investigation auto |

**Non-objectif :** refonte complète du planner — incrément testé par fixture golden dédiée.

### Critères d’acceptation — bloc PF-C

- [ ] PF-C-01 à PF-C-05 : cases ci-dessus + `go test ./internal/executiongraph/...` vert ;
- [ ] Doc EN/FR/DE/ES : comportement `enabled`, `gates`, `resume`, `checkpoint-every` alignée ;
- [ ] `handoff.md` : section écarts spec-my-C vidée après livraison PF-C.

### Commandes de validation — PF-C

```bash
./bin/asa plan graph workspace-saas --flow workspace-onboarding
./bin/asa graph run workspace-saas --flow workspace-onboarding --checkpoint-every node --dry-run
./bin/asa graph resume graph-<id>   # avec et sans checkpoints/
```

---

## 6. Transverse — CLI et plateforme

**IDs** PF-X-01…04 (ex-[`problems.md`](problems.md) GAP-001…004)

### PF-X-01 — `asa resume` exécution agents

Hors `--dry-run`, `asa resume <run-id>` affiche le prochain step sans enchaîner les agents. **Cible :** mode `--execute` production-safe ou enchaînement documenté comme Experimental jusqu’à implémentation.

### PF-X-02 — Tokenizers cost exacts

Heuristique `chars_per_token` seulement. **Cible :** tokenizers provider (ou doc « estimation non facture » partout). Priorité P3.

### PF-X-03 — RAG vectoriel

`asa index` = chunks SQLite LIKE ; pas de retrieval sémantique index. **Cible :** aligner avec PF-A-01 (embeddings) ou documenter dépendance explicite. Peut fusionner avec PF-A-01 si même embedder.

### PF-X-04 — Docgen exemples Cobra

Pages `cli/generated/*` sans args obligatoires. **Cible :** renseigner `cobra.Command.Example` sur commandes critiques ; regénérer `asa docs generate-cli`.

---

## 7. Non-objectifs (phase finale entière)

- Ne pas implémenter **spec-my-D** (multi-agent coordination) ni specs ultérieures ;
- Ne pas refondre analysis / product layer / investigation hors reliquats listés ;
- Ne pas imposer cloud par défaut (embeddings, agents) ;
- Ne pas publier le binaire Go sur npm ;
- Pas de commit/push automatique par l’agent.

---

## 8. Documentation

### docs-site (toutes locales maintenues : en, fr, de, es)

| Sujet | Pages |
|-------|--------|
| PF-A-01 | `configuration/config-file` (`runtime.memory.embedder`), `concepts/runtime` |
| PF-A-02 | `reference/typescript-sdk` |
| PF-C-* | `concepts/execution-graph`, `cli/graph-run`, `cli/graph-resume`, `configuration/config-file` (`execution_graph`) |

### Canon projet

- `docs/ai/active/current-spec.md` — phase active = **spec-phase-finale** ;
- `docs/ai/active/handoff.md` — contrat dérivé de ce fichier (matrice PF-*) ;
- `docs/ai/05-decisions.md` — ADR par bloc (embeddings, npm, graph hardening) ;
- [`problems.md`](problems.md) — synchroniser statuts avec registre §1.

---

## 9. Découpage d’implémentation

| Phase | IDs | Estimation |
|-------|-----|------------|
| **F1** | PF-A-01 embedder + reindex | 1–2 j |
| **F2** | PF-A-01 golden + doc 4 locales | 0,5 j |
| **F3** | PF-A-02 npm + CI | 0,5–1 j |
| **F4** | PF-C-01…03 config/CLI wiring | 1 j |
| **F5** | PF-C-04 trust runner + PF-C-05 resume | 1–2 j |
| **F6** | PF-C-06 inference V2 (optionnel P2) | 1–2 j |
| **F7** | PF-X-01…04 au fil des priorités | variable |

**Total indicatif :** 5–9 j (sans F6/F7 complets).

---

## 10. Definition of Done — phase finale globale

La phase est **terminée** lorsque :

1. Toutes les cases **PF-A-01**, **PF-A-02**, **PF-C-01…05** sont cochées ;
2. `cd application && go test ./...` vert ;
3. `cd sdk/typescript && npm test` vert (PF-A-02) ;
4. `@laprogrammerie/asagiri` publié sur npm (ou registry documenté) ;
5. Registre §1 : tous les P1 en statut **Fermé** ; P2/P3 soit fermés soit reportés vers spec-my-D avec accord explicite ;
6. `handoff.md` et `problems.md` ne listent plus d’écarts ouverts couverts par ce document ;
7. En-têtes `spec-my-A.md`, `spec-my-C.md` annotables : *« Reliquats fermés par spec-phase-finale.md »*.

---

## 11. Commandes de validation globales

```bash
# PF-A — embeddings
ollama pull nomic-embed-text
asa memory doctor
asa memory reindex
asa memory list --query "invitation équipe échoue"

# PF-A — SDK
cd sdk/typescript && npm test && npm run build
npm view @laprogrammerie/asagiri version

# PF-C — execution graph
asa plan graph workspace-saas --flow workspace-onboarding
asa graph run workspace-saas --flow workspace-onboarding --checkpoint-every node --dry-run
asa graph resume graph-<id>

# Non-régression
cd application && go test ./...
```

---

## 12. Références

| Document | Rôle |
|----------|------|
| [`spec-my-A.md`](spec-my-A.md) | Parent PF-A-* |
| [`spec-my-B.md`](spec-my-B.md) | Trust engine (prérequis PF-C-04) |
| [`spec-my-C.md`](spec-my-C.md) | Parent PF-C-* |
| [`spec-my-D.md`](spec-my-D.md) | **Hors scope** — coordination multi-agent |
| [`docs/ai/active/handoff.md`](docs/ai/active/handoff.md) | Contrat d’exécution |
| [`docs/ai/06-spec-my-c.md`](docs/ai/06-spec-my-c.md) | Canon graph livré |
| [`problems.md`](problems.md) | Tracker GAP ↔ PF-X |
| [`sdk/typescript/PUBLISHING.md`](sdk/typescript/PUBLISHING.md) | Brouillon publish npm |
