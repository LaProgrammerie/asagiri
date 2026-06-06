package cli

import (
	"context"
	"errors"
	"os"
	"time"

	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/pipeline"
	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/spf13/cobra"
)

func newWorkCmd(dryRun *bool) *cobra.Command {
	var (
		agent          string
		reviewer       string
		sourceName     string
		planOnly       bool
		yes            bool
		maxTasks       int
		stopAfter      string
		noReview       bool
		instruction    string
		estimateOnly   bool
		budgetEUR      float64
		preferLocal    bool
		maxInTok       int
		maxOutTok      int
		maxDurationMin int
		showCtxPlan    bool
		noCloud        bool
		allowCloud     bool
		allowOver      bool
		noCtxReduce       bool
		investigateFirst       bool
		investigateOnFailure   bool
		investigationDepth     string
		strictTrust            bool
		trustFlow              string
	)

	cmd := &cobra.Command{
		Use:   `work "<instruction>"`,
		Short: "Exécuter une intention en langage naturel",
		Long: `Exécute une intention en langage naturel (resolve → plan → pipeline technique).

Classification V1 (déterministe, sans LLM) — priorité : technical_task > feature_work > product_level_intent.
En cas de doute, feature_work : le Product Layer ne se déclenche pas.

  technical_task      → workflow technique normal (bug, test, refactor, CI, …)
  feature_work        → workflow technique normal (endpoint, page, export, …)
  product_level_intent → Product Layer puis arrêt contrôlé (pas d’auto-enchaînement)

Workflow produit en deux temps (V1) :

  Étape 1 — préparation produit (une invocation) :
    asa work "Créer un CRM pour artisans" --yes
    → product model, prototype, flows, contracts, specs, tasks
    → arrêt après génération (pas de dev automatique dans la même commande)

  Étape 2 — implémentation (invocation séparée) :
    asa work crm-artisans --yes
    ou asa next --feature crm-artisans

Prévisualisation : --dry-run (plan Product Layer) ; --plan-only (plan produit uniquement, sans plan technique).
Commandes expertes (asa prototype, flows, contracts, spec generate-from-product) restent optionnelles.`,
		Example: `  # Feature technique
  asa work "ajoute export CSV" --dry-run

  # Intention produit — étape 1 (préparation)
  asa work "Créer un CRM pour artisans" --dry-run
  asa work "Créer un CRM pour artisans" --yes

  # Intention produit — étape 2 (implémentation)
  asa work crm-artisans --yes
  asa next --feature crm-artisans`,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			instruction = args[0]
			startDir, err := os.Getwd()
			if err != nil {
				return err
			}
			actx, err := loadContext(startDir, *dryRun)
			if err != nil {
				return err
			}
			defer actx.Close()

			snap, err := actx.snapshot()
			if err != nil {
				return err
			}

			depth := investigation.Depth(investigationDepth)
			if depth == "" {
				depth = investigation.DepthStandard
			}
			if investigateFirst && !estimateOnly {
				invReq := investigation.Request{
					Symptom:      instruction,
					Feature:      instruction,
					Depth:        depth,
					NoCloud:      noCloud,
					EstimateOnly: false,
					RepoRoot:     actx.RepoRoot,
				}
				rep, invErr := investigation.RunInvestigation(cmd.Context(), invReq, actx.Config)
				if invErr != nil {
					return invErr
				}
				_ = investigation.FeedMemory(actx.RepoRoot, rep)
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "investigate-first: %s (candidates: %d)\n",
					rep.ID, len(rep.RootCauseCandidates))
			}

			interactive := isInteractive()
			opts := intent.WorkOptions{
				PlanOnly:    planOnly,
				Yes:         yes,
				DryRun:      actx.DryRun || *dryRun,
				MaxTasks:    maxTasks,
				StopAfter:   stopAfter,
				NoReview:    noReview,
				Agent:       agent,
				Reviewer:    reviewer,
				Interactive: interactive,
			}

			resolver := intent.NewHybridResolver()
			resolved, err := resolver.Resolve(cmd.Context(), intent.IntentInput{
				RawInstruction: instruction,
				WorkingDir:     actx.RepoRoot,
				Config:         actx.Config,
				StateSnapshot:  snap,
				Interactive:    interactive,
			})
			if err != nil {
				return err
			}
			if sourceName != "" {
				resolved.Source = sourceName
			}

			scope := intent.ClassifyIntentScope(instruction)
			layerOut, layerErr := handleProductLayer(cmd.OutOrStdout(), instruction, scope, resolved, opts, actx.RepoRoot)
			if layerErr != nil {
				return layerErr
			}
			if layerOut.ResolvedPatch.Feature != "" {
				resolved.Feature = layerOut.ResolvedPatch.Feature
			}
			if layerOut.StopWorkFlow {
				return nil
			}

			planner := &intent.DefaultPlanner{}
			plan, err := planner.BuildPlan(cmd.Context(), resolved, snap, actx.Config, opts)
			if err != nil {
				return err
			}

			v3opts := pipeline.V3Options{
				EstimateOnly:    estimateOnly,
				BudgetMajor:     budgetEUR,
				PreferLocal:     preferLocal,
				MaxInputTokens:  maxInTok,
				MaxOutputTokens: maxOutTok,
				ShowContextPlan: showCtxPlan,
				NoCloud:         noCloud,
				AllowCloud:      allowCloud,
				AllowOverBudget: allowOver,
				NoContextReduce: noCtxReduce,
				Interactive:     interactive,
				PlanOnly:        planOnly,
				DryRun:          opts.DryRun,
				Yes:             yes,
				Agent:           agent,
				Reviewer:        reviewer,
				MaxTasks:        maxTasks,
				StopAfter:       stopAfter,
				NoReview:        noReview,
			}
			if maxDurationMin > 0 {
				v3opts.MaxDuration = time.Duration(maxDurationMin) * time.Minute
			}

			if actx.Config.Intent.DefaultMode == "guided" && !opts.Yes && !opts.PlanOnly && !estimateOnly && actx.Config.Work.RequirePlanConfirmation {
				if err := requireConfirm(opts, "Proceed with execution plan?"); err != nil {
					return err
				}
			}

			app := pipeline.App{
				RepoRoot: actx.RepoRoot,
				Config:   actx.Config,
				Store:    actx.Store,
				Executor: &intent.Executor{
					Workflow: actx.Workflow(),
					Config:   actx.Config,
					SyncFn:   actx.syncPrimitive,
				},
			}

			v3res, err := pipeline.RunV3PreFlight(context.Background(), app, resolved, plan, v3opts)
			if err != nil {
				var pc *cost.BudgetPendingConfirmError
				if errors.As(err, &pc) && !yes {
					printEstimateBoxed(cmd.OutOrStdout(), v3res.Estimate, &v3res.Optimize)
					if confirmErr := requireConfirm(opts, pc.Error()); confirmErr != nil {
						return confirmErr
					}
					v3opts.UserConfirmedBudget = true
					v3res, err = pipeline.RunV3PreFlight(context.Background(), app, resolved, plan, v3opts)
				}
				if err != nil {
					if investigateOnFailure && !actx.DryRun && !*dryRun {
						runInvestigationOnFailure(cmd, actx, instruction, depth, noCloud)
					}
					return err
				}
			}

			printEstimateBoxed(cmd.OutOrStdout(), v3res.Estimate, &v3res.Optimize)

			if estimateOnly || planOnly {
				return nil
			}

			v3res, err = pipeline.RunV3Execute(context.Background(), app, resolved, plan, v3opts, v3res)
			if err != nil {
				if investigateOnFailure && !actx.DryRun && !*dryRun {
					runInvestigationOnFailure(cmd, actx, instruction, depth, noCloud)
				}
				return err
			}

			intent.PrintWorkReport(cmd.OutOrStdout(), resolved, plan, v3res.Exec)
			printWorkSummary(cmd.OutOrStdout(), instruction, v3res.Estimate, v3res.Exec)

			if strictTrust && !opts.DryRun {
				flow, product, err := resolveWorkStrictTrust(actx.RepoRoot, trustFlow, resolved.Feature)
				if err != nil {
					return err
				}
				eng := trust.NewEngineForStrict(actx.RepoRoot, actx.Config)
				if store, err := runtime.Open(actx.RepoRoot); err == nil {
					defer func() { _ = store.Close() }()
					eng.Emitter = trust.NewRuntimeEmitter(store)
				}
				eng.Config = actx.Config
				result, err := trust.RunStrictTrust(cmd.Context(), eng, flow, "", product)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(cmd.OutOrStdout(), trust.FormatTerminalSummary(result.Report))
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nstrict-trust: passed (trust id %s)\n", result.TrustID)
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&agent, "agent", "", "Agent d'implémentation")
	cmd.Flags().StringVar(&reviewer, "reviewer", "", "Agent de review")
	cmd.Flags().StringVar(&sourceName, "source", "", "Source externe (notion|local)")
	cmd.Flags().BoolVar(&planOnly, "plan-only", false, "Afficher le plan sans exécuter")
	cmd.Flags().BoolVar(&yes, "yes", false, "Mode auto sans confirmation")
	cmd.Flags().IntVar(&maxTasks, "max-tasks", 0, "Nombre max de tâches par run")
	cmd.Flags().StringVar(&stopAfter, "stop-after", "", "Arrêter après une étape (verify, dev, …)")
	cmd.Flags().BoolVar(&noReview, "no-review", false, "Désactiver la review")
	cmd.Flags().BoolVar(&estimateOnly, "estimate-only", false, "Estimation sans exécution")
	cmd.Flags().Float64Var(&budgetEUR, "budget", 0, "Budget run EUR")
	cmd.Flags().BoolVar(&preferLocal, "prefer-local", false, "Préférer étapes locales")
	cmd.Flags().IntVar(&maxInTok, "max-input-tokens", 0, "Plafond tokens entrée")
	cmd.Flags().IntVar(&maxOutTok, "max-output-tokens", 0, "Plafond tokens sortie")
	cmd.Flags().IntVar(&maxDurationMin, "max-duration", 0, "Durée max (minutes)")
	cmd.Flags().BoolVar(&showCtxPlan, "show-context-plan", false, "Afficher chemin context pack")
	cmd.Flags().BoolVar(&noCloud, "no-cloud", false, "Interdire cloud")
	cmd.Flags().BoolVar(&allowCloud, "allow-cloud", false, "Autoriser cloud explicitement")
	cmd.Flags().BoolVar(&allowOver, "allow-over-budget", false, "Dépasser le budget")
	cmd.Flags().BoolVar(&noCtxReduce, "no-context-reduction", false, "Désactiver réduction contexte")
	cmd.Flags().BoolVar(&investigateFirst, "investigate-first", false, "Lancer une investigation avant le plan")
	cmd.Flags().BoolVar(&investigateOnFailure, "investigate-on-failure", false, "Investigation locale si l'exécution échoue")
	cmd.Flags().StringVar(&investigationDepth, "investigation-depth", "standard", "Profondeur investigation: quick|standard|deep|ci")
	cmd.Flags().BoolVar(&strictTrust, "strict-trust", false, "Après implémentation, enchaîner verify trust + gates")
	cmd.Flags().StringVar(&trustFlow, "trust-flow", "", "Flow produit pour --strict-trust (défaut: feature résolue)")
	return cmd
}

func runInvestigationOnFailure(cmd *cobra.Command, actx *appContext, instruction string, depth investigation.Depth, noCloud bool) {
	req := investigation.Request{
		Symptom:         instruction,
		Feature:         instruction,
		Depth:           depth,
		FromFailedTests: true,
		NoCloud:         noCloud,
		RepoRoot:        actx.RepoRoot,
	}
	rep, err := investigation.RunInvestigation(cmd.Context(), req, actx.Config)
	if err != nil {
		return
	}
	_ = investigation.FeedMemory(actx.RepoRoot, rep)
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), "investigate-on-failure: %s\n", rep.ID)
}
