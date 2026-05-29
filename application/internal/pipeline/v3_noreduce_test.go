package pipeline

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
)

func TestRunV3PreFlightNoContextReduce(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "go.mod"), "module example.com/t\n\ngo 1.25\n")
	writeFile(t, filepath.Join(dir, "sample.go"), "package main\nfunc main() {}\n")
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
	cfg.Agents = map[string]config.Agent{"cursor": {Command: "true", Model: "gpt-test"}}
	resolved := intent.ResolvedIntent{Feature: "sample", Action: intent.IntentDevelop, Reason: "test"}
	plan := intent.ExecutionPlan{
		Intent: resolved,
		Steps:  []intent.PlanStep{{Command: "plan", Reason: "normalize"}},
	}
	app := App{RepoRoot: dir, Config: cfg}
	res, err := RunV3PreFlight(context.Background(), app, resolved, plan, V3Options{NoContextReduce: true})
	if err != nil {
		t.Fatal(err)
	}
	if res.Optimize.OriginalTokens <= 0 {
		t.Fatalf("expected original tokens, got %+v", res.Optimize)
	}
}
