#!/usr/bin/env bash
set -euo pipefail

python3 <<'PY'
import json
import os
import subprocess
import sys
from pathlib import Path


def run(argv: list[str]) -> str:
    try:
        return subprocess.check_output(argv, stderr=subprocess.DEVNULL, text=True).strip()
    except (subprocess.CalledProcessError, FileNotFoundError):
        return ""


def run_lines(argv: list[str]) -> list[str]:
    output = run(argv)
    if not output:
        return []
    return [line for line in output.splitlines() if line.strip()]


def emit(payload: dict) -> None:
    print(json.dumps(payload))
    sys.exit(0)


def main() -> None:
    project_root = run(["git", "rev-parse", "--show-toplevel"])
    if not project_root:
        emit({})

    env_file = Path(project_root) / ".cursor" / "template-sync.env"
    generic_root = os.environ.get("GENERIC_TEMPLATE_ROOT", "").strip()

    if env_file.is_file():
        for raw in env_file.read_text(encoding="utf-8").splitlines():
            line = raw.strip()
            if not line or line.startswith("#"):
                continue
            if line.startswith("GENERIC_TEMPLATE_ROOT="):
                _, _, value = line.partition("=")
                generic_root = value.strip().strip('"').strip("'")
                break

    generic_path = Path(generic_root).expanduser() if generic_root else None
    if not generic_path or not generic_path.is_dir():
        emit({})

    project_path = Path(project_root).resolve()
    if generic_path.resolve() == project_path:
        emit({})

    lines: list[str] = []
    lines.append("[Template sync] Depot template reference: `{0}`".format(generic_path))

    log_generic = run_lines(
        ["git", "-C", str(generic_path), "log", "-8", "--oneline", "--no-decorate", "--no-color"]
    )
    if log_generic:
        lines.append("")
        lines.append("Recent commits on template (import if relevant):")
        lines.extend(["  " + ln for ln in log_generic[:8]])

    lines.append("")
    lines.append("Recently modified files in template (all areas):")
    recent_changed_files = run_lines(
        [
            "git",
            "-C",
            str(generic_path),
            "log",
            "-10",
            "--name-only",
            "--pretty=format:",
            "--no-color",
        ]
    )
    unique_recent_files = sorted(set(recent_changed_files))
    if unique_recent_files:
        preview = unique_recent_files[:30]
        lines.extend([f"- {rel}" for rel in preview])
        if len(unique_recent_files) > len(preview):
            lines.append(f"- ... (+{len(unique_recent_files) - len(preview)} more files)")
    else:
        lines.append("- No file listed (empty log or unavailable).")

    lines.append("")
    lines.append("Quick template -> project comparison (recent file presence):")

    missing_from_project: list[str] = []
    different_content: list[str] = []
    for rel in unique_recent_files[:50]:
        left = project_path / rel
        right = generic_path / rel
        if not right.exists():
            continue
        if not left.exists():
            missing_from_project.append(rel)
            continue
        if left.is_file() and right.is_file():
            proc = subprocess.run(
                ["diff", "-q", str(left), str(right)],
                stdout=subprocess.PIPE,
                stderr=subprocess.DEVNULL,
                text=True,
            )
            if proc.returncode != 0:
                different_content.append(rel)

    if missing_from_project:
        lines.append("- Present in template but missing here:")
        lines.extend([f"  - {p}" for p in missing_from_project[:15]])
        if len(missing_from_project) > 15:
            lines.append(f"  - ... (+{len(missing_from_project) - 15})")

    if different_content:
        lines.append("- Present in both repos but different:")
        lines.extend([f"  - {p}" for p in different_content[:15]])
        if len(different_content) > 15:
            lines.append(f"  - ... (+{len(different_content) - 15})")

    if not missing_from_project and not different_content:
        lines.append("- Nothing obvious to import from inspected recent files.")

    lines.extend(
        [
            "",
            "Action: classify each item pragmatically as",
            '"exact import", "same principle adaptation", or "project-specific".',
            "Goal: logic alignment, not mandatory byte-to-byte copy.",
            "Variables: `.cursor/template-sync.env.example`.",
        ]
    )

    emit({"additional_context": "\n".join(lines)})


if __name__ == "__main__":
    main()
PY
