package bus

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/pipeline"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/trust"
	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
)

func dispatchStartWork(ctx context.Context, deps Deps, cmd StartWorkCommand) (CommandResult, error) {
	intentInput := strings.TrimSpace(cmd.Intent)
	if intentInput == "" {
		return CommandResult{}, fmt.Errorf("ui start work: intent required")
	}
	cfg, err := resolveConfig(deps)
	if err != nil {
		return CommandResult{}, err
	}
	store, err := openStateStore(deps, cfg)
	if err != nil {
		return CommandResult{}, err
	}
	defer store.Close()

	snapshot, err := intent.BuildSnapshot(deps.RepoRoot, cfg, store)
	if err != nil {
		return CommandResult{}, err
	}
	resolver := intent.NewHybridResolver()
	resolved, err := resolver.Resolve(ctx, intent.IntentInput{
		RawInstruction: intentInput,
		WorkingDir:     deps.RepoRoot,
		Config:         cfg,
		StateSnapshot:  snapshot,
		Interactive:    false,
	})
	if err != nil {
		return CommandResult{}, err
	}

	planner := &intent.DefaultPlanner{}
	plan, err := planner.BuildPlan(ctx, resolved, snapshot, cfg, intent.WorkOptions{
		DryRun:      deps.DryRun,
		Yes:         true,
		Interactive: false,
	})
	if err != nil {
		return CommandResult{}, err
	}

	executor := &intent.Executor{
		Workflow: workflow.NewService(deps.RepoRoot, cfg, store, deps.DryRun),
		Config:   cfg,
	}
	opts := pipeline.V3Options{
		Interactive:         false,
		Yes:                 true,
		DryRun:              deps.DryRun,
		AllowOverBudget:     true,
		UserConfirmedBudget: true,
	}
	app := pipeline.App{
		RepoRoot: deps.RepoRoot,
		Config:   cfg,
		Store:    store,
		Executor: executor,
	}
	pre, err := pipeline.RunV3PreFlight(ctx, app, resolved, plan, opts)
	if err != nil {
		return CommandResult{}, err
	}
	out := pre
	out, err = pipeline.RunV3Execute(ctx, app, resolved, plan, opts, pre)
	if err != nil {
		return CommandResult{}, err
	}
	message := "work dispatched"
	if deps.DryRun {
		message = "work dry-run planned"
	}
	if strings.TrimSpace(out.Exec.LastRunID) != "" {
		message = fmt.Sprintf("%s (run %s)", message, out.Exec.LastRunID)
	}
	return CommandResult{
		Accepted:      true,
		Message:       message,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchRunInvestigation(ctx context.Context, deps Deps, cmd RunInvestigationCommand) (CommandResult, error) {
	symptom := strings.TrimSpace(cmd.Symptom)
	if symptom == "" {
		return CommandResult{}, fmt.Errorf("ui investigation: symptom required")
	}
	cfg, err := resolveConfig(deps)
	if err != nil {
		return CommandResult{}, err
	}
	res, err := investigation.RunInvestigation(ctx, investigation.Request{
		Symptom:      symptom,
		Feature:      symptom,
		Depth:        investigation.DepthStandard,
		Output:       "markdown",
		EstimateOnly: deps.DryRun,
		RepoRoot:     deps.RepoRoot,
	}, cfg)
	if err != nil {
		return CommandResult{}, err
	}
	if !deps.DryRun {
		_ = investigation.FeedMemory(deps.RepoRoot, res)
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("investigation completed: %s", res.ID),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchVerifyTrust(ctx context.Context, deps Deps, cmd VerifyTrustCommand) (CommandResult, error) {
	target := strings.TrimSpace(cmd.Target)
	if target == "" {
		return CommandResult{}, fmt.Errorf("ui verify trust: target required")
	}
	cfg, err := resolveConfig(deps)
	if err != nil {
		return CommandResult{}, err
	}
	eng := trust.NewEngine(deps.RepoRoot)
	eng.Gates = trust.NewGateEvaluator(&cfg.Verification)
	eng.Config = cfg

	result, err := eng.Verify(ctx, trust.VerificationRequest{
		Flow:   target,
		Strict: false,
	})
	if err != nil {
		return CommandResult{}, err
	}
	accepted := result.Report.Gate.Status != trust.GateStatusBlocked
	return CommandResult{
		Accepted:      accepted,
		Message:       fmt.Sprintf("trust verification %s (%s)", result.TrustID, result.Report.Gate.Status),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func resolveConfig(deps Deps) (*config.Config, error) {
	if deps.Config != nil {
		return deps.Config, nil
	}
	cfgPath := config.ConfigPath(deps.RepoRoot)
	return config.Load(cfgPath, deps.RepoRoot)
}

func openStateStore(deps Deps, cfg *config.Config) (*sqlite.Store, error) {
	path := deps.StateDBPath
	if strings.TrimSpace(path) == "" {
		path = cfg.StateDBPath(deps.RepoRoot)
	}
	store, err := sqlite.Open(path)
	if err != nil {
		return nil, err
	}
	if err := store.Migrate(); err != nil {
		_ = store.Close()
		return nil, err
	}
	return store, nil
}

func sanitizeIntentAsFallbackFeature(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "ui-work"
	}
	value = strings.ReplaceAll(value, " ", "-")
	value = strings.ReplaceAll(value, "/", "-")
	value = strings.Trim(value, "-")
	if value == "" {
		return "ui-work"
	}
	return value
}

func fallbackRunID(prefix string) string {
	return fmt.Sprintf("%s-%s", prefix, time.Now().UTC().Format("20060102-150405"))
}
