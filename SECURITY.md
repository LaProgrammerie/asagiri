# Security Policy

## Supported versions

Security fixes are applied on the default branch (`main`). Tagged releases follow the project roadmap in `ROADMAP.md`.

## Reporting a vulnerability

**Do not** open a public GitHub issue for sensitive security reports.

Please report vulnerabilities privately to the maintainers (GitHub Security Advisories on this repository, or contact via your existing LaProgrammerie channel if you are a customer).

Include:

- Description and impact
- Steps to reproduce
- Affected versions or commits
- Suggested fix (optional)

We aim to acknowledge reports within a few business days.

## Scope

In scope:

- AgentFlow CLI and libraries under `application/`
- Default configuration and local state under `.agentflow/`
- Documentation site build pipeline (no secrets in static output)

Out of scope:

- Third-party coding agents (Kiro, Cursor, Codex, Ollama, etc.) — follow their vendors' policies
- User-provided API keys and environment secrets — never commit them; AgentFlow redacts common patterns in logs but cannot guarantee zero leakage from misconfiguration

## Safe usage

- Run `agentflow doctor` after `init`
- Keep `policies` and `validation` enabled for production repos
- Treat MCP and Notion sync as **experimental** until marked stable in the docs
- Review cloud model routing and budgets before enabling `allow-cloud` flags

See the [security documentation](https://laprogrammerie.github.io/hyper-fast-builder/docs/security/) for the full model.
