# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.  
> **Tranche active :** **project-onboarding** — TUI wizard interactif livré (`2026-05-30`).  
> **Précédent :** onboarding core CLI/readiness — FULL (`2026-05-29`).

## Objectif (lot 4 — wizard TUI)

Remplacer l'écran onboarding read-only par un formulaire navigable : champs préremplis, Prev/Next, Advanced, Apply → config + docs + readiness.

| Livrable | Statut |
|----------|--------|
| `internal/onboarding/form.go` — état, validation, ApplyForm | ✅ |
| `ui/screens/onboarding/model.go` — Model interactif | ✅ |
| Bus Advance/SetField/Apply/Wizard query | ✅ |
| `asa onboard --ui` wizard mode | ✅ |
| Tests unit + intégration Castor | ✅ |

---

## Objectif (clôturé — lots 1–3)

Parcours **`asa onboard`** → détection stack → config merge idempotent → bootstrap docs/Kiro → **`asa ready`** / **`asa doctor --full`**. TUI Mission Control (bannière + écran onboarding) et docs-site 4 locales inclus.

---

## Matrice traçabilité (100 %)

| ID spec | Livrable | Lot | Statut |
|---------|----------|-----|--------|
| FR-1.1–1.4 | Détecteurs Go/Castor/Node + `--stack` | 1 | ✅ |
| FR-2.1–2.4 | Writer merge, backup, dry-run, validate | 1 | ✅ |
| FR-3.1–3.3 | `asa ready` JSON/plain/ci/strict + report.json | 1 | ✅ |
| FR-4.1–4.2 | `asa doctor --full` + macOS `/usr/bin/asa` | 1 | ✅ |
| FR-5.1–5.4 | Docs bootstrap + `--force-docs` | 1 | ✅ |
| FR-6.1–6.2 | Idempotence + `--resume` | 1/3 | ✅ |
| FR-7.1–7.2 | TUI banner, écran, palette, parité CLI | 2 | ✅ |
| §12 CLI | Fixtures Go/Castor, dry-run, ready CI | 1 | ✅ |
| §12 UI | Mission Control readiness | 2 | ✅ |
| §12 Docs | docs-site en/fr/de/es + generate-cli | 3 | ✅ |
| OB-3.4 | `examples/onboarding/` | 3 | ✅ |

---

## Definition of Done (lots 1–3)

- [x] `asa onboard --yes --non-interactive` sur fixture Castor → validation Castor
- [x] `asa onboard --yes` sur fixture Go → validation Go
- [x] `asa onboard --dry-run` n'écrit pas
- [x] `asa ready --json` retourne `ready`, `score`, `checks`, `next_actions`
- [x] `asa doctor --full` inclut agents + gitignore + kiro spec
- [x] Idempotence : second onboard ne duplique pas validation
- [x] Bootstrap crée `.kiro/specs/<feature>/` ; garde-fou handoff
- [x] Mission Control bannière readiness + palette
- [x] `--resume` + backups config
- [x] docs-site en/fr/de/es + `asa docs generate-cli`
- [x] `go test ./... -count=1` vert sous `application/`

---

## Audit clôture (`2026-05-29`)

| Vérification | Résultat |
|--------------|----------|
| Commandes spec (`onboard`, `ready`, `doctor --full`) | OK |
| Package `internal/onboarding/` | OK |
| TUI bus-only (pas de logique métier dans `ui/`) | OK |
| Doc EN/FR/DE/ES | OK (`concepts/project-onboarding`, `cli/onboard`, `cli/ready`) |
| CLI généré EN | OK (`en/cli/generated/onboard`, `ready`) |
| `current-spec` ↔ handoff | alignés |
| Tests | `go test ./... -count=1` vert |

---

## Écarts / notes

- **Palette `doctor --full`** : équivalent CLI affiché (pas d'exécution in-process) — parité ADR-027 respectée.
- **Check macOS `/usr/bin/asa`** : info/warn si conflit détecté (comportement spec).

---

## Prochaine spec

Aucune tranche onboarding restante. Spec suivante à définir par product (hors scope onboarding).

---

## Références

- [`spec-onboarding.md`](../archives/specs/spec-onboarding.md)
- [`06-spec-onboarding.md`](../06-spec-onboarding.md)
- `.kiro/specs/project-onboarding/`
- ADR-028, ADR-027, ADR-005
