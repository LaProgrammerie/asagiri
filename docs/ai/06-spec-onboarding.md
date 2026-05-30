# Spec-onboarding — Project Onboarding & Readiness (canon `docs/ai`)

**Statut :** **FULL livré** + **TUI wizard interactif** (`2026-05-30`)  
**Spec racine :** [`spec-onboarding.md`](archives/specs/spec-onboarding.md)  
**Handoff :** [`active/handoff.md`](active/handoff.md)

---

## 1. Résumé

Spec-onboarding introduit un parcours **`asa onboard`** qui :

1. **détecte** la stack du dépôt (Go, PHP/Castor, Node…) ;
2. **génère ou adapte** `.asagiri/config.yaml` (validation, agents, specs) ;
3. **valide** prérequis via doctor étendu et **`asa ready`** ;
4. **bootstrap** le canon minimal (`docs/ai/`, première feature `.kiro/specs/`) ;
5. laisse le projet **prêt** pour `asa work` et Mission Control.

L’onboarding ne spécifie pas le métier — il prépare le socle technique.

---

## 2. Commandes livrées (cible)

| Commande | Rôle |
|----------|------|
| `asa onboard` | Wizard CLI/TUI |
| `asa onboard --check-only` | Readiness sans mutation |
| `asa ready` | Alias readiness (`--json` pour CI) |
| `asa doctor --full` | Doctor + checks onboarding |

---

## 3. Architecture (cible)

| Zone | Rôle |
|------|------|
| `application/internal/onboarding/` | Détection stack, wizard, writer config, readiness |
| `application/internal/onboarding/detect/` | Détecteurs Go, Castor/PHP, Node |
| `application/internal/bootstrap/` | Extension doctor (checks onboarding) |
| `application/internal/cli/onboard_cmd.go` | Surface Cobra |
| `application/internal/ui/screens/onboarding/` | Wizard TUI interactif (`Model` + formulaire) |
| `application/internal/ui/bus/` | Queries/commands readiness |

---

## 4. Readiness (contrat)

Score 0–100 ; `ready: true` si aucun check **error**.

Checks minimum : config valide, git, `.gitignore` Asagiri, validation commands définies,
agent default détectable, spec Kiro présente, `01-product.md` non vide.

---

## 5. Phasage

| Lot | Focus |
|-----|-------|
| 1 | Core CLI + detect + ready |
| 2 | TUI + Mission Control bandeau |
| 3 | Docs site + reprise `--resume` |
| 4 | Wizard TUI interactif (`asa onboard --ui`) |

## TUI wizard (lot 4)

- `asa onboard --ui` : formulaire Bubble Tea full-screen (pas un bilan read-only)
- Étapes : Welcome → Project → Stack → Agents → Docs → Feature → Review → Apply → Ready
- Champs préremplis (detect, config, dirname) ; panneau **Advanced** repliable
- Bus : `GetOnboardingWizard`, `AdvanceOnboardingStep`, `SetOnboardingField`, `ApplyOnboardingConfig`
- Logique métier dans `internal/onboarding/form.go` ; UI consomme le bus (ADR-027)

Voir matrice complète dans [`active/handoff.md`](active/handoff.md).
