package graphcli

import (
	"context"
	"os"

	"github.com/LaProgrammerie/asagiri/application/internal/bootstrap"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
)

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

type strategyOverrides struct {
	MaxParallel int
	StopOnRisk  string
	Budget      float64
	CI          bool
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

// GraphRunOptionsFromPersisted maps persisted graph strategy to run CLI options.
func GraphRunOptionsFromPersisted(graph executiongraph.ExecutionGraph) graphRunOptions {
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
