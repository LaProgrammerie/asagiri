# ADR-040 — Enterprise gouvernance

**Date :** 2026-06-10  
**Statut :** accepté (spec M4)  
**Spec :** `.kiro/specs/monetization-distribution-v1/m4-enterprise.md`

## Contexte

Organisations régulées exigent SSO, audit immuable, policies et déploiement on-prem tout en conservant le moteur OSS local (ADR-037).

## Décision

1. **Enterprise** = Team Cloud (M3) + SSO (SAML/OIDC) + SCIM + Policy Center + audit trail WORM.
2. **Enforcement policies** côté client **`asagiri-cloud enterprise`** — refus sync si violation — **pas** dans `cmd/asa` core.
3. **Audit trail** append-only hash chaîné ; export SIEM (CEF, JSON Lines).
4. **On-prem / air-gap** : appliance ou Helm ; mises à jour par bundle signé.
5. **Registre packs privé** : format `asagiri-pack-v1` + clés signature organisation.
6. **SLA contractuel** : tiers Standard / Premium — hors repo technique.
7. **Résidence données** : région EU/US ; VPC dédié option.

## Conséquences

- Moteur OSS et inviolables inchangés.
- Clients air-gap peuvent adopter sans SaaS public.
- Break-glass et mode policy `advisory` avant `strict` recommandés au rollout.

## Références

- ADR-037, ADR-039, `m4-enterprise.md`, ADR-038 (packs privés)
