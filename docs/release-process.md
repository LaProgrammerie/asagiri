# Release process

AgentFlow binaries are published from **hyper-fast-builder** via [GoReleaser](https://goreleaser.com/) when a SemVer tag `v*` is pushed.

## Prerequisites

- `main` is green (`go test ./...`, CI).
- Repository secrets configured (see below).
- [Homebrew tap](https://github.com/LaProgrammerie/homebrew-tap) exists and is writable by the tap token.

## Steps

1. Ensure `main` is green and changelog expectations are met.
2. Create and push a tag:

```bash
git tag v0.1.0
git push origin v0.1.0
```

3. GitHub Actions workflow [`.github/workflows/release.yml`](../.github/workflows/release.yml) runs tests and GoReleaser.
4. A [GitHub Release](https://github.com/LaProgrammerie/hyper-fast-builder/releases) is published with multi-platform archives and `checksums.txt`.
5. GoReleaser updates `Formula/agentflow.rb` in `LaProgrammerie/homebrew-tap` (requires `HOMEBREW_TAP_GITHUB_TOKEN`).
6. Verify installation:

```bash
brew update
brew install LaProgrammerie/tap/agentflow
agentflow version
agentflow doctor
```

## Local dry run

```bash
make release-check
make release-snapshot
ls -la dist/
```

Snapshot builds write to `dist/` without publishing. Use this to confirm archive layout (binary + `LICENSE` + `README.md` at archive root) before tagging.

## Repository secrets

| Secret | Purpose |
| --- | --- |
| `GITHUB_TOKEN` | Provided by Actions; publishes release assets to **this** repo |
| `HOMEBREW_TAP_GITHUB_TOKEN` | PAT with `contents:write` on `LaProgrammerie/homebrew-tap` only |

Never commit tokens. Do not echo secrets in CI logs.

## Version injection

Builds set `application/internal/version` via ldflags (`Version`, `Commit`, `Date`). `agentflow version` prints:

```
AgentFlow v0.1.0
commit: abc1234
built: 2026-05-17T12:00:00Z
```

## Out of scope (V2)

cosign signatures, SBOM, Debian/RPM, Scoop, Winget, Docker images, macOS notarization.
