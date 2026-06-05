package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	uiapp "github.com/LaProgrammerie/asagiri/application/internal/ui/app"
)

var (
	errGraphCIFailed     = errors.New("execution graph failed CI policy")
	errGraphNotEnabled   = errors.New("execution graph disabled in config")
	errGraphFlowRequired = errors.New("flow required: use --flow")
)

// PlanGraphResult is the structured JSON output for `asa plan graph --json`.
type PlanGraphResult struct {
	Graph     executiongraph.ExecutionGraph    `json:"graph"`
	Schedule  executiongraph.ExecutionSchedule `json:"schedule"`
	Estimate  executiongraph.GraphEstimate     `json:"estimate"`
	Artifacts executiongraph.GraphArtifacts    `json:"artifacts"`
}

// GraphRunJSONResult is the structured JSON output for `asa graph run --json`.
type GraphRunJSONResult struct {
	Graph     executiongraph.ExecutionGraph    `json:"graph"`
	Schedule  executiongraph.ExecutionSchedule `json:"schedule"`
	Estimate  executiongraph.GraphEstimate     `json:"estimate"`
	Artifacts executiongraph.GraphArtifacts    `json:"artifacts"`
	Result    executiongraph.GraphRunResult    `json:"result"`
	DryRun    bool                             `json:"dry_run"`
}

// GraphStatusResult is the structured JSON output for `asa graph status --json`.
type GraphStatusResult struct {
	Graph    executiongraph.ExecutionGraph `json:"graph"`
	Estimate executiongraph.GraphEstimate  `json:"estimate"`
}

// GraphResumeResult is the structured JSON output for `asa graph resume --json`.
type GraphResumeResult struct {
	Result executiongraph.GraphRunResult `json:"result"`
}

func newGraphCmd(dryRun *bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "graph",
		Short: "Exécuter et inspecter les graphes d'exécution",
		RunE:  runUIScreenCommand(dryRun, uiapp.ScreenGraph),
	}
	cmd.AddCommand(
		newGraphRunCmd(),
		newGraphStatusCmd(),
		newGraphResumeCmd(),
		newGraphRollbackCmd(dryRun),
		newGraphVisualizeCmd(),
	)
	return cmd
}

