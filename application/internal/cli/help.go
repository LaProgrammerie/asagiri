package cli

const rootLong = `AgentFlow — orchestrateur CLI local pour workflows de développement agentique.

Transforme une spec en tâches exécutables, isole le travail dans des git worktrees,
trace chaque run et enchaîne validation + review avant une PR.

Prérequis : dépôt Git. Première utilisation : agentflow init puis agentflow doctor.`

const rootExample = `  Exemple — développer une feature de bout en bout
  ─────────────────────────────────────────────

  # 1. Bootstrap (une fois par dépôt)
  agentflow init
  agentflow doctor
  agentflow index                    # index RAG local (optionnel, utile pour enrich)

  # 2. Spécifier et planifier
  agentflow spec billing-v2 --agent kiro
  agentflow plan billing-v2
  agentflow enrich billing-v2 --task task-003 --agent ollama

  # 3. Implémenter une tâche (worktree dédié + agent)
  agentflow dev billing-v2 --task task-003 --agent cursor

  # 4. Qualité et review
  agentflow verify billing-v2 --task task-003
  agentflow review billing-v2 --task task-003 --agent codex

  # 5. Rapport, statut, PR
  agentflow status
  agentflow report run-2026-05-17-001
  agentflow pr billing-v2

  # Reprise après interruption
  agentflow resume run-2026-05-17-001

  Exemple — intention (specv2)
  ────────────────────────────
  agentflow work "développe billing-v2" --plan-only
  agentflow work "reprends billing-v2" --yes
  agentflow continue
  agentflow next --feature billing-v2
  agentflow inbox --source local
  agentflow sync notion --page https://notion.so/...

  Exemple — une seule tâche (sans repasser par plan)
  ─────────────────────────────────────────────────
  agentflow dev billing-v2 --task task-003 --agent cursor
  agentflow verify billing-v2 --task task-003

  Essai sans agents réels (CI / découverte)
  ─────────────────────────────────────────
  agentflow --dry-run plan billing-v2
  agentflow --dry-run dev billing-v2 --task task-001 --agent cursor

  Aide détaillée : agentflow <commande> --help`
