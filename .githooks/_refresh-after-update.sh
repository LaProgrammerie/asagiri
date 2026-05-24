#!/usr/bin/env bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "${REPO_ROOT}"

needs_install_for_changed_files() {
    local changed_files="$1"
    [[ -z "${changed_files}" ]] && return 1

    grep -Eq '^(application/composer\.lock|application/package-lock\.json|application/importmap\.php|application/composer\.json|application/package\.json)$' <<< "${changed_files}"
}

warn_if_branch_behind_upstream() {
    if ! git rev-parse --abbrev-ref --symbolic-full-name '@{u}' >/dev/null 2>&1; then
        return 0
    fi

    local counts behind
    counts="$(git rev-list --left-right --count '@{u}...HEAD' 2>/dev/null || true)"
    [[ -z "${counts}" ]] && return 0

    behind="$(awk '{print $1}' <<< "${counts}")"
    if [[ "${behind}" =~ ^[0-9]+$ ]] && (( behind > 0 )); then
        echo "Info: branch is behind upstream by ${behind} commit(s). Consider: git pull --rebase"
    fi
}

run_castor_install_if_needed() {
    local changed_files="$1"

    if needs_install_for_changed_files "${changed_files}"; then
        echo "Hooks: dependency files changed, running castor install…"
        castor install
    else
        echo "Hooks: no dependency lock/config change, skipping castor install."
    fi
}