func newPlanGraphCmd() *cobra.Command {
	var (
		flow           string
		fromProduct    bool
		fromSpec       bool
		includeReviews bool
		includeDocs    bool
		estimate       bool
		output         string
		jsonOut        bool
		ci             bool
	)

	cmd := &cobra.Command{
		Use:     "graph <product>",
		Short:   "Générer un graphe d'exécution sans lancer d'agents",
		Example: "  asa plan graph workspace-saas --flow workspace-onboarding",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			if strings.TrimSpace(flow) == "" {
				return errGraphFlowRequired
			}

			result, err := runPlanGraph(cmd.Context(), repoRoot, cfg, graphPlanOptions{
				Product:        args[0],
				Flow:           flow,
				FromProduct:    fromProduct,
				FromSpec:       fromSpec,
				IncludeReviews: includeReviews,
				IncludeDocs:    includeDocs,
				Estimate:       estimate,
				Output:         output,
				JSON:           jsonOut,
				CI:             ci,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if jsonOut {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(result); err != nil {
					return fmt.Errorf("encode plan graph result: %w", err)
				}
			} else if strings.EqualFold(output, "markdown") {
				md, err := executiongraph.Render(result.Graph, executiongraph.RenderFormatMarkdown)
				if err != nil {
					return err
				}
				_, _ = fmt.Fprint(out, md)
			} else {
				_, _ = fmt.Fprint(out, executiongraph.FormatTerminalSummary(result.Graph, result.Schedule, result.Estimate))
				_, _ = fmt.Fprintf(out, "\nArtifacts: %s\n", result.Artifacts.Dir)
			}

			stopOn := executiongraph.RiskLevel(cfg.ExecutionGraph.StopOnRisk)
			if ci && executiongraph.CIShouldFailPlan(result.Estimate, stopOn) {
				return errGraphCIFailed
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flow, "flow", "", "Flow id under .asagiri/products/<product>/flows/")
	cmd.Flags().BoolVar(&fromProduct, "from-product", false, "Plan from product layer inputs")
	cmd.Flags().BoolVar(&fromSpec, "from-spec", false, "Include spec-derived tasks")
	cmd.Flags().BoolVar(&includeReviews, "include-reviews", true, "Include review and validation nodes")
	cmd.Flags().BoolVar(&includeDocs, "include-docs", false, "Include documentation nodes")
	cmd.Flags().BoolVar(&estimate, "estimate", true, "Compute cost and duration estimates")
	cmd.Flags().StringVar(&output, "output", "", "Output format: markdown")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Structured JSON output on stdout")
	cmd.Flags().BoolVar(&ci, "ci", false, "CI mode: conservative scheduling and non-zero exit on policy failure")
	return cmd
}

func newPlanExplainCmd() *cobra.Command {
	var flow string

	cmd := &cobra.Command{
		Use:     "explain <product>",
		Short:   "Expliquer les dépendances, le parallélisme et les risques du plan",
		Example: "  asa plan explain workspace-saas --flow workspace-onboarding",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			if strings.TrimSpace(flow) == "" {
				return errGraphFlowRequired
			}

			result, err := runPlanGraph(cmd.Context(), repoRoot, cfg, graphPlanOptions{
				Product:        args[0],
				Flow:           flow,
				IncludeReviews: true,
				Estimate:       true,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), executiongraph.FormatExplain(result.Graph, result.Schedule))
			return nil
		},
	}
	cmd.Flags().StringVar(&flow, "flow", "", "Flow id under .asagiri/products/<product>/flows/")
	return cmd
}

func newGraphRunCmd() *cobra.Command {
	var (
		flow            string
		maxParallel     int
		stopOnRisk      string
		strictTrust     bool
		budget          float64
		checkpointEvery string
		dryRun          bool
		ci              bool
		jsonOut         bool
	)

	cmd := &cobra.Command{
		Use:     "run <product>",
		Short:   "Planifier et exécuter (ou simuler) un graphe d'exécution",
		Example: "  asa graph run workspace-saas --flow workspace-onboarding --dry-run",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			if strings.TrimSpace(flow) == "" {
				return errGraphFlowRequired
			}
			if err := executiongraph.ValidateCheckpointEvery(checkpointEvery); err != nil {
				return err
			}

			result, runResult, err := runGraphRun(cmd.Context(), repoRoot, cfg, graphRunOptions{
				Product:         args[0],
				Flow:            flow,
				MaxParallel:     maxParallel,
				StopOnRisk:      stopOnRisk,
				StrictTrust:     strictTrust,
				Budget:          budget,
				CheckpointEvery: checkpointEvery,
				DryRun:          dryRun,
				CI:              ci,
			})
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if jsonOut {
				payload := GraphRunJSONResult{
					Graph:     result.Graph,
					Schedule:  result.Schedule,
					Estimate:  result.Estimate,
					Artifacts: result.Artifacts,
					Result:    runResult,
					DryRun:    dryRun,
				}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(payload); err != nil {
					return fmt.Errorf("encode graph run result: %w", err)
				}
			} else {
				_, _ = fmt.Fprint(out, executiongraph.FormatTerminalSummary(result.Graph, result.Schedule, result.Estimate))
				_, _ = fmt.Fprintf(out, "\nRun status: %s\n", runResult.Status)
				_, _ = fmt.Fprintf(out, "Artifacts: %s\n", result.Artifacts.Dir)
				if dryRun {
					_, _ = fmt.Fprintln(out, "Dry-run: no agents executed")
				}
			}

			if ci && executiongraph.CIShouldFailRun(result.Graph, result.Estimate) {
				return errGraphCIFailed
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&flow, "flow", "", "Flow id under .asagiri/products/<product>/flows/")
	cmd.Flags().IntVar(&maxParallel, "max-parallel", 0, "Override max parallel nodes (default from config)")
	cmd.Flags().StringVar(&stopOnRisk, "stop-on-risk", "", "Stop when highest risk reaches this level")
	cmd.Flags().BoolVar(&strictTrust, "strict-trust", false, "Treat trust warnings as failures")
	cmd.Flags().Float64Var(&budget, "budget", 0, "Budget cap in EUR (0 = use config/default)")
	cmd.Flags().StringVar(&checkpointEvery, "checkpoint-every", "", "Checkpoint cadence: node|group")
	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Plan and persist artefacts without executing agents")
	cmd.Flags().BoolVar(&ci, "ci", false, "CI mode: conservative scheduling and non-zero exit on policy failure")
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Structured JSON output on stdout")
	return cmd
}

func newGraphStatusCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:     "status <graph-id>",
		Short:   "Afficher l'état d'un graphe",
		Example: "  asa graph status graph-20260529-001",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			repo := executiongraph.NewRepository(repoRoot)
			graph, err := repo.Load(args[0])
			if err != nil {
				return err
			}
			est := executiongraph.EstimateGraph(graph, nil)

			out := cmd.OutOrStdout()
			if jsonOut {
				payload := GraphStatusResult{Graph: graph, Estimate: est}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(payload); err != nil {
					return fmt.Errorf("encode graph status: %w", err)
				}
				return nil
			}

			_, _ = fmt.Fprintf(out, "Graph: %s\n", graph.ID)
			_, _ = fmt.Fprintf(out, "Product: %s\n", graph.Product)
			_, _ = fmt.Fprintf(out, "Flow: %s\n", graph.Flow)
			_, _ = fmt.Fprintf(out, "Status: %s\n", graph.Status)
			_, _ = fmt.Fprintf(out, "Nodes: %d\n", len(graph.Nodes))
			_, _ = fmt.Fprintf(out, "Edges: %d\n", len(graph.Edges))
			_, _ = fmt.Fprintf(out, "Checkpoints: %d\n", len(graph.Checkpoints))
			_, _ = fmt.Fprintf(out, "Estimated cost: €%.2f\n", est.EstimatedCost)
			_, _ = fmt.Fprintf(out, "Highest risk: %s\n", est.HighestRisk)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Structured JSON output on stdout")
	return cmd
}

func newGraphRollbackCmd(dryRun *bool) *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:     "rollback <graph-id>",
		Short:   "Marquer un graphe et ses nœuds actifs comme rolled back",
		Example: "  asa graph rollback graph-2026-05-29-abcdef01",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			impact, err := executiongraph.AssessRollbackImpact(repoRoot, args[0])
			if err != nil {
				return err
			}
			result, err := executiongraph.ExecuteGraphRollback(cmd.Context(), repoRoot, args[0], dryRun != nil && *dryRun)
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			if jsonOut {
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				return enc.Encode(struct {
					executiongraph.RollbackImpact
					Result executiongraph.GraphRollbackResult `json:"result"`
				}{impact, result})
			}
			_, _ = fmt.Fprintf(out, "%s\n", impact.Title)
			for _, line := range impact.ImpactLines {
				_, _ = fmt.Fprintf(out, "- %s\n", line)
			}
			_, _ = fmt.Fprintf(out, "Graph %s: status=%s nodes_rolled_back=%d dry_run=%t\n", result.GraphID, result.Status, result.NodesRolledBack, result.DryRun)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Structured JSON output on stdout")
	return cmd
}

func newGraphResumeCmd() *cobra.Command {
	var jsonOut bool

	cmd := &cobra.Command{
		Use:     "resume <graph-id>",
		Short:   "Reprendre un graphe interrompu",
		Example: "  asa graph resume graph-20260529-001",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			repo := executiongraph.NewRepository(repoRoot)
			graph, err := repo.Load(args[0])
			if err != nil {
				return err
			}
			runner := executiongraph.NewRunner(repoRoot)
			result, err := runner.Resume(cmd.Context(), args[0], graphRunOptionsFromConfig(repoRoot, cfg, graphRunOptionsFromPersisted(graph)))
			if err != nil {
				return err
			}

			out := cmd.OutOrStdout()
			if jsonOut {
				payload := GraphResumeResult{Result: result}
				enc := json.NewEncoder(out)
				enc.SetIndent("", "  ")
				if err := enc.Encode(payload); err != nil {
					return fmt.Errorf("encode graph resume: %w", err)
				}
				return nil
			}
			_, _ = fmt.Fprintf(out, "Graph %s resumed: status=%s\n", result.GraphID, result.Status)
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "Structured JSON output on stdout")
	return cmd
}

func newGraphVisualizeCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:     "visualize <graph-id>",
		Short:   "Exporter un graphe (mermaid, json, dot, markdown)",
		Example: "  asa graph visualize graph-20260529-001 --format mermaid",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			repoRoot, cfg, err := loadGraphRepoConfig()
			if err != nil {
				return err
			}
			if !cfg.ExecutionGraph.IsEnabled() {
				return errGraphNotEnabled
			}
			repo := executiongraph.NewRepository(repoRoot)
			graph, err := repo.Load(args[0])
			if err != nil {
				return err
			}

			renderFormat := executiongraph.RenderFormat(strings.ToLower(strings.TrimSpace(format)))
			body, err := executiongraph.Render(graph, renderFormat)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprint(cmd.OutOrStdout(), body)
			if renderFormat != executiongraph.RenderFormatJSON && !strings.HasSuffix(body, "\n") {
				_, _ = fmt.Fprintln(cmd.OutOrStdout())
			}
			return nil
		},
	}
	cmd.Flags().StringVar(&format, "format", "mermaid", "Export format: mermaid|json|dot|markdown")
	return cmd
}

