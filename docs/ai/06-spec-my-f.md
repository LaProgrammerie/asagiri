# Spec-my-F — Replay & Deterministic Execution (canon `docs/ai`)

**Statut :** livrée (`2026-05-29`)  
**Spec racine :** [`spec-my-F.md`](../archives/specs/spec-my-F.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)  
**Prérequis :** [`06-spec-my-e.md`](06-spec-my-e.md) (knowledge graph) ; stacks A–E livrées

---

## 1. Résumé

Spec-my-F ajoute un **moteur de replay** local-first pour capturer, rejouer, comparer et auditer les workflows d’ingénierie :

```text
run / graph / investigation
  → replay package (.asagiri/replays/<id>/)
  → replay run (full | simulation | offline | audit | compare)
  → compare / explain (divergences)
  → snapshot
```

**Principe :** reproductibilité **pratique** (pas de déterminisme LLM parfait).  
**Distinct de** `asa trust replay` : voir §9 (trust manifest vs replay package).

---

## 2. Arborescence `.asagiri/replays/`

```text
.asagiri/replays/
  replay-2026-05-29-<suffix>/
    replay.yaml
    context/
    prompts/
    outputs/
    graph/
    trust/
    investigations/
    runtime/
    reports/
    snapshots/<name>/    # via replay snapshot
```

ID : `replay-YYYY-MM-DD-<suffix>` (validation stricte, pas de traversal).

---

## 3. Packages Go

| Package | Rôle |
|---------|------|
| `internal/replay/` | `Manager`, capture, run, compare, divergence, provenance, compression, snapshots, policies |
| `internal/config/` | `ReplayConfig` (bloc `replay:`) |
| `internal/cli/` | `replay create|run|compare|explain|snapshot` |

Fichiers principaux : `package.go`, `capture.go`, `replay.go`, `compare.go`, `divergence.go`, `provenance.go`, `compression.go`, `snapshots.go`, `policies.go`, `display.go`.

**Captures intégrées (lots 5) :** graph state, trust reports, investigation packs, handoffs, runtime events, prompts (selon policies).

---

## 4. CLI

| Commande | Rôle |
|----------|------|
| `asa replay create` | Package depuis `--from-run`, `--from-graph`, ou `--from-investigation` ; `--include-runtime`, `--include-prompts`, `--include-events` |
| `asa replay run <id>` | Rejouer ; `--dry-run`, `--compare`, `--strict`, `--offline`, `--simulation` |
| `asa replay compare <a> <b>` | Comparaison structurée (coût, trust, graph, artefacts) |
| `asa replay explain <a> <b>` | Divergences lisibles |
| `asa replay snapshot <id> --name` | Snapshot nommé sous le package |

`--json` sur toutes les sous-commandes.

---

## 5. Modes (§11)

| Mode | Comportement |
|------|----------------|
| `full` | Rejoue le workflow avec agents (hors offline) |
| `simulation` | Graph + events + outputs existants, sans réexécution agents (`--simulation`) |
| `offline` | Pas d’appels cloud / APIs externes (`--offline` ou `offline_mode_default`) |
| `audit` | Validation package + provenance |
| `compare` | Prépare / affiche diff (`--compare` sur run, ou `replay compare`) |

---

## 6. Divergences (§15)

`DetectDivergences` / comparator : outputs, trust scores, validations manquantes, nœuds graph, artefacts absents, stale knowledge (warning).

`--strict` sur `replay run` → erreur si divergence détectée (`ErrStrictDivergence`).

---

## 7. Policies & secrets (§22–23)

Config `replay:` : `capture_prompts`, `capture_runtime_events`, `capture_agent_outputs`, `redact_secrets`, `offline_mode_default`, `compress_threshold_bytes`.

`RedactSecrets` : tokens, bearer, lignes `.env`, clés `sk-*` ; exclusion fichiers `.env` / credentials.

---

## 8. UX terminal (§26)

`WriteReplayCreate` / `WriteReplayRun` / `WriteReplayComparison` / `WriteReplayExplain` — gabarit « Asagiri Replay Engine » (artefacts, mode, comparaison, warnings).

---

## 9. Trust replay vs engineering replay

| | Trust (`spec-my-B`) | Engineering replay (`spec-my-F`) |
|--|---------------------|----------------------------------|
| Package | `internal/trust/replay/` | `internal/replay/` |
| Artefacts | `.asagiri/trust/<id>/replay.yaml` | `.asagiri/replays/<replay-id>/` |
| CLI | `asa trust replay <trust-id>` | `asa replay create|run|compare|…` |
| Scope | Rejouer une **vérification** gates/checks | Capturer et rejouer un **workflow** complet |

Les packages trust peuvent être **copiés** dans un replay package (`trust/`).

---

## 10. Tests

```bash
cd application && go test ./internal/replay/... -count=1
```

Golden : `internal/replay/testdata/replay/{basic-run,graph-run,trust-validation,investigation,divergence}/`.

---

## 11. Documentation publique (site)

| Sujet | Chemins |
|-------|---------|
| Concept | `concepts/replay-engine` (en/fr/de/es) |
| Config | section `replay` dans `configuration/config-file` |
| CLI | `cli/replay` (manuel) ; regen : `asa docs generate-cli` → `en/cli/generated/replay*.mdx` |

---

## 12. DoD global (§28)

- [x] `asa replay create` génère un replay package avec artefacts principaux
- [x] `asa replay run` (offline, simulation, strict)
- [x] Comparaisons et détection de divergences
- [x] Trust reports et graph state capturés / comparables
- [x] Handoffs et investigations intégrés à la capture
- [x] Secrets redacted ; tests unitaires / golden / CLI

Traçabilité : matrice F-* dans [`active/handoff.md`](active/handoff.md).

---

## 13. Validation

```bash
cd application && go test ./internal/replay/... -count=1
make build && ./bin/asa docs generate-cli
./bin/asa replay create --from-graph graph-2026-05-29-test0001 --include-runtime
./bin/asa replay run replay-2026-05-29-<id> --offline --dry-run
./bin/asa replay compare replay-a replay-b
./bin/asa replay snapshot replay-2026-05-29-<id> --name before-review
```
