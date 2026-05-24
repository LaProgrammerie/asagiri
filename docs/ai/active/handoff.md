# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-rename` : rebranding AgentFlow → Asagiri / `asa`** (`2026-05-20`).

## Immediate objective

**Tranche `spec-rename` + module** : Asagiri / `asa`, module `github.com/LaProgrammerie/asagiri`, URLs `LaProgrammerie/asagiri`. **Action humaine restante** : renommer le dépôt GitHub (`hyper-fast-builder` → `asagiri`) — [`docs/migration/github-rename-asagiri.md`](../../migration/github-rename-asagiri.md). Pas de tag/push/commit automatique par l’agent.

## Allowed scope (spec-rename)

- `application/cmd/asa/`, `application/pkg/asagiri/`
- `application/internal/**` (CLI, version, bootstrap, config, env/compat, workflow, MCP, TUI, docgen, tests)
- `.asagiri/config.yaml`, `.asagiri/config.yaml.example`
- `Makefile`, `.goreleaser.yaml`, `.gitignore`
- `.github/workflows/` (CI, release, docs Cloudflare)
- `docs/homebrew/asa.rb.example` (et suppression `agentflow.rb.example` si présent)
- `docs-site/` (branding, installation MDX toutes langues, `basePath` legacy)
- `README.md`, `docs/release-process.md`, `scripts/benchmark-workflow.sh`
- `examples/` (commandes et env `ASA_*`)
- `docs/ai/active/handoff.md`, `current-spec.md`, `05-decisions.md` (ADR-016)
- `docs/ai/00-overview.md`, `01-product.md`, `02-architecture.md`, `03-standards.md`, `context-map.md`
- `docs/migration/github-rename-asagiri.md`

## Definition of Done — spec-rename

- [x] Produit public **Asagiri** (README, help CLI, TUI, logs utilisateur, docs-site titres)
- [x] Commande et binaire **`asa`** (Cobra root, Makefile `bin/asa`, completions, golden/integration tests)
- [x] `asa version` affiche identité Asagiri (pas AgentFlow)
- [x] Config locale **`.asagiri/`** ; exemple versionné ; init/doctor pointent ce chemin ; fallback `.agentflow/` corrigé
- [x] Variables **`ASA_*`** documentées ; fallback `AGENTFLOW_*` via `internal/env/compat` avec warning stderr
- [x] GoReleaser : archives `asa_{OS}_{ARCH}` ; brews → formule **`asa`** sur `LaProgrammerie/homebrew-tap`
- [x] Workflows release/CI : noms d’artefacts et jobs cohérents `asa`
- [x] Docs-site : exemples `asa` ; i18n en/fr/de/es sans `agentflow` résiduel critique (`AGENTFLOW_*` legacy documenté)
- [x] Cloudflare : `CLOUDFLARE_PAGES_PROJECT` = **`asagiri-docs`** (doc + secrets) ; `basePath` legacy **`/asagiri`** si `GITHUB_PAGES=true`
- [x] Recherche globale : zéro occurrence critique `AgentFlow` / `agentflow` hors specs historiques, ADR, migration, fixtures test
- [x] `go test ./...`, `go vet ./...`, `goreleaser check`, `make release-snapshot` (local)
- [x] `asa doctor` OK sur ce dépôt (`asa init` non rejoué — config `.asagiri/` présente)
- [x] Runbook phase 2 rédigé : [`docs/migration/github-rename-asagiri.md`](../../migration/github-rename-asagiri.md)
- [ ] `docs-site` : `pnpm build` / `lint` non exécutés ici (`pnpm` absent du PATH ; `typecheck` OK via sous-agent)

## Hors scope

- Renommage dépôt GitHub sur github.com (humain, voir runbook migration)
- Refactor produit, nouvelles features, changement comportement workflow
- Mise à jour exhaustive des specs racine historiques (`spec.md`, `specv2.md`, …) sauf si bloquant pour DoD public
- Commit / push / tag / release réelle par l’agent

## References

- [`spec-rename.md`](../../../spec-rename.md)
- ADR-016 dans [`05-decisions.md`](../05-decisions.md)
- Phase 2 repo/module : [`docs/migration/github-rename-asagiri.md`](../../migration/github-rename-asagiri.md)
- Hérité livré : `spec-release` (ADR-015), `spec-deploy-doc`, `spec-doc-v2`
