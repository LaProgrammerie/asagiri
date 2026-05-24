# Consolidation Asagiri (spec-postv123, historique AgentFlow)

Index des livrables de la mission consolidation / open source readiness.

| Doc | Sujet |
|-----|--------|
| [01-architecture-audit.md](01-architecture-audit.md) | Architecture, drift, décisions |
| [02-api-stability.md](02-api-stability.md) | CLI, primitives, `pkg/asagiri` |
| [03-security-reliability-audit.md](03-security-reliability-audit.md) | Sécurité, fiabilité, garde-fous |
| [04-performance-cost-audit.md](04-performance-cost-audit.md) | Performance, tokens, benchmark |
| [05-agent-workflows.md](05-agent-workflows.md) | Matrice tests agents |
| [06-quality-roadmap.md](06-quality-roadmap.md) | Couverture, tests, roadmap qualité |
| [08-oss-readiness.md](08-oss-readiness.md) | Score OSS /100 |
| [09-explainability.md](09-explainability.md) | Explainability & confiance |
| [TODO-prioritized.md](TODO-prioritized.md) | TODO priorisée |
| [rationalization-plan.md](rationalization-plan.md) | Simplifications proposées |

## Scores (2026-05-17)

| Dimension | Score | Commentaire |
|-----------|-------|-------------|
| **Open source readiness** | **74/100** | LICENSE Apache 2.0, CONTRIBUTING, ROADMAP, examples, README enrichi ; CI release et badges manquants |
| **Confiance / fiabilité** | **71/100** | MCP off par défaut, subprocess sans shell, validations config ; couverture workflow/intent encore faible |

## Specs

- Mission : [`spec-postv123.md`](../../spec-postv123.md)
- Historique : `spec.md`, `specv2.md`, `specv3.md`
