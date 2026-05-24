# AgentFlow documentation site

Static [Fumadocs](https://fumadocs.dev) + Next.js export (`docs-site/out/`).

## Local development

```bash
# From repository root — regenerate CLI reference (optional if unchanged)
go run ./application/cmd/agentflow docs generate-cli --output docs-site/content/docs/en/cli/generated

cd docs-site
corepack enable
pnpm install
pnpm dev
```

## Quality gate (matches CI)

```bash
cd docs-site
pnpm install --frozen-lockfile
pnpm typecheck && pnpm lint && pnpm build
test -d out
```

## Deployment (Cloudflare Pages)

Production and PR previews are deployed by [`.github/workflows/docs-cloudflare-pages.yml`](../.github/workflows/docs-cloudflare-pages.yml) via [Wrangler direct upload](https://developers.cloudflare.com/pages/how-to/use-direct-upload-with-continuous-integration/).

Configure these **GitHub repository secrets** (Settings → Secrets and variables → Actions):

| Secret | Description |
|--------|-------------|
| `CLOUDFLARE_API_TOKEN` | API token with permission to deploy to Cloudflare Pages |
| `CLOUDFLARE_ACCOUNT_ID` | Cloudflare account ID |
| `CLOUDFLARE_PAGES_PROJECT` | Pages project name (not the full URL) |

Create the Pages project in the Cloudflare dashboard first, then add the secrets. PRs deploy a preview branch; pushes to `main` deploy production (`--branch=main`).

Forks without these secrets still get validation from [`.github/workflows/docs-check.yml`](../.github/workflows/docs-check.yml) (build only, no deploy).

## Build notes

- `output: 'export'` — no SSR or API routes.
- `basePath` `/hyper-fast-builder` applies only when `GITHUB_PAGES=true` (legacy); Cloudflare custom domain builds use site root.
