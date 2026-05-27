# Spec — Phase finale spec-my-A (embeddings sémantiques + distribution SDK)

**Date :** 2026-05-27  
**Statut :** À implémenter  
**Prérequis :** [`spec-my-A.md`](spec-my-A.md) livré (handoff intégral)  
**Objectif :** fermer les **deux derniers critères de qualité / distribution** encore ouverts après la spec A, sans rouvrir le périmètre fonctionnel déjà validé.

---

## 1. Contexte

La tranche `spec-my-A` couvre les sections 1–26 (couche produit, business intent, runtime, analysis, investigation). Le handoff la considère **complète** au niveau fonctionnel.

Deux points restent toutefois en **niveau V1** :

| # | Sujet | État actuel | Cible phase finale |
|---|--------|-------------|-------------------|
| 1 | **Embeddings mémoire** (§24.10) | Vecteurs déterministes bag-of-words (`internal/embedutil`) | Embeddings **sémantiques** exploitables (synonymes, paraphrases) |
| 2 | **SDK TypeScript** (§24.18) | Package local `sdk/typescript/`, tests OK | Package **`@laprogrammerie/asagiri`** publié sur npm + CI release |

Cette spec ne rouvre **pas** les blocs A–D de spec-my-A. Elle les **durcit** sur ces deux axes uniquement.

---

## 2. Vision

```text
spec-my-A (livré)
  ↓
phase finale
  ├─ mémoire sémantique locale-first (Ollama par défaut)
  └─ SDK TS distribué (npm) + doc consommateur
  ↓
spec-my-A considérée « finale » sans réserve documentée
```

---

## 3. Objectifs

### 3.1 Embeddings sémantiques

1. Remplacer ou compléter le hash bag-of-words par un **embedder pluggable** ;
2. Conserver le **local-first** : embedder par défaut = **Ollama** sur la machine de l’utilisateur ;
3. Permettre un fallback **`hash`** (comportement actuel) sans réseau ;
4. Option **cloud** uniquement sur opt-in explicite (config + flag), jamais par défaut ;
5. Ré-indexer la mémoire existante (`asa memory reindex`) ;
6. Améliorer `asa memory list --query` et la retrieval API (`GET /v1/memory?query=`).

### 3.2 Distribution SDK TypeScript

1. Publier `@laprogrammerie/asagiri` sur le registry npm public (ou registry org) ;
2. Automatiser build + test + publish sur tag `sdk-v*` (GitHub Actions) ;
3. Documenter installation consommateur (HTTP + socket Unix si supporté) ;
4. Versionner le SDK indépendamment du binaire `asa` (semver npm).

---

## 4. Non-objectifs

- Ne pas refondre `internal/analysis/`, runtime, investigation ou product layer ;
- Ne pas imposer un modèle cloud par défaut pour les embeddings ;
- Ne pas embarquer un modèle ONNX lourd dans le binaire `asa` (hors scope sauf décision ultérieure ADR) ;
- Ne pas garantir la reproductibilité bit-à-bit des vecteurs entre versions de modèle Ollama ;
- Ne pas publier le binaire Go sur npm (uniquement le client TS).

---

## 5. Périmètre technique — Embeddings

### 5.1 Interface embedder

Créer :

```text
application/internal/memory/embedder/
  embedder.go      # interface Embedder
  hash.go          # impl actuelle (embedutil)
  ollama.go        # appels API Ollama /api/embeddings
  cloud.go         # optionnel : OpenAI-compatible si config
```

```go
type Embedder interface {
    Embed(ctx context.Context, text string) ([]float32, error)
    Dimensions() int
    Name() string
}
```

Sélection via config :

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

### 5.2 Stockage

- Conserver `embedding_json` sur `memory_entries` ;
- Ajouter métadonnées : `embedder_name`, `embedder_version` (colonne ou champ JSON) ;
- `UpsertMemory` : appeler l’embedder actif, pas uniquement `embedutil.Vector`.

### 5.3 Commandes CLI

| Commande | Rôle |
|----------|------|
| `asa memory reindex` | Re-calculer tous les embeddings avec l’embedder courant |
| `asa memory list --query "..."` | Retrieval par similarité cosinus (déjà présent, à valider sémantique) |
| `asa memory doctor` | Vérifier Ollama joignable, dimensions cohérentes, entrées sans embedding |

### 5.4 API runtime

- `GET /v1/memory?query=...&limit=...` : retrieval sémantique ;
- `POST /v1/memory/reindex` : déclenchement reindex (admin, token requis).

### 5.5 Critères d’acceptation — embeddings

- [ ] `embedder: hash` conserve le comportement actuel (non-régression tests) ;
- [ ] `embedder: ollama` avec `nomic-embed-text` (ou modèle documenté) produit des similarités **> 0.7** sur paires synonymes en test golden ;
- [ ] `asa memory reindex` traite toutes les entrées sans crash ;
- [ ] `cloud` refusé si `enabled: false` même si clé API présente ;
- [ ] Doc EN + FR : section « Memory embeddings » dans runtime / configuration ;
- [ ] ADR-020 : choix Ollama par défaut, fallback hash, opt-in cloud.

### 5.6 Tests

- Tests unitaires : `hash`, `ollama` (mock HTTP) ;
- Test golden : paires (`invite member`, `invitation équipe`) → similarité > seuil ;
- Test intégration optionnel : tag `integration` si Ollama présent en CI (skip sinon).

---

## 6. Périmètre technique — SDK npm

### 6.1 Package

Répertoire existant : `sdk/typescript/`

À compléter :

