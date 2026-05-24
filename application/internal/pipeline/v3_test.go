package pipeline

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
)

func TestRunV3PipelineEstimateOnly(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/t\n\ngo 1.25\n")
	writeFile(t, filepath.Join(dir, ".asagiri", "config.yaml"), `
project:
  name: t
state:
  backend: sqlite
  path: .asagiri/state.sqlite
pricing:
  currency: EUR
  models:
    gpt-test:
      input_per_1m_tokens: 1
      output_per_1m_tokens: 2
`)
	cfg, err := config.Load(filepath.Join(dir, ".asagiri", "config.yaml"), dir)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Pricing.Models == nil {
		cfg.Pricing.Models = map[string]config.ModelPricing{}
	}
	cfg.Pricing.Models["gpt-test"] = config.ModelPricing{InputPer1MTokens: 1, OutputPer1MTokens: 2}
	if cfg.Agents == nil {
		cfg.Agents = map[string]config.Agent{}
	}
	cfg.Agents["cursor"] = config.Agent{Command: "true", Model: "gpt-test"}
	resolved := intent.ResolvedIntent{
		Feature: "feat",
		Action:  intent.IntentDevelop,
		Reason:  "test",
	}
	plan := intent.ExecutionPlan{
		Intent: resolved,
		Steps: []intent.PlanStep{
			{Command: "plan", Reason: "normalize tasks"},
			{Command: "dev", Args: []string{"--agent", "cursor"}, Reason: "implement"},
		},
	}
	app := App{RepoRoot: dir, Config: cfg}
	res, err := RunV3Pipeline(context.Background(), app, resolved, plan, V3Options{EstimateOnly: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Estimate.TotalInputTokens <= 0 {
		t.Fatalf("expected input tokens, got %+v", res.Estimate)
	}
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
