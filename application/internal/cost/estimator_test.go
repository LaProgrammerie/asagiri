package cost

import (
	"context"
	"strings"
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/contextopt"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/investigation"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/intent"
)

func TestBuildEstimateLocalAgent(t *testing.T) {
	cfg := config.NewTestConfig("x")
	cfg.Agents["ollama"] = config.Agent{Endpoint: "http://localhost:11434", Model: "qwen"}
	cfg.Pricing.Models["qwen"] = config.ModelPricing{InputPer1MTokens: 0, OutputPer1MTokens: 0}
	pl := intent.ExecutionPlan{}
	pl.Steps = []intent.PlanStep{{Command: "dev", Args: []string{"--agent", "ollama"}, Reason: "dev"}}
	inv := investigation.InvestigationResult{}
	pack := contextopt.ContextPack{TaskObjective: strings.Repeat("x", 400)}
	dm := DefaultDurationModel{Cfg: cfg}
	est, err := BuildEstimate(context.Background(), pl, inv, pack, cfg, dm, BuildOpts{RunID: "r1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(est.PlannedSteps) != 1 {
		t.Fatalf("steps %d", len(est.PlannedSteps))
	}
	if !est.PlannedSteps[0].Local {
		t.Fatal("expected local ollama step")
	}
}
