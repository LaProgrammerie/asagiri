#!/usr/bin/env bash
# Validate build, tests, and npm tarball without publishing.
set -euo pipefail

root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$root"

npm ci
npm run build
npm test
npm pack --dry-run

echo "publish-dry-run: OK (no package uploaded)"
