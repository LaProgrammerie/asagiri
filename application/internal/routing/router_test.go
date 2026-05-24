package routing

import (
	"testing"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/config"
)

func TestRoutePreferLocal(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {PreferLocalFor: []string{"enrich"}},
	}
	cfg.Work.DefaultEnricher = "ollama"
	d := Route(cfg, "enrich", true, false, false)
	if !d.Local {
		t.Fatalf("expected local, got %+v", d)
	}
	if d.Reason != "prefer_local" {
		t.Fatalf("reason: %q", d.Reason)
	}
}

func TestRouteNoCloud(t *testing.T) {
	cfg := config.NewTestConfig("t")
	d := Route(cfg, "dev", false, true, false)
	if !d.Local || d.Reason != "no_cloud" {
		t.Fatalf("got %+v", d)
	}
}

func TestRouteCloudHeavy(t *testing.T) {
	cfg := config.NewTestConfig("t")
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {UseCloudHeavyFor: []string{"review"}},
	}
	d := Route(cfg, "review", false, false, true)
	if d.Reason != "cloud_heavy" {
		t.Fatalf("reason: %q got %+v", d.Reason, d)
	}
}
