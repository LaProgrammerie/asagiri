package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestExecutorDryRunSkipsWorkflow(t *testing.T) {
	cfg := config.NewTestConfig("t")
	ex := &Executor{Config: cfg}
	plan := ExecutionPlan{
		Intent: ResolvedIntent{Feature: "f", Action: IntentDevelop},
		Steps: []PlanStep{
			{Command: "plan", Args: []string{"f"}},
			{Command: "dev", Args: []string{"--task", "t1"}},
		},
	}
	res, err := ex.Execute(context.Background(), plan, StateSnapshot{}, WorkOptions{DryRun: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Executed) != 2 {
		t.Fatalf("executed: %v", res.Executed)
	}
	for _, line := range res.Executed {
		if line[:6] != "[dry] " {
			t.Fatalf("expected dry prefix: %q", line)
		}
	}
}

func TestExecutorSkipsCondition(t *testing.T) {
	cfg := config.NewTestConfig("t")
	ex := &Executor{Config: cfg}
	plan := ExecutionPlan{
		Intent: ResolvedIntent{Feature: "f", Action: IntentDevelop},
		Steps: []PlanStep{
			{Command: "review", Condition: "review_enabled", Args: []string{"f"}},
		},
	}
	res, err := ex.Execute(context.Background(), plan, StateSnapshot{}, WorkOptions{DryRun: true, NoReview: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Executed) != 0 || len(res.Skipped) != 1 {
		t.Fatalf("executed=%d skipped=%v", len(res.Executed), res.Skipped)
	}
}

func TestExecutorPlanOnlyPrefix(t *testing.T) {
	cfg := config.NewTestConfig("t")
	ex := &Executor{Config: cfg}
	plan := ExecutionPlan{
		Intent: ResolvedIntent{Feature: "f"},
		Steps:  []PlanStep{{Command: "dev", Args: []string{"f"}}},
	}
	res, err := ex.Execute(context.Background(), plan, StateSnapshot{}, WorkOptions{PlanOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Executed) != 1 || res.Executed[0][:7] != "[plan] " {
		t.Fatalf("expected [plan] prefix, got: %v", res.Executed)
	}
}

func TestExecutorMaxTasksLimitsDevSteps(t *testing.T) {
	cfg := config.NewTestConfig("t")
	ex := &Executor{Config: cfg}
	plan := ExecutionPlan{
		Intent: ResolvedIntent{Feature: "f"},
		Steps: []PlanStep{
			{Command: "dev", Args: []string{"f", "--task", "t1"}},
			{Command: "dev", Args: []string{"f", "--task", "t2"}},
			{Command: "dev", Args: []string{"f", "--task", "t3"}},
		},
	}
	// DryRun simulates real execution order including maxTasks gate.
	res, err := ex.Execute(context.Background(), plan, StateSnapshot{}, WorkOptions{DryRun: true, MaxTasks: 1})
	if err != nil {
		t.Fatal(err)
	}
	if len(res.Executed) != 1 {
		t.Fatalf("expected 1 dev executed, got %v", res.Executed)
	}
	if len(res.Skipped) != 2 {
		t.Fatalf("expected 2 dev skipped by maxTasks, got %v", res.Skipped)
	}
}