type graphPlanOptions struct {
	Product        string
	Flow           string
	FromProduct    bool
	FromSpec       bool
	IncludeReviews bool
	IncludeDocs    bool
	Estimate       bool
	Output         string
	JSON           bool
	CI             bool
}

type graphRunOptions struct {
	Product         string
	Flow            string
	MaxParallel     int
	StopOnRisk      string
	StrictTrust     bool
	Budget          float64
	CheckpointEvery string
	DryRun          bool
	CI              bool
}

func loadGraphRepoConfig() (string, *config.Config, error) {
	startDir, err := os.Getwd()
	if err != nil {
		return "", nil, err
	}
	repoRoot, err := bootstrap.GitRoot(startDir)
	if err != nil {
		return "", nil, err
	}
	cfg, err := config.Load(config.ConfigPath(repoRoot), repoRoot)
	if err != nil {
		return "", nil, err
	}
	return repoRoot, cfg, nil
}

func runPlanGraph(ctx context.Context, repoRoot string, cfg *config.Config, opts graphPlanOptions) (PlanGraphResult, error) {
	planner := executiongraph.NewPlanner(repoRoot)
	graph, err := planner.Build(ctx, executiongraph.GraphPlanRequest{
		Product:        opts.Product,
		Flow:           opts.Flow,
		FromProduct:    opts.FromProduct,
		FromSpec:       opts.FromSpec,
		IncludeReviews: opts.IncludeReviews,
		IncludeDocs:    opts.IncludeDocs,
		Estimate:       opts.Estimate,
		Gates: executiongraph.TrustGateConfig{
			TrustRequiredForHighRisk: cfg.ExecutionGraph.Gates.TrustRequiredForHighRisk,
			HumanApprovalFor:         append([]string(nil), cfg.ExecutionGraph.Gates.HumanApprovalFor...),
		},
	})
	if err != nil {
		return PlanGraphResult{}, err
	}

	applyGraphStrategy(&graph, cfg, strategyOverrides{
		CI: opts.CI,
	})

	sched := executiongraph.DefaultScheduler{}
	schedule, err := sched.Schedule(ctx, executiongraph.ScheduleRequest{
		Graph:  graph,
		CIMode: opts.CI,
	})
	if err != nil {
		return PlanGraphResult{}, err
	}

	repo := executiongraph.NewRepository(repoRoot)
	artifacts, err := repo.SaveAll(graph, &schedule)
	if err != nil {
		return PlanGraphResult{}, err
	}

	est := executiongraph.EstimateGraph(graph, &schedule)
	return PlanGraphResult{
		Graph:     graph,
		Schedule:  schedule,
		Estimate:  est,
		Artifacts: artifacts,
	}, nil
}

