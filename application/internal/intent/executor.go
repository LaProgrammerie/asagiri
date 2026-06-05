package intent

import (
	"context"
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/workflow"
)

// Executor runs high-level plans via workflow primitives.
type Executor struct {
	Workflow *workflow.Service
	Config   *config.Config
	SyncFn   func(ctx context.Context, args []string, force bool) error
}

// ExecuteResult captures executed primitives.
type ExecuteResult struct {
	Intent     ResolvedIntent
	Plan       ExecutionPlan
	Executed   []string
	LastRunID  string
	Skipped    []string
}

// Execute runs applicable plan steps.
func (e *Executor) Execute(ctx context.Context, plan ExecutionPlan, snap StateSnapshot, opts WorkOptions) (ExecuteResult, error) {
	res := ExecuteResult{Intent: plan.Intent, Plan: plan}
	fs := featureState(snap, plan.Intent.Feature)

	for _, step := range plan.Steps {
		if step.Condition != "" && !EvaluateCondition(step.Condition, plan.Intent, fs, opts) {
			res.Skipped = append(res.Skipped, step.Command+": "+step.Condition)
			continue
		}
		if opts.PlanOnly || opts.DryRun {
			line := formatPrimitive(step)
			res.Executed = append(res.Executed, "[dry] "+line)
			continue
		}
		runID, err := e.runStep(ctx, step, plan.Intent, opts)
		if err != nil {
			return res, fmt.Errorf("%s: %w", step.Command, err)
		}
		if runID != "" {
			res.LastRunID = runID
		}
		res.Executed = append(res.Executed, formatPrimitive(step))
	}
	return res, nil
}

func (e *Executor) runStep(ctx context.Context, step PlanStep, intent ResolvedIntent, opts WorkOptions) (string, error) {
	feature := intent.Feature
	taskID := intent.TaskID
	if taskID == "" {
		taskID = taskFlag(step.Args)
	}
	agent := opts.Agent
	if agent == "" {
		agent = e.Config.Work.DefaultAgent
	}
	reviewer := opts.Reviewer
	if reviewer == "" {
		reviewer = e.Config.Work.DefaultReviewer
	}
	specAgent := "kiro"
	if e.Config != nil && e.Config.Work.DefaultSpecAgent != "" {
		specAgent = e.Config.Work.DefaultSpecAgent
	}
	enricher := "ollama"
	if e.Config != nil && e.Config.Work.DefaultEnricher != "" {
		enricher = e.Config.Work.DefaultEnricher
	}
	force := hasFlag(step.Args, "--force")

	switch step.Command {
	case "sync":
		if e.SyncFn != nil {
			return "", e.SyncFn(ctx, step.Args, force)
		}
		return "", nil
	case "spec":
		return e.Workflow.SpecFeature(ctx, feature, agentOr(step.Args, specAgent))
	case "plan":
		runID, _, err := e.Workflow.PlanFeature(feature)
		return runID, err
	case "enrich":
		return e.Workflow.EnrichFeature(ctx, feature, taskID, agentOr(step.Args, enricher), force)
	case "dev":
		return e.Workflow.DevFeature(ctx, feature, taskID, agentOr(step.Args, agent), force)
	case "verify":
		return e.Workflow.VerifyFeature(ctx, feature, taskID, force)
	case "review":
		return e.Workflow.ReviewFeature(ctx, feature, taskID, agentOr(step.Args, reviewer), force)
	case "status":
		_, err := e.Workflow.Status(20)
		return "", err
	default:
		return "", fmt.Errorf("commande primitive inconnue: %s", step.Command)
	}
}

func formatPrimitive(step PlanStep) string {
	parts := append([]string{"asa", step.Command}, step.Args...)
	return strings.Join(parts, " ")
}

func agentOr(args []string, def string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--agent" {
			return args[i+1]
		}
	}
	return def
}

func taskFlag(args []string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == "--task" {
			return args[i+1]
		}
	}
	return ""
}

func hasFlag(args []string, flag string) bool {
	for _, a := range args {
		if a == flag {
			return true
		}
	}
	return false
}
