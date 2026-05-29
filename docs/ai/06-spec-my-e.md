# Spec-my-E — Engineering Knowledge Graph (canon `docs/ai`)

**Statut :** livrée (`2026-05-29`)  
**Spec racine :** [`spec-my-E.md`](../archives/specs/spec-my-E.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)  
**Prérequis :** [`06-spec-my-d.md`](06-spec-my-d.md) (multi-agent coordination)

---

## 1. Résumé

Spec-my-E ajoute un **graphe de connaissance structurelle** local-first reliant produit, contrats, code, tests et opérations :

```text
.asagiri/products + flows + contracts + code + tests
  → knowledge graph (SQLite + graph.json)
  → query / impact / explain / snapshot
  → investigation / trust / execution graph / coordination (--from-graph)
```

---

## 2. Arborescence `.asagiri/knowledge/`

```text
.asagiri/knowledge/
  graph.sqlite
  graph.json
  snapshots/<name>/
    metadata.json
    graph.json
```

---

## 3. Packages Go

| Package | Rôle |
|---------|------|
| `internal/knowledge/` | Modèles, store SQLite, builder, query, impact, staleness, snapshot, display |
| `internal/knowledge/extractors/` | flows, contracts, code, tests, adr, infra, runtime, … |
| `internal/knowledge/renderers/` | JSON, DOT, Mermaid |
| `internal/knowledge/sqlite/` | Persistance `GraphStore` |
| Bridges | `investigation`, `trust`, `executiongraph`, `coordination` (`knowledge_bridge.go`) |

**CLI :** `asa knowledge build|query|explain|snapshot`, `asa impact analyze`, `asa context build --from-graph`.

---

## 4. Build incrémental et staleness (§21)

- Métadonnées `index_metadata.build` : `built_at`, `source_mtimes` par catégorie.
- `--incremental` : skip extractors dont les mtimes sources sont inchangés ; fusion avec graphe existant.
- `GraphStalenessDetector` : compare `source_mtimes` ou fichiers post-`built_at` ; gabarit terminal §21.

---

## 5. Snapshots (§6.5)

`asa knowledge snapshot --name <n>` → `.asagiri/knowledge/snapshots/<n>/` (metadata + copie `graph.json`).

---

## 6. UX terminal (§23)

`FormatKnowledgeBuild` / `WriteKnowledgeBuild` : nodes, edges, sources, confidence moyenne, stale count, warnings.

Mode `--json` sur `knowledge build|query|explain|snapshot` et `impact analyze`.

---

## 7. Tests

```bash
cd application && go test ./internal/knowledge/... -count=1
```

Golden : `testdata/knowledge-graph/{onboarding-flow,api-events,missing-tests,stale-graph}/`.

---

## 8. Documentation publique (site)

| Sujet | Chemins |
|-------|---------|
| Concept | `concepts/engineering-knowledge-graph` |
| Config | section `knowledge` dans `configuration/config-file` |
| CLI | `cli/knowledge`, `cli/impact` (manuel) ; regen : `asa docs generate-cli` |

---

## 9. Décisions

- **ADR-024** — Engineering Knowledge Graph (`internal/knowledge/`, `.asagiri/knowledge/`)

---

## 10. Validation

```bash
cd application && go test ./internal/knowledge/... -count=1
make build && ./bin/asa docs generate-cli
./bin/asa knowledge build --include-flows --include-contracts
./bin/asa knowledge snapshot --name smoke
./bin/asa impact analyze --flow onboarding --action invite_member --json
```

Traçabilité : matrice E-* dans [`active/handoff.md`](active/handoff.md).