func runGraphRun(ctx context.Context, repoRoot string, cfg *config.Config, opts graphRunOptions) (PlanGraphResult, executiongraph.GraphRunResult, error) {
	plan, err := runPlanGraph(ctx, repoRoot, cfg, graphPlanOptions{
		Product:        opts.Product,
		Flow:           opts.Flow,
		IncludeReviews: true,
		Estimate:       true,
		CI:             opts.CI,
	})
	if err != nil {
		return PlanGraphResult{}, executiongraph.GraphRunResult{}, err
	}

	graph := plan.Graph
	applyGraphStrategy(&graph, cfg, strategyOverrides{
		MaxParallel: opts.MaxParallel,
		StopOnRisk:  opts.StopOnRisk,
		Budget:      opts.Budget,
		CI:          opts.CI,
	})

	graph.Strategy.StrictTrust = opts.StrictTrust
	if opts.CheckpointEvery != "" {
		graph.Strategy.CheckpointEvery = opts.CheckpointEvery
	}

	runner := executiongraph.NewRunner(repoRoot)
	runOpts := graphRunOptionsFromConfig(repoRoot, cfg, opts)

	if opts.DryRun {
		runResult, schedule, artifacts, err := runner.DryRun(ctx, graph, opts.CI)
		if err != nil {
			return PlanGraphResult{}, executiongraph.GraphRunResult{}, err
		}
		plan.Graph = graph
		plan.Schedule = schedule
		plan.Artifacts = artifacts
		plan.Estimate = executiongraph.EstimateGraph(graph, &schedule)
		return plan, runResult, nil
	}

	repo := executiongraph.NewRepository(repoRoot)
	artifacts, err := repo.SaveAll(graph, &plan.Schedule)
	if err != nil {
		return PlanGraphResult{}, executiongraph.GraphRunResult{}, err
	}
	plan.Artifacts = artifacts

	runResult, err := runner.Run(ctx, graph, plan.Schedule, runOpts)
	if err != nil {
		return plan, runResult, err
	}
	plan.Graph = graph
	loaded, loadErr := repo.Load(graph.ID)
	if loadErr == nil {
		plan.Graph = loaded
	}
	return plan, runResult, nil
}

