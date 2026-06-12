package pipeline

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/contextopt"
	"github.com/LaProgrammerie/asagiri/application/internal/cost"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/telemetry"
)

// V3Options controls cost-aware preprocessing and execution gates (specv3 §12).
type V3Options struct {
	EstimateOnly        bool
	BudgetMajor         float64
	PreferLocal         bool
	MaxInputTokens      int
	MaxOutputTokens     int
	MaxDuration         time.Duration
	ShowContextPlan     bool
	NoCloud             bool
	AllowCloud          bool
	AllowOverBudget     bool
	UserConfirmedBudget bool
	NoContextReduce     bool
	Interactive         bool
	RunID               string
	PlanOnly            bool
	DryRun              bool
	Yes                 bool
	Agent               string
	Reviewer            string
	MaxTasks            int
	StopAfter           string
	NoReview            bool
}

// V3Result aggregates V3 pipeline outputs for CLI / reports.
type V3Result struct {
	Estimate      cost.ExecutionEstimate
	Investigation investigation.InvestigationResult
	ContextPack   contextopt.ContextPack
	Optimize      contextopt.OptimizeResult
	MetricsRunID  string
	Exec          intent.ExecuteResult
}

type sqliteHistory struct {
	store *sqlite.Store
}

func (h sqliteHistory) DurationsFor(stepName, model string, limit int) []time.Duration {
	if h.store == nil {
		return nil
	}
	d, err := h.store.DurationSamples(stepName, model, limit)
	if err != nil {
		return nil
	}
	return d
}

// App holds minimal dependencies for the V3 pipeline (no TUI).
type App struct {
	RepoRoot string
	Config   *config.Config
	Store    *sqlite.Store
	Executor *intent.Executor
}

// RunV3Pipeline runs investigation → context → estimate → budget → optional execution (specv3).
func RunV3Pipeline(ctx context.Context, app App, resolved intent.ResolvedIntent, plan intent.ExecutionPlan, opts V3Options) (V3Result, error) {
	pre, err := RunV3PreFlight(ctx, app, resolved, plan, opts)
	if err != nil {
		return pre, err
	}
	if opts.EstimateOnly {
		return pre, nil
	}
	return RunV3Execute(ctx, app, resolved, plan, opts, pre)
}

// RunV3PreFlight runs local investigation, context optimization, estimate, and budget checks.
func RunV3PreFlight(ctx context.Context, app App, resolved intent.ResolvedIntent, plan intent.ExecutionPlan, opts V3Options) (V3Result, error) {
	var out V3Result
	if app.Config == nil {
		return out, fmt.Errorf("pipeline v3: config nil")
	}
	runID := opts.RunID
	if runID == "" {
		runID = uuid.NewString()
	}
	out.MetricsRunID = runID

	inv, err := investigation.Run(ctx, app.RepoRoot, resolved.Feature, resolved.TaskID, app.Config)
	if err != nil {
		return out, err
	}
	out.Investigation = inv

	entries, err := contextopt.CollectForPipeline(app.RepoRoot, resolved.Feature, app.Config, contextopt.CollectOpts{}, inv.CandidateFiles)
	if err != nil {
		return out, err
	}
	rawTok := contextopt.SumEntryTokens(entries, app.Config.TokenEst)

	keywords := resolved.Feature + " " + resolved.TaskID
	contextopt.ScoreByKeywords(entries, keywords, resolved.Feature)

	var reduced []contextopt.FileEntry
	var redWarns []string
	if opts.NoContextReduce {
		reduced = entries
	} else {
		reduced, redWarns = contextopt.Reduce(entries, app.Config, contextopt.ReduceOpts{})
	}

	packIn := contextopt.PackInput{
		Feature:         resolved.Feature,
		TaskID:          resolved.TaskID,
		Inv:             inv,
		ReducedFiles:    reduced,
		OutputFormat:    "markdown",
		TaskDescription: plan.Intent.Reason,
	}
	out.ContextPack = contextopt.BuildPack(app.Config, packIn)
	out.Optimize = contextopt.ComputeOptimize(entries, reduced, out.ContextPack, app.Config.TokenEst)
	out.Optimize.Warnings = redWarns
	if rawTok > 0 && out.Optimize.OriginalTokens == 0 {
		out.Optimize.OriginalTokens = rawTok
	}
	if opts.ShowContextPlan {
		_ = redWarns
	}

	dm := cost.DefaultDurationModel{Cfg: app.Config, Reader: sqliteHistory{store: app.Store}}
	est, err := cost.BuildEstimate(ctx, plan, inv, out.ContextPack, app.Config, dm, cost.BuildOpts{
		RunID:               runID,
		MaxOutputTokens:     opts.MaxOutputTokens,
		DefaultOutputTokens: 4096,
		PreferLocal:         opts.PreferLocal,
		NoCloud:             opts.NoCloud,
		AllowCloud:          opts.AllowCloud,
	})
	if err != nil {
		return out, err
	}
	out.Estimate = est

	over := cost.BudgetOverrides{
		MaxCostMajor:     opts.BudgetMajor,
		AllowOverBudget:  opts.AllowOverBudget || opts.UserConfirmedBudget,
		InteractiveAllow: opts.Interactive,
	}
	check, err := cost.CheckBudget(est, app.Config, over)
	if err != nil {
		return out, err
	}
	out.Estimate.BudgetStatus = check.Status
	switch check.Status {
	case cost.BudgetBlock:
		return out, fmt.Errorf("budget: %s", check.Reason)
	case cost.BudgetConfirm:
		if !opts.AllowOverBudget && !opts.UserConfirmedBudget {
			return out, &cost.BudgetPendingConfirmError{Reason: check.Reason}
		}
	}

	if app.Store != nil {
		_ = savePlannedStepMetrics(ctx, app.Store, runID, est)
	}
	return out, nil
}

