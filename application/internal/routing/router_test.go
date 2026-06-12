package routing

import (
	"testing"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

func TestRoutePreferLocal(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {PreferLocalFor: []string{"enrich"}},
	}
	cfg.Work.DefaultEnricher = "ollama"
	cfg.Agents["ollama"] = config.Agent{Endpoint: "http://localhost:11434"}
	d, err := Route(cfg, "enrich", true, false, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.Local {
		t.Fatalf("expected local, got %+v", d)
	}
	if d.Reason != "prefer_local" {
		t.Fatalf("reason: %q", d.Reason)
	}
	if d.Agent != "ollama" {
		t.Fatalf("agent: %q", d.Agent)
	}
}

func TestRouteNoCloud(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Work.DefaultEnricher = "ollama"
	cfg.Agents["ollama"] = config.Agent{Endpoint: "http://localhost:11434"}
	d, err := Route(cfg, "dev", false, true, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !d.Local || d.Reason != "no_cloud" {
		t.Fatalf("got %+v", d)
	}
}

func TestRouteCloudHeavy(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {UseCloudHeavyFor: []string{"review"}},
	}
	cfg.Work.DefaultAgent = "cursor"
	cfg.Agents["cursor"] = config.Agent{Command: "cursor"}
	d, err := Route(cfg, "review", false, false, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if d.Reason != "cloud_heavy" {
		t.Fatalf("reason: %q got %+v", d.Reason, d)
	}
	if d.Agent != "cursor" {
		t.Fatalf("agent: %q", d.Agent)
	}
}
