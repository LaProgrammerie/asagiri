package routing

// Feature: audit-coherence-consolidation, Property 8
//
// Property 8 (design P8) — Stratégie par défaut.
// Pour toute entrée où ni noCloud ni preferLocal ne sont demandés et où la classe
// d'étape ne figure pas dans prefer_local_for, Route — lorsqu'il réussit —
// applique la stratégie par défaut et expose la raison "cloud_heavy",
// "cloud_fast" ou "default" selon l'appartenance de la classe aux listes de
// stratégie, en respectant la précédence heavy > fast > default. Ces décisions
// sont des décisions cloud (Local == false).
//
// Validates: Requirements 4.5

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// p8ClassPool is the deterministic pool of (lowercase) step classes used both to
// build the strategy lists and to pick the queried step class.
var p8ClassPool = []string{"spec", "plan", "enrich", "dev", "verify", "review", "implement", "analyze"}

// p8Case is one generated routing input that satisfies the Property 8
// precondition by construction: preferLocal and noCloud are always false and the
// queried step class never appears in prefer_local_for (that list is kept empty
// in build). Only primitive/slice fields are carried so testing/quick prints
// readable counterexamples; the config is rebuilt deterministically from them.
type p8Case struct {
	StepClass  string   // queried step class (from p8ClassPool)
	HeavyFor   []string // strategy.UseCloudHeavyFor
	FastFor    []string // strategy.UseCloudFastFor
	AllowCloud bool     // allowCloud flag (must not change the default-strategy outcome)
}

// Generate builds a deterministic, precondition-satisfying routing input.
func (p8Case) Generate(r *rand.Rand, _ int) reflect.Value {
	c := p8Case{
		StepClass:  p8ClassPool[r.Intn(len(p8ClassPool))],
		HeavyFor:   p8RandSubset(r, p8ClassPool),
		FastFor:    p8RandSubset(r, p8ClassPool),
		AllowCloud: r.Intn(2) == 0,
	}
	return reflect.ValueOf(c)
}

// p8RandSubset returns a deterministic random subset of items.
func p8RandSubset(r *rand.Rand, items []string) []string {
	var out []string
	for _, item := range items {
		if r.Intn(2) == 0 {
			out = append(out, item)
		}
	}
	return out
}

// build materializes a config whose declared cloud Default agent lets every
// default-strategy branch (cloud_heavy / cloud_fast / default) resolve a declared
// backend, so Route succeeds and the property stays non-vacuous. prefer_local_for
// is left empty to guarantee the Property 8 precondition (class ∉ prefer_local_for).
func (c p8Case) build() (cfg *config.Config, stepClass string) {
	cfg = config.NewTestConfig("p8")
	cfg.Agents = map[string]config.Agent{
		"cursor": {Command: "cursor"},                  // cloud (no localhost endpoint)
		"claude": {Command: "claude"},                  // cloud
		"ollama": {Endpoint: "http://localhost:11434"}, // local (fallback only)
	}
	// A declared cloud agent as DefaultAgent: cloudHeavyDeclared returns it for
	// the cloud_heavy branch, and the cloud_fast/default branches use it directly.
	cfg.Work.DefaultAgent = "cursor"
	cfg.Work.DefaultEnricher = "ollama"

	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{
		"cost_aware": {
			PreferLocalFor:   nil, // precondition: class never in prefer_local_for
			UseCloudHeavyFor: c.HeavyFor,
			UseCloudFastFor:  c.FastFor,
		},
	}
	return cfg, c.StepClass
}

// p8Contains mirrors the router's case-insensitive membership test (EqualFold).
func p8Contains(list []string, item string) bool {
	for _, x := range list {
		if strings.EqualFold(x, item) {
			return true
		}
	}
	return false
}

// p8ExpectedReason computes the reason the default strategy must expose for the
// queried class, mirroring Route's precedence heavy > fast > default.
func p8ExpectedReason(heavy, fast []string, cls string) string {
	lc := strings.ToLower(strings.TrimSpace(cls))
	switch {
	case p8Contains(heavy, lc):
		return "cloud_heavy"
	case p8Contains(fast, lc):
		return "cloud_fast"
	default:
		return "default"
	}
}

// TestRouteDefaultStrategyProperty asserts Property 8: with neither noCloud nor
// preferLocal and the class outside prefer_local_for, Route exposes the reason
// dictated by the strategy lists (heavy > fast > default) and yields a cloud
// decision (Local == false).
func TestRouteDefaultStrategyProperty(t *testing.T) {
	property := func(c p8Case) bool {
		cfg, stepClass := c.build()

		// preferLocal and noCloud are always false; allowCloud varies freely.
		d, err := Route(cfg, stepClass, false, false, c.AllowCloud)
		if err != nil {
			// A declared cloud DefaultAgent is always present, so every
			// default-strategy branch must resolve a declared backend.
			t.Errorf("unexpected error (class=%q heavy=%v fast=%v): %v",
				stepClass, c.HeavyFor, c.FastFor, err)
			return false
		}

		want := p8ExpectedReason(c.HeavyFor, c.FastFor, stepClass)
		if d.Reason != want {
			t.Errorf("expected Reason=%q, got %q (class=%q heavy=%v fast=%v)",
				want, d.Reason, stepClass, c.HeavyFor, c.FastFor)
			return false
		}
		if d.Local {
			t.Errorf("expected Local=false for default strategy, got %+v (class=%q)", d, stepClass)
			return false
		}
		return true
	}

	// testing/quick with a fixed seed: deterministic, >= 100 iterations.
	qcfg := &quick.Config{MaxCount: 200, Rand: rand.New(rand.NewSource(8))}
	if err := quick.Check(property, qcfg); err != nil {
		t.Fatalf("Property 8 (stratégie par défaut) failed: %v", err)
	}
}
