package cost

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/contextopt"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/investigation"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/intent"
)

func TestBuildEstimateGoldenShape(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Pricing.Currency = "EUR"
	cfg.Pricing.Models["m1"] = config.ModelPricing{InputPer1MTokens: 3, OutputPer1MTokens: 6}
	cfg.Agents["cursor"] = config.Agent{Command: "true", Model: "m1"}
	plan := intent.ExecutionPlan{
		Intent: intent.ResolvedIntent{Feature: "billing", TaskID: "t1", Action: intent.IntentDevelop},
		Steps: []intent.PlanStep{
			{Command: "dev", Args: []string{"--agent", "cursor"}, Reason: "implement"},
		},
	}
	pack := contextopt.ContextPack{TaskObjective: "### Objective\ntest"}
	est, err := BuildEstimate(context.Background(), plan, investigation.InvestigationResult{}, pack, cfg, DefaultDurationModel{Cfg: cfg}, BuildOpts{RunID: "g1"})
	if err != nil {
		t.Fatal(err)
	}
	if est.TotalInputTokens <= 0 {
		t.Fatalf("tokens: %+v", est)
	}
	golden := filepath.Join("testdata", "estimate_shape.json")
	if os.Getenv("UPDATE_GOLDEN") == "1" {
		_ = os.MkdirAll(filepath.Dir(golden), 0o755)
		b, _ := json.MarshalIndent(struct {
			Steps int `json:"planned_steps"`
			In    int `json:"total_input_tokens"`
		}{len(est.PlannedSteps), est.TotalInputTokens}, "", "  ")
		_ = os.WriteFile(golden, b, 0o644)
	}
	b, err := os.ReadFile(golden)
	if err != nil {
		t.Skip("golden file missing; run with UPDATE_GOLDEN=1")
	}
	var want struct {
		Steps int `json:"planned_steps"`
		In    int `json:"total_input_tokens"`
	}
	if err := json.Unmarshal(b, &want); err != nil {
		t.Fatal(err)
	}
	if len(est.PlannedSteps) != want.Steps {
		t.Fatalf("steps: got %d want %d", len(est.PlannedSteps), want.Steps)
	}
}
