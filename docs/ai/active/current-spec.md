# Current spec — audit-coherence-consolidation (active)

**Phase :** Correction & simplification (audit `AUD-001 … AUD-007`) — **livré**
(`2026-06-05`, ADR-030 ; Quality_Gate complet vert, Regeneration_Without_Diff
sans divergence).
**Handoff :** [`handoff.md`](handoff.md)

## Objet

Appliquer les **corrections concrètes** des sept constats d'audit `AUD-001 …
AUD-007`, **simplifier** la couche d'orchestration perçue comme opaque, et poser
les **garde-fous** anti-régression — sans nouvelle couche ni moteur d'audit
runtime. Cause unique de la dérive : la commande `runs` (ADR-029) n'a pas été
suivie d'une régénération de la doc CLI.

## Spec

- **Kiro :** `.kiro/specs/audit-coherence-consolidation/` (requirements, design, tasks)
- **Audit source :** `.kiro/specs/audit-coherence-consolidation/audit-report.md`
- **Code cible :** fichiers existants — `cli/docgen`, `internal/routing`,
  `internal/policy`, `internal/cli` (help, explain), `internal/onboarding`,
  `problems.md`, `.golangci.yml`, `.github/workflows/go-ci.yml`, `docs/ai/03-standards.md`

## Constats → corrections (livrés)

| Constat | Sévérité | Correction (plus petit changement) | Statut |
|---------|----------|-------------------------------------|--------|
| AUD-001 | error | Régénérer la doc CLI → ajoute `runs.mdx` | clôturé |
| AUD-002 | error | Même régénération → lien fratrie `> - [Runs](./runs.mdx)` | clôturé |
| AUD-003 | error | `golangci-lint` v2 pinné + `.golangci.yml` v2 + workflow Go CI | clôturé |
| AUD-004 | warn | `problems.md` = Remediation_Register (automate de statut) | clôturé |
| AUD-005 | warn | Routing config-driven : `Route → (Decision, error)`, précédence `no_cloud`, `asa explain routing` | clôturé |
| AUD-006 | info | Rôles Ollama reliés au canon courant + check de cohérence | clôturé |
| AUD-007 | info | Guided_Path mis en avant (help + page docs 4 locales), sans retrait | clôturé |

## Garde-fous (prouvés verts)

- **Quality_Gate** : `make build` ∧ `go vet ./...` ∧ `go test ./...` ∧
  `golangci-lint run` (binaire v2.12.2 pinné), tous exit 0.
- **Regeneration_Without_Diff** : `asa docs generate-cli` → tmp +
  `diff --exclude=meta.json`, aucune divergence.
- **Go CI** : `.github/workflows/go-ci.yml` (push/PR) = Quality_Gate +
  Regeneration_Check + scan secrets en clair.
- **PBT** : propriétés `P1 … P21` (docgen, routing, policy, onboarding, locales).

## Invariants

- Aucune nouvelle couche, aucun package `internal/audit`, aucune commande `asa audit`.
- UI = client du bus (ADR-027) ; aucune logique métier dans `internal/ui`.
- Moteur local-first et déterministe (sorties identiques pour entrées identiques).
- Toutes les Unitary_Command préservées (`asa spec | plan | enrich | dev | verify | review`).

## Précédent (livré)

- **cockpit-consolidation** — Operations Cockpit Direction 4 — livré
  (`2026-05-31`, ADR-029).
- **project-onboarding** — Project Onboarding & Readiness + TUI wizard — livré
  (`2026-05-30`).

Branding : **Asagiri** / **`asa`**.
