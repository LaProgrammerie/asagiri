#!/usr/bin/env bash
set -euo pipefail

export CURSOR_HOOK_PAYLOAD="$(cat || true)"

python3 <<'PY'
import json
import os
import re
import sys


def emit_empty() -> None:
    print("{}")
    sys.exit(0)


def collect_paths(obj, out: set[str]) -> None:
    if isinstance(obj, dict):
        for key, value in obj.items():
            if key in {"path", "file_path", "target_file", "uri"} and isinstance(value, str):
                out.add(value)
            collect_paths(value, out)
    elif isinstance(obj, list):
        for item in obj:
            collect_paths(item, out)
    elif isinstance(obj, str):
        if re.match(r"^(\\.?/)?([a-zA-Z0-9_.-]+/)+[a-zA-Z0-9_.-]+", obj):
            out.add(obj)


payload = os.environ.get("CURSOR_HOOK_PAYLOAD", "")

try:
    data = json.loads(payload) if payload.strip() else {}
except json.JSONDecodeError:
    emit_empty()

paths: set[str] = set()
collect_paths(data, paths)

prefixes = (
    ".githooks/",
    "scripts/install-git-hooks.sh",
    ".cursor/hooks/",
    ".cursor/hooks.json",
    ".cursor/rules/",
    ".kiro/hooks/",
    ".kiro/steering/",
)


def normalize_repo_path(p: str) -> str:
    n = p.strip()
    if n.startswith("./"):
        return n[2:]
    return n


def is_candidate(p: str) -> bool:
    normalized = normalize_repo_path(p)
    return any(normalized.startswith(pref) for pref in prefixes)


hits = sorted({p for p in paths if is_candidate(p)})

if not hits:
    emit_empty()

lines = [
    "[Template sync] Modified files that may be reusable in template:",
    *[f"- {h}" for h in hits[:12]],
]
if len(hits) > 12:
    lines.append(f"- ... (+{len(hits) - 12} other paths)")

lines.extend(
    [
        "",
        "Decision required: is this change project-specific or should it also go to Generic project?",
        "If generic: port it to template (copy/cherry-pick), exact or adapted on same principle.",
        "See `.cursor/rules/template-generic-sync.mdc` and `.kiro/steering/35-template-generic-sync.md`.",
    ]
)

message = "\n".join(lines)
print(json.dumps({"additional_context": message}))
PY
