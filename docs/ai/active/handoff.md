# Handoff — execution

> **Contrat d'exécution** Cursor / Copilot / humain.
> **Tranche active :** **audit-coherence-consolidation** — correction &
> simplification des constats d'audit `AUD-001 … AUD-007`. **Livré**
> (`2026-06-05`, ADR-030) — Quality_Gate complet vert (`make build` ∧ `go vet` ∧
> `go test` ∧ `golangci-lint run`, tous exit 0) ; Regeneration_Without_Diff sans
> divergence ; `problems.md` registre clôturé.
> **Précédent :** cockpit-consolidation — Operations Cockpit Direction 4 — livré
> (`2026-05-31`, ADR-029).

## Objectif

Appliquer le **plus petit changement correct** par constat (pas de refonte), sur
des **fichiers existants** uniquement. Cause unique de la dérive : `asa runs`
(ADR-029) enregistrée sans régénération de la doc CLI. Poser les garde-fous qui
empêchent la régression (tests + étape CI), sans moteur d'audit runtime.

## Scope (livré)

- `application/internal/cli/docgen/` — tests régénération (bijection,
  déterminisme, Regeneration_Check hors `meta.json`, no-diff).
- `application/internal/routing/router.go` — `Route(...) (Decision, error)` +
  `ErrNoDeclaredBackend`, config-driven, précédence `no_cloud`, raison exposée.
- `application/internal/cost/estimator.go` — adapté à la nouvelle signature.
- `application/internal/cli/root_ui.go` — `asa explain routing` (DTO
  `RoutingExplanation`, parité plain/json).
- `application/internal/policy/ollama.go` — source canonique unique + check de
  cohérence.
- `application/internal/cli/help.go` — bloc « Pour commencer » (Guided_Path),
  sans retrait de commande.
- `application/internal/onboarding/doctor.go` — clamp score `[0,100]`,
  innocuité `--check-only`, Guided_Remediation.
- `docs-site/content/docs/{en,fr,de,es}/guided-path.mdx` — page d'entrée 4 locales.
- `docs-site/content/docs/en/cli/generated/` — régénérée (`runs.mdx` + liens).
- `.golangci.yml` (v2), `.github/workflows/go-ci.yml` (nouveau),
  `docs/ai/03-standards.md`, `problems.md` (Remediation_Register).

## Hors scope (interdit sans MAJ spec)

- Tout nouveau package `internal/audit` ; toute commande `asa audit`.
- Logique métier dans `internal/ui` (ADR-027) ; refonte de docgen/onboarding.
- Retrait d'une Unitary_Command.

## Definition of Done

- [x] Regeneration_Without_Diff vrai (garde-fou docgen vert)
- [x] Quality_Gate vert (`build`/`test`/`vet`/`lint` exit 0)
- [x] Routing config-driven, explicable, sans `panic` (erreur guidée)
- [x] Policy Ollama reliée au canon courant + check de cohérence
- [x] Guided_Path mis en avant (help + docs 4 locales), Unitary_Command préservées
- [x] Garde-fous onboarding (monotonie readiness, `--check-only`, resume round-trip)
- [x] `problems.md` registre : `AUD-001…007` clôturés ; zéro `blocking` ouvert
- [x] Go CI ajouté (Quality_Gate + Regeneration_Check + scan secrets)
- [x] ADR-030 enregistrée ; `current-spec.md` / `handoff.md` synchro

## Garde-fous

- Pas de `panic` aux frontières CLI : erreurs retournées comme valeurs (`03-standards.md`).
- Déterminisme local-first (ADR-002, ADR-022) : sorties identiques pour entrées identiques.
- UI = client du bus (ADR-027). Routing/policy restent hors `internal/ui`.
- Parité Plain_Output / JSON_Output, jamais conditionnée au mode de rendu.

## Quality_Gate (commande reproductible)

```bash
make build && go vet ./... && go test ./...
# golangci-lint v2 pinné (binaire officiel bâti go >= 1.25) :
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh \
  | sh -s -- -b "$(go env GOPATH)/bin" v2.12.2
golangci-lint run
# Regeneration_Without_Diff :
go run ./application/cmd/asa docs generate-cli --output /tmp/cli-regen
diff -ruq /tmp/cli-regen docs-site/content/docs/en/cli/generated --exclude=meta.json
```

## Références

- `.kiro/specs/audit-coherence-consolidation/` (requirements, design, tasks, audit-report)
- `docs/ai/05-decisions.md` — ADR-030 (et ADR-027/029 pour le contexte UI)
- `docs/ai/03-standards.md` — Quality_Gate, install golangci-lint pinné
- `problems.md` — Remediation_Register (`AUD-001 … AUD-007`)