| Fichier | Contenu |
|---------|---------|
| `package.json` | `repository`, `license`, `keywords`, `files`, `prepublishOnly` |
| `CHANGELOG.md` | Keep a Changelog |
| `README.md` | Install, HTTP, token, exemple session |
| `.npmignore` / `files` | Publier uniquement `dist/` |

### 6.2 Client TS — socket Unix (optionnel mais recommandé)

Si faisable sans dépendance lourde :

- Helper `connectUnix(socketPath: string)` via `node:http` + `fetch` avec dispatcher undici, **ou** doc « HTTP only » + issue reportée si Unix non supporté en Node sans lib.

**Critère minimal :** HTTP documenté et testé ; Unix = bonus si implémenté.

### 6.3 CI release

Workflow `.github/workflows/sdk-npm-publish.yml` :

```yaml
on:
  push:
    tags:
      - 'sdk-v*'
jobs:
  publish:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'
          registry-url: 'https://registry.npmjs.org'
      - run: cd sdk/typescript && npm ci && npm test && npm run build
      - run: cd sdk/typescript && npm publish --provenance --access public
        env:
          NODE_AUTH_TOKEN: ${{ secrets.NPM_TOKEN }}
```

Secrets requis : `NPM_TOKEN` (compte avec droit publish sur `@laprogrammerie`).

### 6.4 Critères d’acceptation — npm

- [ ] `npm install @laprogrammerie/asagiri` fonctionne depuis un projet vierge ;
- [ ] README montre connexion à `asa runtime serve --port 8765` ;
- [ ] Tag `sdk-v0.1.0` déclenche la CI et publie sans intervention manuelle locale ;
- [ ] Version npm alignée sur `CHANGELOG.md` ;
- [ ] Doc site : `reference/typescript-sdk` mise à jour (install npm, pas clone repo).

---

## 7. Documentation

### 7.1 docs-site

| Page | Action |
|------|--------|
| `en/configuration/config-file.mdx` | Section `runtime.memory.embedder` |
| `fr/configuration/config-file.mdx` | Idem FR |
| `en/reference/typescript-sdk.mdx` | `npm install`, exemple complet |
| `fr/reference/typescript-sdk.mdx` | Idem FR |
| `en/concepts/runtime.mdx` | Lien embeddings + npm SDK |

### 7.2 Canon projet

- `docs/ai/05-decisions.md` : **ADR-020** (embeddings), **ADR-021** (npm SDK) ;
- `docs/ai/active/current-spec.md` : pointer `spec-phase-finale.md` comme phase active ;
- `docs/ai/active/handoff.md` : contrat d’exécution dérivé de cette spec.

### 7.3 Exemple config

Mettre à jour `.asagiri/config.yaml.example` :

```yaml
runtime:
  mode: guided
  api:
    port: 8765
    socket: .asagiri/runtime/runtime.sock
  memory:
    embedder: ollama
    ollama:
      base_url: http://127.0.0.1:11434
      model: nomic-embed-text
```

---

## 8. Découpage d’implémentation

### Phase 1 — Embedder abstractions (1–2 j)

- Interface + `hash` (migration depuis `embedutil`) ;
- Config + validation ;
- Tests non-régression.

### Phase 2 — Ollama embedder (1–2 j)

- Client HTTP, timeouts, erreurs explicites ;
- Intégration `UpsertMemory` + `RetrieveByQuery` ;
- `asa memory reindex`, `asa memory doctor`.

### Phase 3 — Golden sémantique + doc (0,5 j)

- Tests golden ;
- Pages docs EN/FR ;
- ADR-020.

### Phase 4 — npm publish (0,5–1 j)

- Finition `package.json`, CHANGELOG ;
- Workflow CI + secret NPM ;
- Tag `sdk-v0.1.0` ;
- ADR-021.

**Estimation totale :** 3–5 jours développeur.

---

## 9. Risques et mitigations

| Risque | Mitigation |
|--------|------------|
| Ollama absent sur la machine | Fallback `hash` + `asa memory doctor` message clair |
| Dimensions modèle ≠ entrées stockées | `reindex` obligatoire après changement de modèle |
| Coût / fuite cloud | `cloud.enabled: false` par défaut, `token_env` uniquement |
| Publish npm cassé | `npm publish --dry-run` en CI sur PR touchant `sdk/typescript/` |
| Scope creep | Interdiction de modifier hors §5 et §6 |

---

## 10. Definition of Done — phase finale

La phase est **terminée** lorsque :

1. Toutes les cases §5.5 et §6.4 sont cochées ;
2. `go test ./...` vert ;
3. `cd sdk/typescript && npm test` vert ;
4. Au moins une version `@laprogrammerie/asagiri` visible sur npm (ou registry documenté) ;
5. `handoff.md` ne mentionne plus d’« écarts assumés » embeddings / npm ;
6. `spec-my-A.md` peut être annoté en tête : *« Complété sans réserve par spec-phase-finale.md »*.

---

## 11. Commandes de validation

```bash
# Embeddings
ollama pull nomic-embed-text
asa memory doctor
asa memory reindex
asa memory list --query "invitation équipe échoue"

# SDK
cd sdk/typescript && npm test && npm run build
npm view @laprogrammerie/asagiri version   # après publish

# Non-régression globale
cd application && go test ./...
```

---

## 12. Références

- [`spec-my-A.md`](spec-my-A.md) — spec parente (§24.10, §24.18)
- [`docs/ai/active/handoff.md`](docs/ai/active/handoff.md) — contrat d’exécution actuel
- ADR-019 — Runtime API + analysis (socket, métriques)
- [`sdk/typescript/PUBLISHING.md`](sdk/typescript/PUBLISHING.md) — brouillon publish manuel
