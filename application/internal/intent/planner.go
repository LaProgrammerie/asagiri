package intent

import (
	"context"
	"fmt"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

// HighLevelPlanner builds primitive step plans (specv2 §6).
type HighLevelPlanner interface {
	BuildPlan(ctx context.Context, intent ResolvedIntent, snap StateSnapshot, cfg *config.Config, opts WorkOptions) (ExecutionPlan, error)
}

// DefaultPlanner implements HighLevelPlanner.
type DefaultPlanner struct{}

// BuildPlan converts resolved intent into primitive CLI steps.
func (p *DefaultPlanner) BuildPlan(ctx context.Context, intent ResolvedIntent, snap StateSnapshot, cfg *config.Config, opts WorkOptions) (ExecutionPlan, error) {
	_ = ctx
	if intent.Feature == "" && intent.Action != IntentSync && intent.Action != IntentImport {
		if snap.ActiveFeature != "" {
			intent.Feature = snap.ActiveFeature
		}
	}

	agent := opts.Agent
	if agent == "" && cfg != nil {
		agent = cfg.Work.DefaultAgent
	}
	if agent == "" {
		agent = "cursor"
	}
	reviewer := opts.Reviewer
	if reviewer == "" && cfg != nil {
		reviewer = cfg.Work.DefaultReviewer
	}
	if reviewer == "" {
		reviewer = "codex"
	}
	enricher := "ollama"
	if cfg != nil && cfg.Work.DefaultEnricher != "" {
		enricher = cfg.Work.DefaultEnricher
	}

	feat := intent.Feature
	fs := featureState(snap, feat)
	taskID := intent.TaskID
	if taskID == "" {
		taskID = fs.NextTaskID
	}

	steps := []PlanStep{}

	add := func(cmd string, args []string, cond, reason string) {
		steps = append(steps, PlanStep{Command: cmd, Args: args, Condition: cond, Reason: reason})
	}

	switch intent.Action {
	case IntentSync, IntentImport:
		if intent.SourceRef != "" {
			add("sync", []string{"notion", "--page", intent.SourceRef}, "source_requires_sync", "import notion page")
		} else if feat != "" {
			add("sync", []string{"notion", "--feature", feat}, "source_requires_sync", "sync feature from notion")
		}
	case IntentStatus:
		add("status", nil, "", "show status")
	case IntentVerify:
		add("verify", withTask(feat, taskID), "implementation_done", "verify task")
	case IntentReview:
		add("review", append(withTask(feat, taskID), "--agent", reviewer), "review_enabled", "review task")
	case IntentFix:
		add("verify", append(withTask(feat, taskID), "--force"), "verification_failed", "re-verify after fix")
		add("dev", append(append(withTask(feat, taskID), "--agent", agent), "--force"), "task_pending_or_enriched", "retry dev")
	default:
		if intent.RequiresSync || fs.SourceType == "notion" {
			add("sync", []string{"notion", "--feature", feat}, "source_requires_sync", "sync before work")
		}
		if !fs.HasLocalSpec {
			add("spec", []string{feat, "--agent", "kiro"}, "no_local_spec", "create local spec")
		}
		if !fs.HasTasks {
			add("plan", []string{feat}, "no_tasks", "generate tasks")
		}
		add("enrich", append(withTask(feat, taskID), "--agent", enricher), "task_not_enriched", "enrich task")
		add("dev", append(withTask(feat, taskID), "--agent", agent), "task_pending_or_enriched", "implement task")
		if cfg == nil || cfg.Work.AutoVerify {
			add("verify", withTask(feat, taskID), "implementation_done", "verify implementation")
		}
		autoReview := cfg != nil && cfg.Work.AutoReview && !opts.NoReview
		if autoReview {
			add("review", append(withTask(feat, taskID), "--agent", reviewer), "review_enabled", "independent review")
		}
	}

	stopAfter := opts.StopAfter
	if stopAfter == "" && cfg != nil {
		stopAfter = cfg.Work.StopAfter
	}
	steps = trimAfter(steps, stopAfter)

	maxTasks := opts.MaxTasks
	if maxTasks == 0 && cfg != nil {
		maxTasks = cfg.Work.MaxTasksPerRun
	}
	if maxTasks == 0 {
		maxTasks = 1
	}
	_ = maxTasks // enforced at execution

	return ExecutionPlan{Intent: intent, Steps: steps}, nil
}

func featureState(snap StateSnapshot, name string) FeatureState {
	for _, f := range snap.Features {
		if f.Name == name {
			return f
		}
	}
	return FeatureState{Name: name}
}

func withTask(feature, taskID string) []string {
	args := []string{feature}
	if taskID != "" {
		args = append(args, "--task", taskID)
	}
	return args
}

func trimAfter(steps []PlanStep, stopAfter string) []PlanStep {
	if stopAfter == "" || stopAfter == "report" {
		return steps
	}
	order := map[string]int{"sync": 0, "spec": 1, "plan": 2, "enrich": 3, "dev": 4, "verify": 5, "review": 6, "report": 7, "pr": 8}
	limit, ok := order[stopAfter]
	if !ok {
		return steps
	}
	out := make([]PlanStep, 0, len(steps))
	for _, s := range steps {
		if rank, ok := order[s.Command]; ok && rank > limit {
			continue
		}
		out = append(out, s)
	}
	return out
}

// EvaluateCondition checks whether a plan step should run (specv2 §6.2).
func EvaluateCondition(cond string, intent ResolvedIntent, fs FeatureState, opts WorkOptions) bool {
	switch cond {
	case "", "always":
		return true
	case "source_requires_sync":
		return intent.RequiresSync || fs.SourceType == "notion"
	case "no_local_spec":
		return !fs.HasLocalSpec
	case "no_tasks":
		return !fs.HasTasks
	case "task_not_enriched":
		return fs.NextTaskStatus == asagiri.StatusPending || fs.NextTaskStatus == asagiri.StatusPlanned
	case "task_pending_or_enriched":
		switch fs.NextTaskStatus {
		case asagiri.StatusPending, asagiri.StatusPlanned, asagiri.StatusEnriched, "":
			return true
		default:
			return fs.NextTaskStatus == asagiri.StatusEnriched
		}
	case "implementation_done":
		return fs.NextTaskStatus == asagiri.StatusImplemented || fs.NextTaskStatus == asagiri.StatusRunning
	case "verification_failed":
		return fs.NextTaskStatus == asagiri.StatusVerifyFailed
	case "review_enabled":
		return !opts.NoReview
	case "requires_human_approval":
		return intent.RequiresHuman
	default:
		return true
	}
}

// RecommendNext computes next primitive for `asa next`.
func RecommendNext(snap StateSnapshot, feature string) (NextRecommendation, error) {
	fs := featureState(snap, feature)
	if feature == "" {
		feature = snap.ActiveFeature
		fs = featureState(snap, feature)
	}
	if feature == "" {
		return NextRecommendation{}, fmt.Errorf("aucune feature active")
	}
	taskID := fs.NextTaskID
	switch fs.NextTaskStatus {
	case asagiri.StatusImplemented, asagiri.StatusRunning:
		cmd := fmt.Sprintf("asa verify %s --task %s", feature, taskID)
		return NextRecommendation{
			Feature: feature, TaskID: taskID, Action: "verify",
			Reason: "implementation completed but validation missing", Primitive: cmd,
		}, nil
	case asagiri.StatusVerified:
		cmd := fmt.Sprintf("asa review %s --task %s --agent codex", feature, taskID)
		return NextRecommendation{
			Feature: feature, TaskID: taskID, Action: "review",
			Reason: "verified but review missing", Primitive: cmd,
		}, nil
	case asagiri.StatusEnriched, asagiri.StatusPending, asagiri.StatusPlanned, "":
		cmd := fmt.Sprintf("asa dev %s --task %s --agent cursor", feature, taskID)
		if taskID == "" {
			cmd = fmt.Sprintf("asa plan %s", feature)
			return NextRecommendation{Feature: feature, Action: "plan", Reason: "no tasks planned", Primitive: cmd}, nil
		}
		return NextRecommendation{
			Feature: feature, TaskID: taskID, Action: "dev",
			Reason: "task ready for implementation", Primitive: cmd,
		}, nil
	case asagiri.StatusVerifyFailed:
		cmd := fmt.Sprintf("asa verify %s --task %s --force", feature, taskID)
		return NextRecommendation{Feature: feature, TaskID: taskID, Action: "verify", Reason: "verification failed", Primitive: cmd}, nil
	default:
		cmd := fmt.Sprintf("asa status")
		return NextRecommendation{Feature: feature, Action: "status", Reason: "inspect current state", Primitive: cmd}, nil
	}
}
