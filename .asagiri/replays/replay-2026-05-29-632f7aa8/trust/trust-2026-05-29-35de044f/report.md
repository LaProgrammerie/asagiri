# Trust Report

- Trust ID: `trust-2026-05-29-35de044f`
- Generated: `2026-05-29T08:20:18.914887Z`
- Flow: `workspace-onboarding`
- Residual risk: **high**

## Confidence

- Architecture: 56%
- Implementation: 0%
- Flow integrity: 58%
- Observability: 36%
- Security: 0%
- Regression: 0%
- Overall: 25%

## Warnings

- confidence aggregated from verification checks with per-dimension weighting (spec §7, §11)

## Blast Radius

Flows impacted: 5
Critical APIs: 0
Shared modules: 127
Migration risk: high
Public contract risk: high

## Checks

- `static-analysis-trust-2026-05-29-35de044f` [failed] Static analysis
- `contracts-trust-2026-05-29-35de044f` [warn] Contracts
- `flows-trust-2026-05-29-35de044f` [passed] Flow integrity
- `permissions-trust-2026-05-29-35de044f` [warn] Permissions
- `observability-trust-2026-05-29-35de044f` [warn] Observability
- `analytics-trust-2026-05-29-35de044f` [warn] Analytics
- `architecture-trust-2026-05-29-35de044f` [passed] Architecture
- `security-trust-2026-05-29-35de044f` [failed] Security
- `performance-trust-2026-05-29-35de044f` [failed] Performance
- `cost-trust-2026-05-29-35de044f` [passed] Cost
- `backward-compatibility-trust-2026-05-29-35de044f` [warn] Backward compatibility
- `migration-safety-trust-2026-05-29-35de044f` [failed] Migration safety
- `blast-radius-trust-2026-05-29-35de044f` [warn] Blast radius
- `tests-trust-2026-05-29-35de044f` [warn] Tests

## Suggested reviews

- **product**: overall confidence 25% below 70%
- **architecture**: elevated blast radius on flows or public contracts
- **observability**: observability check: observability contract has no requirements
- **security**: security check: sensitive step step-2 without auth=[REDACTED]
- **performance**: performance check: investigation: config nil

## Gate

- Status: `not_configured`
- Reason: verification gates not configured