type strategyOverrides struct {
	MaxParallel int
	StopOnRisk  string
	Budget      float64
	CI          bool
}

func graphRunOptionsFromPersisted(graph executiongraph.ExecutionGraph) graphRunOptions {
	return graphRunOptions{
		StrictTrust:     graph.Strategy.StrictTrust,
		CheckpointEvery: graph.Strategy.CheckpointEvery,
	}
}

func graphRunOptionsFromConfig(repoRoot string, cfg *config.Config, opts graphRunOptions) executiongraph.RunOptions {
	return executiongraph.RunOptions{
		DryRun:          opts.DryRun,
		CIMode:          opts.CI,
		StrictTrust:     opts.StrictTrust,
		CheckpointEvery: opts.CheckpointEvery,
		Gates:           trust.NewGateEvaluator(&cfg.Verification),
		TrustEngine:     trust.NewEngineForStrict(repoRoot, cfg),
	}
}

func applyGraphStrategy(graph *executiongraph.ExecutionGraph, cfg *config.Config, overrides strategyOverrides) {
	if overrides.MaxParallel > 0 {
		graph.Strategy.MaxParallel = overrides.MaxParallel
	} else if cfg.ExecutionGraph.MaxParallel > 0 {
		graph.Strategy.MaxParallel = cfg.ExecutionGraph.MaxParallel
	}
	if overrides.StopOnRisk != "" {
		graph.Strategy.StopOnRisk = executiongraph.RiskLevel(overrides.StopOnRisk)
	} else if cfg.ExecutionGraph.StopOnRisk != "" {
		graph.Strategy.StopOnRisk = executiongraph.RiskLevel(cfg.ExecutionGraph.StopOnRisk)
	}
	if overrides.Budget > 0 {
		graph.Strategy.Budget = overrides.Budget
	}
	if overrides.CI && graph.Strategy.MaxParallel > 1 {
		graph.Strategy.MaxParallel = 1
	}
}
