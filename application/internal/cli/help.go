package cli

const rootLong = `Asagiri — orchestrateur CLI local pour workflows de développement agentique.

Transforme une spec en tâches exécutables, isole le travail dans des git worktrees,
trace chaque run et enchaîne validation + review avant une PR.

Prérequis : dépôt Git. Première utilisation : asa init puis asa doctor.

Pour commencer (chemin guidé) :
  1. asa onboard           # préparer le dépôt (état prêt)
  2. asa work "<besoin>"   # décrire → produire → valider aux jalons
  3. (jalons) confirmation de plan, budget, actions sensibles → validation humaine

Commandes unitaires (toujours disponibles) : asa spec | plan | enrich | dev | verify | review …
Aide détaillée : asa <commande> --help`

const rootExample = `  Exemple — développer une feature de bout en bout
  ─────────────────────────────────────────────

  # 1. Bootstrap (une fois par dépôt)
  asa init
  asa doctor
  asa index                    # index RAG local (optionnel, utile pour enrich)

  # 2. Spécifier et planifier
  asa spec billing-v2 --agent kiro
  asa plan billing-v2
  asa enrich billing-v2 --task task-003 --agent ollama

  # 3. Implémenter une tâche (worktree dédié + agent)
  asa dev billing-v2 --task task-003 --agent cursor

  # 4. Qualité et review
  asa verify billing-v2 --task task-003
  asa review billing-v2 --task task-003 --agent codex

  # 5. Rapport, statut, PR
  asa status
  asa report run-2026-05-17-001
  asa pr billing-v2

  # Reprise après interruption
  asa resume run-2026-05-17-001

  Exemple — intention (specv2)
  ────────────────────────────
  asa work "développe billing-v2" --plan-only
  asa work "reprends billing-v2" --yes
  asa continue
  asa next --feature billing-v2
  asa inbox --source local
  asa sync notion --page https://notion.so/...

  Exemple — une seule tâche (sans repasser par plan)
  ─────────────────────────────────────────────────
  asa dev billing-v2 --task task-003 --agent cursor
  asa verify billing-v2 --task task-003

  Essai sans agents réels (CI / découverte)
  ─────────────────────────────────────────
  asa --dry-run plan billing-v2
  asa --dry-run dev billing-v2 --task task-001 --agent cursor

  Aide détaillée : asa <commande> --help`
