# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-release` : distribution release & Homebrew** (`2026-05-17`).

## Immediate objective

**Tranche `spec-release`** : chaîne GoReleaser + GitHub Actions + Homebrew tap + docs d’installation. Pas de tag/push/commit automatique par l’agent.

## Allowed scope (spec-release)

- `application/internal/version/`
- `application/internal/cli/root.go`, `root_test.go`
- `.goreleaser.yaml`
- `.github/workflows/release.yml`, `.github/workflows/release-check.yml`
- `Makefile` (RELEASE_LDFLAGS, `release-snapshot`, `release-check`)
- `docs/release-process.md`, `docs/homebrew/`
- `README.md`
- `docs-site/content/docs/{en,fr,de,es}/getting-started/installation.mdx`
- `docs/ai/active/handoff.md`, `current-spec.md`, `05-decisions.md` (ADR-015)
- `.gitignore` (`dist/`)

## Definition of Done — spec-release

- [x] `application/internal/version` : `Version`, `Commit`, `Date` (defaults `dev` / `unknown` / `unknown`)
- [x] `agentflow version` : `AgentFlow vX`, `commit:`, `built:`
- [x] Tests CLI + package version
- [x] `.goreleaser.yaml` v2 : builds multi-OS/arch, archives LICENSE+README, Windows zip, checksums, changelog filters, brews → `LaProgrammerie/homebrew-tap`, release `hyper-fast-builder`
- [x] `release.yml` (tags `v*`, tests, goreleaser-action, tokens documentés)
- [x] `release-check.yml` (PR paths, `goreleaser check`)
- [x] Makefile : `RELEASE_LDFLAGS`, `release-snapshot`, `release-check` ; `build` préservé
- [x] `docs/release-process.md`, `docs/homebrew/agentflow.rb.example`
- [x] README + installation MDX (4 langues) synchronisés
- [ ] Première release réelle `v0.1.0` sur GitHub (action humaine)
- [ ] `brew install LaProgrammerie/tap/agentflow` validé post-release
- [x] `go test ./...`, `go vet ./...`, `goreleaser check`, `make release-snapshot` (local)

## Hors scope

- Packages Debian/RPM, Scoop, Winget, Docker image
- cosign / SBOM / notarization macOS
- Renommage module Go
- Commit / push / tag par l’agent

## References

- [`spec-release.md`](../../../spec-release.md)
- ADR-015 dans [`05-decisions.md`](../05-decisions.md)
- Hérité : `spec-postv123`, `spec-doc-v2`, `spec-deploy-doc`
