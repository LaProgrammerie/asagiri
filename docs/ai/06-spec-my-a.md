# Spec-my-A — couche produit exécutable (canon `docs/ai`)

**Statut :** livré — **FULL** avec PF-A P1 (`2026-05-29`)  
**Spec racine :** [`spec-my-A.md`](../archives/specs/spec-my-A.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)  
**Reliquats P1 fermés :** [`spec-phase-finale.md`](../archives/specs/spec-phase-finale.md) PF-A-01 (ADR-025), PF-A-02 (ADR-026)

---

## 1. Résumé

Spec-my-A étend Asagiri de :

```text
Intent → spec → tasks → agents
```

vers :

```text
Intent → prototype → flows → contracts → specs → tasks → agents
         ↘ runtime (sessions, memory, API) ↘ investigation → code
```

Quatre blocs livrés : **Product Layer**, **Business Intent**, **Runtime**, **Investigation** + **Analysis layer**.

---

## 2. Arborescence `.asagiri/` (produit + runtime)

```text
.asagiri/
  config.yaml
  products/<product>/
    intent.md
    product.yaml
    business.yaml
    prototype/
    flows/
    screens/
    contracts/
    extraction/
    generated-specs/
    reviews/
  specs/<product>/
  tasks/<product>/
  runtime/
    runtime.db
    hooks.yaml
    api.token
    runtime.sock
  analysis/<product>/
    graphs.json
  investigations/<id>/
  skills/
```

---

## 3. Packages Go

| Package | Rôle |
|---------|------|
| `internal/product/` | Prototype, flows, contracts, specs, review |
| `internal/product/derivation/` | Architecture, observability, permissions |
| `internal/runtime/` | Daemon, sessions, events, memory store, metrics |
| `internal/runtime/api/` | REST HTTP + Unix socket |
| `internal/memory/` | Retrieval, scoring, embeddings (hash) |
| `internal/skills/` | Chargement `.asagiri/skills/` |
| `internal/analysis/` | 7 graphes structuraux |
| `internal/investigation/` | Pipeline §25, impact, graph |
| `internal/embedutil/` | Vecteurs déterministes partagés |
| `pkg/asagiri/` | SDK Go in-process + HTTP |

---

## 4. Commandes CLI (index)

| Domaine | Commandes |
|---------|-----------|
| Produit | `prototype create\|run\|patch`, `flows extract\|inspect\|review`, `contracts extract`, `product review`, `architecture derive`, `spec generate-from-product` |
| Runtime | `daemon start\|run\|status\|stop`, `session *`, `runtime events`, `runtime serve`, `memory list\|consolidate`, `skills list` |
| Analysis | `analysis build` |
| Investigation | `investigate`, `investigate graph`, `investigate impact` |
| Work | `--investigate-first`, `--investigate-on-failure`, `--investigation-depth` |
| Verify | `--investigate-on-failure` |

---

## 5. Documentation publique (site)

Pages dédiées par locale sous `docs-site/content/docs/{en,fr,de,es}/` :

| Sujet | Chemins |
|-------|---------|
| CLI | `cli/runtime`, `cli/runtime-serve`, `cli/investigate`, `cli/analysis` |
| Concepts | `concepts/runtime`, `concepts/investigation`, `concepts/analysis-layer`, `concepts/executable-product-layer` |
| Référence | `reference/typescript-sdk` |
| Config | section `runtime` dans `configuration/config-file` |

Référence CLI générée : `asa docs generate-cli` → `en/cli/generated/`.

---

## 6. Décisions

- **ADR-018** — spec-my-A (product, runtime, investigation V1)
- **ADR-019** — API runtime, analysis, métriques, UX rich, socket Unix

---

## 7. Validation

```bash
cd application && go test ./...
cd sdk/typescript && npm test
asa prototype create "…" --product workspace-saas
asa flows extract workspace-saas && asa spec generate-from-product workspace-saas
asa analysis build --product workspace-saas
asa daemon start --detach
asa investigate "…" --no-cloud
```
