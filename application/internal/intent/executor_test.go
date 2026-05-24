package intent

import (
	"context"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
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
