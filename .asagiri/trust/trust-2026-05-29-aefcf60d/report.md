# Trust Report

- Trust ID: `trust-2026-05-29-aefcf60d`
- Generated: `2026-05-29T08:10:01.456674Z`
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

- `static-analysis-trust-2026-05-29-aefcf60d` [failed] Static analysis
- `contracts-trust-2026-05-29-aefcf60d` [warn] Contracts
- `flows-trust-2026-05-29-aefcf60d` [passed] Flow integrity
- `permissions-trust-2026-05-29-aefcf60d` [warn] Permissions
- `observability-trust-2026-05-29-aefcf60d` [warn] Observability
- `analytics-trust-2026-05-29-aefcf60d` [warn] Analytics
- `architecture-trust-2026-05-29-aefcf60d` [passed] Architecture
- `security-trust-2026-05-29-aefcf60d` [failed] Security
- `performance-trust-2026-05-29-aefcf60d` [failed] Performance
- `cost-trust-2026-05-29-aefcf60d` [passed] Cost
- `backward-compatibility-trust-2026-05-29-aefcf60d` [warn] Backward compatibility
- `migration-safety-trust-2026-05-29-aefcf60d` [failed] Migration safety
- `blast-radius-trust-2026-05-29-aefcf60d` [warn] Blast radius
- `tests-trust-2026-05-29-aefcf60d` [warn] Tests

## Gate

- Status: `not_configured`
- Reason: verification gates not configured
