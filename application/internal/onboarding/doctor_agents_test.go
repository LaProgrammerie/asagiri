package onboarding_test

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
	"github.com/stretchr/testify/require"
)

func TestReadinessOllamaEndpointWithoutCommandOK(t *testing.T) {
	repo := t.TempDir()
	cfg := config.NewTestConfig("demo")
	cfg.Work.DefaultEnricher = "ollama"
	cfg.Agents["ollama"] = config.Agent{
		Endpoint: "http://127.0.0.1:11434",
		Model:    "qwen2.5-coder:7b",
	}
	report, err := onboarding.AssessReadiness(repo, cfg, false)
	require.NoError(t, err)
	for _, c := range report.Checks {
		if c.ID == "agents.ollama" {
			t.Fatalf("unexpected agents.ollama check: %+v", c)
		}
	}
}
