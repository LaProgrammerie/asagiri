# Handoff — execution

> **Prescriptive contract** for Cursor / Copilot / implementation.  
> **Tranche `spec-my-A` complète** (`2026-05-27`).

## Immediate objective

`spec-my-A.md` sections 1–26 : implémentation **complète** sans raccourci documenté.

## Definition of Done — spec-my-A (intégral)

### Blocs A–D
- [x] Couche produit exécutable (§1–22, §19)
- [x] Business intent (§23)
- [x] Runtime persistant (§24.3–24.15, §24.21)
- [x] Runtime modes config `runtime.mode` (§24.17)
- [x] Runtime API HTTP + **Unix socket** (§24.18)
- [x] SDK Go in-process + HTTP + **SDK TypeScript** (`sdk/typescript`)
- [x] Runtime observability metrics (§24.19)
- [x] Terminal UX enrichi `asa daemon status --rich` (§24.20)
- [x] Memory engine + **embeddings** `internal/memory/embedding.go` (§24.10)
- [x] Analysis layer complète (§24.16)
- [x] Investigation complète (§25) dont `context-pack.md`, `investigate impact`, `work --investigate-on-failure`

### Documentation
- [x] docs-site EN + FR (pages runtime, investigate, analysis, SDK — pas de stubs)

## Hors scope

- Commit / push par l'agent
- Investigation cloud par défaut

## Validation

```bash
cd application && go test ./...
cd sdk/typescript && npm test
asa daemon status --rich
asa runtime serve --socket .asagiri/runtime/runtime.sock
asa investigate impact --flow onboarding --change "async invitations"
asa memory list --query "onboarding failure"
```
