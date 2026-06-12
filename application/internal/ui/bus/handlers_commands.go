package bus

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/extractors"
	_ "github.com/LaProgrammerie/asagiri/application/internal/knowledge/sqlite"
	"github.com/LaProgrammerie/asagiri/application/internal/pipeline"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
	"github.com/LaProgrammerie/asagiri/application/internal/replay"
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
	defer func() { _ = store.Close() }()

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
		RepoRoot: deps.RepoRoot,
		Store:    store,
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
	out, err := pipeline.RunV3Execute(ctx, app, resolved, plan, opts, pre)
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

func dispatchBuildKnowledgeGraph(ctx context.Context, deps Deps, cmd BuildKnowledgeGraphCommand) (CommandResult, error) {
	cfg, err := resolveConfig(deps)
	if err != nil {
		return CommandResult{}, err
	}
	req := knowledge.BuildRequestFromConfig(deps.RepoRoot, cfg)
	if cmd.Incremental {
		req.Incremental = true
	}
	if deps.DryRun {
		return CommandResult{
			Accepted:      true,
			Message:       "knowledge build dry-run planned",
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	}
	result, err := knowledge.DefaultBuilder().Build(ctx, req)
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("knowledge graph built: %d nodes, %d edges", result.Nodes, result.Edges),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchReplayRun(ctx context.Context, deps Deps, cmd ReplayRunCommand) (CommandResult, error) {
	replayID := strings.TrimSpace(cmd.RunID)
	if replayID == "" {
		return CommandResult{}, fmt.Errorf("ui replay run: replay id required")
	}
	cfg, err := resolveConfig(deps)
	if err != nil {
		return CommandResult{}, err
	}
	policies := replay.DefaultCapturePolicies(cfg)
	mgr := replay.DefaultManager(deps.RepoRoot, policies)
	result, err := mgr.Run(ctx, replay.ReplayRunRequest{
		RepoRoot:   deps.RepoRoot,
		ReplayID:   replayID,
		DryRun:     deps.DryRun,
		Offline:    cmd.Offline,
		Simulation: cmd.Simulation,
	})
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("replay %s finished: mode=%s", replayID, result.Mode),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchGraphRollback(ctx context.Context, deps Deps, cmd GraphRollbackCommand) (CommandResult, error) {
	graphID := strings.TrimSpace(cmd.GraphID)
	if graphID == "" {
		return CommandResult{}, fmt.Errorf("ui graph rollback: graph id required")
	}
	result, err := executiongraph.ExecuteGraphRollback(ctx, deps.RepoRoot, graphID, deps.DryRun)
	if err != nil {
		return CommandResult{}, err
	}
	msg := fmt.Sprintf("graph %s rolled back (%d nodes)", result.GraphID, result.NodesRolledBack)
	if deps.DryRun {
		msg = fmt.Sprintf("graph %s rollback dry-run (%d nodes)", result.GraphID, result.NodesRolledBack)
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchExportEvents(ctx context.Context, deps Deps, cmd ExportEventsCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	store, err := deps.RuntimeOpen(deps.RepoRoot)
	if err != nil {
		return CommandResult{}, err
	}
	defer func() { _ = store.Close() }()

	rows, err := store.ListEvents(500)
	if err != nil {
		return CommandResult{}, err
	}
	filtered := make([]EventSummary, 0, len(rows))
	for _, row := range rows {
		ev := mapRuntimeEvent(row)
		if !eventMatchesCategory(ev, cmd.TypeFilter) {
			continue
		}
		if !eventMatchesSearchText(ev, cmd.Search) {
			continue
		}
		filtered = append(filtered, ev)
	}

	outPath := strings.TrimSpace(cmd.OutputPath)
	if outPath == "" {
		dir := filepath.Join(deps.RepoRoot, ".asagiri", "exports")
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return CommandResult{}, err
		}
		outPath = filepath.Join(dir, "events-"+time.Now().UTC().Format("20060102-150405")+".json")
	} else if !filepath.IsAbs(outPath) {
		outPath = filepath.Join(deps.RepoRoot, outPath)
	}
	if deps.DryRun {
		return CommandResult{
			Accepted:      true,
			Message:       fmt.Sprintf("events export dry-run: %d rows -> %s", len(filtered), outPath),
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	}
	body, err := json.MarshalIndent(filtered, "", "  ")
	if err != nil {
		return CommandResult{}, err
	}
	if err := os.WriteFile(outPath, body, 0o644); err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("exported %d events to %s", len(filtered), outPath),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchExportGraph(ctx context.Context, deps Deps, cmd ExportGraphCommand) (CommandResult, error) {
	if err := ctx.Err(); err != nil {
		return CommandResult{}, err
	}
	graphID := strings.TrimSpace(cmd.GraphID)
	if graphID == "" {
		return CommandResult{}, fmt.Errorf("ui export graph: graph id required")
	}
	repo := executiongraph.NewRepository(deps.RepoRoot)
	graph, err := repo.Load(graphID)
	if err != nil {
		return CommandResult{}, err
	}
	format := executiongraph.RenderFormat(strings.ToLower(strings.TrimSpace(cmd.Format)))
	if format == "" {
		format = executiongraph.RenderFormatMermaid
	}
	body, err := executiongraph.Render(graph, format)
	if err != nil {
		return CommandResult{}, err
	}
	dir := filepath.Join(deps.RepoRoot, ".asagiri", "exports")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return CommandResult{}, err
	}
	ext := string(format)
	if ext == "markdown" {
		ext = "md"
	}
	outPath := filepath.Join(dir, graphID+"."+ext)
	if deps.DryRun {
		return CommandResult{
			Accepted:      true,
			Message:       fmt.Sprintf("graph export dry-run: %s -> %s", graphID, outPath),
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	}
	if err := os.WriteFile(outPath, []byte(body), 0o644); err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("graph exported to %s", outPath),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchGraphResume(ctx context.Context, deps Deps, cmd GraphResumeCommand) (CommandResult, error) {
	graphID := strings.TrimSpace(cmd.GraphID)
	if graphID == "" {
		return CommandResult{}, fmt.Errorf("ui graph resume: graph id required")
	}
	if deps.DryRun {
		return CommandResult{
			Accepted:      true,
			Message:       fmt.Sprintf("graph %s resume dry-run planned", graphID),
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	}
	runner := executiongraph.NewRunner(deps.RepoRoot)
	result, err := runner.Resume(ctx, graphID, executiongraph.RunOptions{DryRun: deps.DryRun})
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("graph %s resumed: status=%s", result.GraphID, result.Status),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchAnalyzeKnowledgeImpact(ctx context.Context, deps Deps, cmd AnalyzeKnowledgeImpactCommand) (CommandResult, error) {
	store, err := knowledge.OpenStoreIfExists(deps.RepoRoot)
	if err != nil {
		return CommandResult{}, fmt.Errorf("ui impact analyze: knowledge store unavailable")
	}
	defer func() { _ = store.Close() }()
	analyzer := knowledge.NewImpactAnalyzer(store)
	result, err := analyzer.Analyze(ctx, knowledge.ImpactRequest{
		File:   strings.TrimSpace(cmd.File),
		Flow:   strings.TrimSpace(cmd.Flow),
		Action: strings.TrimSpace(cmd.Action),
	})
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("impact analyzed: %d flows, %d apis, risk=%s", len(result.ImpactedFlows), len(result.ImpactedAPIs), result.Risk),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchBuildKnowledgeContext(ctx context.Context, deps Deps, cmd BuildKnowledgeContextCommand) (CommandResult, error) {
	nodeID := strings.TrimSpace(cmd.NodeID)
	if nodeID == "" {
		return CommandResult{}, fmt.Errorf("ui build knowledge context: node id required")
	}
	store, err := knowledge.OpenStoreIfExists(deps.RepoRoot)
	if err != nil {
		return CommandResult{}, fmt.Errorf("ui build knowledge context: knowledge store unavailable")
	}
	defer func() { _ = store.Close() }()
	querier := knowledge.NewQuerier(store)
	result, err := querier.Query(ctx, knowledge.GraphQuery{
		StartID:  nodeID,
		MaxDepth: 2,
		Limit:    24,
	})
	if err != nil {
		return CommandResult{}, err
	}
	return CommandResult{
		Accepted:      true,
		Message:       fmt.Sprintf("context built from %s: %d nodes, %d edges", nodeID, len(result.Nodes), len(result.Edges)),
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchCompareReplay(ctx context.Context, deps Deps, cmd CompareReplayCommand) (CommandResult, error) {
	replayA := strings.TrimSpace(cmd.ReplayA)
	replayB := strings.TrimSpace(cmd.ReplayB)
	if replayA == "" || replayB == "" {
		return CommandResult{}, fmt.Errorf("ui replay compare: both replay ids required")
	}
	cmp, err := replay.NewComparator(deps.RepoRoot).Compare(ctx, replayA, replayB)
	if err != nil {
		return CommandResult{}, err
	}
	msg := fmt.Sprintf("compared %s vs %s: %d differences", replayA, replayB, len(cmp.Differences))
	if len(cmp.Differences) > 0 {
		msg += " (" + cmp.Differences[0] + ")"
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchExplainReplayDivergence(ctx context.Context, deps Deps, cmd ExplainReplayDivergenceCommand) (CommandResult, error) {
	replayA := strings.TrimSpace(cmd.ReplayA)
	replayB := strings.TrimSpace(cmd.ReplayB)
	if replayA == "" || replayB == "" {
		return CommandResult{}, fmt.Errorf("ui replay explain: both replay ids required")
	}
	cmp, err := replay.NewComparator(deps.RepoRoot).Compare(ctx, replayA, replayB)
	if err != nil {
		return CommandResult{}, err
	}
	lines := cmp.Differences
	if len(lines) == 0 {
		lines = replay.ExplainDivergences(cmp.Divergences)
	}
	msg := fmt.Sprintf("divergence explained for %s vs %s", replayA, replayB)
	if len(lines) > 0 {
		msg += ": " + lines[0]
	} else {
		msg += ": no divergence detected"
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchPrototypeCreate(_ context.Context, deps Deps, cmd PrototypeCreateCommand) (CommandResult, error) {
	intent := strings.TrimSpace(cmd.Intent)
	if intent == "" {
		return CommandResult{}, fmt.Errorf("ui prototype create: intent required")
	}
	svc := product.NewService(deps.RepoRoot)
	name, err := svc.CreatePrototype(product.CreatePrototypeOptions{
		Intent:  intent,
		Product: cmd.Product,
		Stack:   defaultString(cmd.Stack, "react"),
		Style:   defaultString(cmd.Style, "minimal"),
		DryRun:  deps.DryRun,
	})
	if err != nil {
		return CommandResult{}, err
	}
	msg := fmt.Sprintf("prototype created: %s", name)
	if deps.DryRun {
		msg = fmt.Sprintf("prototype create dry-run: %s", name)
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchFlowsExtract(_ context.Context, deps Deps, cmd FlowsExtractCommand) (CommandResult, error) {
	productID := strings.TrimSpace(cmd.Product)
	if productID == "" {
		return CommandResult{}, fmt.Errorf("ui flows extract: product required")
	}
	svc := product.NewService(deps.RepoRoot)
	if err := svc.ExtractFlows(productID, deps.DryRun); err != nil {
		return CommandResult{}, err
	}
	msg := fmt.Sprintf("flows extracted: %s", product.Slug(productID))
	if deps.DryRun {
		msg = fmt.Sprintf("flows extract dry-run: %s", product.Slug(productID))
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchContractsExtract(_ context.Context, deps Deps, cmd ContractsExtractCommand) (CommandResult, error) {
	productID := strings.TrimSpace(cmd.Product)
	if productID == "" {
		return CommandResult{}, fmt.Errorf("ui contracts extract: product required")
	}
	svc := product.NewService(deps.RepoRoot)
	if err := svc.ExtractContracts(productID, deps.DryRun); err != nil {
		return CommandResult{}, err
	}
	msg := fmt.Sprintf("contracts extracted: %s", product.Slug(productID))
	if deps.DryRun {
		msg = fmt.Sprintf("contracts extract dry-run: %s", product.Slug(productID))
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func dispatchSpecGenerateFromProduct(_ context.Context, deps Deps, cmd SpecGenerateFromProductCommand) (CommandResult, error) {
	productID := strings.TrimSpace(cmd.Product)
	if productID == "" {
		return CommandResult{}, fmt.Errorf("ui spec generate-from-product: product required")
	}
	svc := product.NewService(deps.RepoRoot)
	if err := svc.GenerateSpecFromProduct(productID, deps.DryRun); err != nil {
		return CommandResult{}, err
	}
	msg := fmt.Sprintf("spec generated from product: %s", product.Slug(productID))
	if deps.DryRun {
		msg = fmt.Sprintf("spec generate-from-product dry-run: %s", product.Slug(productID))
	}
	return CommandResult{
		Accepted:      true,
		Message:       msg,
		CLIEquivalent: cmd.CLIEquivalent(),
	}, nil
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}