// RunV3Execute continues after a successful pre-flight (metrics + agent steps).
func RunV3Execute(ctx context.Context, app App, resolved intent.ResolvedIntent, plan intent.ExecutionPlan, opts V3Options, pre V3Result) (V3Result, error) {
	out := pre
	if app.Executor == nil {
		return out, fmt.Errorf("pipeline v3: executor nil")
	}
	runID := out.MetricsRunID
	if app.Store != nil {
		_ = telemetry.SaveRunStarted(ctx, app.Store, runID, resolved.Feature, resolved.TaskID, time.Now().UTC())
	}

	workOpts := intent.WorkOptions{
		PlanOnly:    opts.PlanOnly,
		Yes:         opts.Yes,
		DryRun:      opts.DryRun,
		Interactive: opts.Interactive,
		MaxTasks:    opts.MaxTasks,
		StopAfter:   opts.StopAfter,
		NoReview:    opts.NoReview,
		Agent:       opts.Agent,
		Reviewer:    opts.Reviewer,
	}
	start := time.Now()
	snap, err := intent.BuildSnapshot(app.RepoRoot, app.Config, app.Store)
	if err != nil {
		return out, err
	}
	execRes, err := app.Executor.Execute(ctx, plan, snap, workOpts)
	out.Exec = execRes
	if err != nil {
		return out, err
	}
	if app.Store != nil {
		fin := telemetry.RunMetric{
			RunID:                 runID,
			Feature:               resolved.Feature,
			TaskID:                resolved.TaskID,
			StartedAt:             start,
			FinishedAt:            time.Now().UTC(),
			EstimatedInputTokens:  out.Estimate.TotalInputTokens,
			EstimatedOutputTokens: out.Estimate.TotalOutputTokens,
			EstimatedCostCents:    out.Estimate.EstimatedCost.Cents,
			ActualDuration:        time.Since(start),
			Status:                "done",
		}
		_ = telemetry.SaveRunFinished(ctx, app.Store, fin)
	}
	return out, nil
}

func savePlannedStepMetrics(ctx context.Context, st *sqlite.Store, runID string, est cost.ExecutionEstimate) error {
	if st == nil {
		return nil
	}
	for _, step := range est.PlannedSteps {
		m := telemetry.StepMetric{
			ID:                    runID + ":" + step.Name,
			RunID:                 runID,
			StepName:              step.Name,
			Agent:                 step.Agent,
			Model:                 step.Model,
			Local:                 step.Local,
			EstimatedInputTokens:  step.InputTokens,
			EstimatedOutputTokens: step.OutputTokens,
			EstimatedCostCents:    step.EstimatedCost.Cents,
			EstimatedDuration:     step.EstimatedDuration,
			Status:                "planned",
		}
		if err := st.SaveStepMetric(ctx, m); err != nil {
			return err
		}
	}
	return nil
}
