package routing

import (
	"math/rand"
	"reflect"
	"strings"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Feature: audit-coherence-consolidation, Property 6
//
// Property 6 (design P6) — Règle prefer_local.
// Pour toute entrée où preferLocal est demandé (ou la classe d'étape figure dans
// prefer_local_for) et où noCloud n'est pas demandé, Route — lorsqu'il réussit —
// produit une décision avec Local = true et la raison "prefer_local".
//
// Validates: Requirements 4.3

// preferLocalStepClasses is the deterministic pool of step classes used by the
// Property 6 generator.
var preferLocalStepClasses = []string{"spec", "plan", "enrich", "dev", "verify", "review"}

// preferLocalAgentNames lists candidate local Agent_Backend names. They are
// declared with a localhost endpoint so isLocalAgent treats them as local and
// Route can resolve a declared local backend (instead of returning
// ErrNoDeclaredBackend) when the prefer_local branch fires.
var preferLocalAgentNames = []string{"llamacpp", "localai", "ollama"}

// preferLocalCase is one generated routing input that satisfies the Property 6
// precondition by construction: noCloud is always false and either preferLocal
// is true or the step class belongs to prefer_local_for (sometimes both). Only
// primitive fields are carried so quick.Check prints readable counterexamples;
// the config is rebuilt deterministically from these fields.
type preferLocalCase struct {
	StepClassIdx int  // index into preferLocalStepClasses
	Mode         int  // 0: preferLocal only, 1: class∈prefer_local_for only, 2: both
	NumAgents    int  // number of declared local agents (1..len(preferLocalAgentNames))
	SetEnricher  bool // pin Work.DefaultEnricher to a declared local agent
	EnricherIdx  int  // which declared local agent to pin
	UpperClass   bool // store the class in prefer_local_for uppercased (EqualFold coverage)
	AllowCloud   bool // allowCloud flag (must never change the prefer_local outcome)
}

// Generate builds a deterministic, precondition-satisfying routing input.
func (preferLocalCase) Generate(rnd *rand.Rand, _ int) reflect.Value {
	c := preferLocalCase{
		StepClassIdx: rnd.Intn(len(preferLocalStepClasses)),
		Mode:         rnd.Intn(3),
		NumAgents:    1 + rnd.Intn(len(preferLocalAgentNames)),
		SetEnricher:  rnd.Intn(2) == 0,
		UpperClass:   rnd.Intn(2) == 0,
		AllowCloud:   rnd.Intn(2) == 0,
	}
	c.EnricherIdx = rnd.Intn(c.NumAgents)
	return reflect.ValueOf(c)
}

// build materializes the config and the Route arguments for this case.
func (c preferLocalCase) build() (cfg *config.Config, stepClass string, preferLocal bool) {
	cfg = config.NewTestConfig("t")

	declared := preferLocalAgentNames[:c.NumAgents]
	for _, name := range declared {
		cfg.Agents[name] = config.Agent{Endpoint: "http://localhost:11434"}
	}
	if c.SetEnricher {
		cfg.Work.DefaultEnricher = declared[c.EnricherIdx]
	}

	stepClass = preferLocalStepClasses[c.StepClassIdx]
	preferLocal = c.Mode == 0 || c.Mode == 2
	classInPreferLocal := c.Mode == 1 || c.Mode == 2

	strat := config.RoutingStrategy{
		// Unrelated cloud classes ensure prefer_local wins by precedence, not by
		// the mere absence of competing rules.
		UseCloudHeavyFor: []string{"unrelated-heavy"},
		UseCloudFastFor:  []string{"unrelated-fast"},
	}
	if classInPreferLocal {
		entry := stepClass
		if c.UpperClass {
			entry = strings.ToUpper(stepClass)
		}
		strat.PreferLocalFor = []string{entry}
	}

	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{"cost_aware": strat}
	return cfg, stepClass, preferLocal
}

func TestPreferLocalRuleProperty(t *testing.T) {
	property := func(c preferLocalCase) bool {
		cfg, stepClass, preferLocal := c.build()

		// noCloud is always false for this property; allowCloud varies freely.
		d, err := Route(cfg, stepClass, preferLocal, false, c.AllowCloud)
		if err != nil {
			// The generator always declares a local backend, so the prefer_local
			// branch must resolve a declared agent and succeed. An error here is a
			// genuine defect for this property.
			t.Errorf("unexpected error (class=%q preferLocal=%v mode=%d): %v",
				stepClass, preferLocal, c.Mode, err)
			return false
		}
		if !d.Local {
			t.Errorf("expected Local=true, got %+v (class=%q preferLocal=%v mode=%d)",
				d, stepClass, preferLocal, c.Mode)
			return false
		}
		if d.Reason != "prefer_local" {
			t.Errorf("expected Reason=prefer_local, got %q (%+v class=%q preferLocal=%v mode=%d)",
				d.Reason, d, stepClass, preferLocal, c.Mode)
			return false
		}
		return true
	}

	// testing/quick with a fixed seed: deterministic, >= 100 iterations.
	qcfg := &quick.Config{MaxCount: 200, Rand: rand.New(rand.NewSource(6))}
	if err := quick.Check(property, qcfg); err != nil {
		t.Fatalf("Property 6 (prefer_local) failed: %v", err)
	}
}
