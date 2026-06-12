#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HOOKS_DIR="${REPO_ROOT}/.git/hooks"
SOURCE_DIR="${REPO_ROOT}/.githooks"

if [[ ! -d "${HOOKS_DIR}" ]]; then
    echo "Missing .git/hooks directory. Run from repository clone."
    exit 1
fi

install -m 0755 "${SOURCE_DIR}/pre-commit" "${HOOKS_DIR}/pre-commit"
install -m 0755 "${SOURCE_DIR}/pre-push" "${HOOKS_DIR}/pre-push"
install -m 0755 "${SOURCE_DIR}/commit-msg" "${HOOKS_DIR}/commit-msg"
install -m 0755 "${SOURCE_DIR}/_refresh-after-update.sh" "${HOOKS_DIR}/_refresh-after-update.sh"
install -m 0755 "${SOURCE_DIR}/post-merge" "${HOOKS_DIR}/post-merge"
install -m 0755 "${SOURCE_DIR}/post-checkout" "${HOOKS_DIR}/post-checkout"
install -m 0755 "${SOURCE_DIR}/pre-rebase" "${HOOKS_DIR}/pre-rebase"

echo "Git hooks installed:"
echo "- pre-commit"
echo "- pre-push"
echo "- commit-msg"
echo "- post-merge"
echo "- post-checkout"
echo "- pre-rebase"
