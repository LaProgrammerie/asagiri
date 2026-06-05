package routing

// Feature: audit-coherence-consolidation, Property 9
//
// Property 9 (design P9) — Erreur guidée sans panic quand aucun backend ne
// correspond:
// For any config where no adequate declared Agent_Backend exists (cfg.Agents
// empty, or only cloud backends declared while a local backend is required),
// Route returns an error wrapping ErrNoDeclaredBackend, yields the zero
// Decision (so no undeclared backend is ever selected), and never panics.
//
// Validates: Requirements 4.7

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// p9Case is the generated input space for Property 9. Every generated case is
// constructed so that Route cannot resolve a declared backend: either the agent
// map is empty (scenario A), or only cloud backends are declared while a local
// backend is forced (scenario B). Both families guarantee err != nil, which
// keeps the property non-vacuous.
type p9Case struct {
	emptyAgents     bool     // scenario A: no Agent_Backend declared at all
	cloudAgents     []string // scenario B: declared cloud-only backends
	defaultAgent    string   // "" or an undeclared name (never a declared key)
	defaultEnricher string   // "" or an undeclared name (never a declared key)
	stepClass       string
	preferLocal     bool
	noCloud         bool
	allowCloud      bool
}

// p9CloudPool is the pool of cloud Agent_Backend names used by scenario B. None
// are declared as local (they carry a Command, never a localhost Endpoint).
var p9CloudPool = []string{"cursor", "claude", "codex", "gemini"}

// p9ClassPool is the pool of step classes exercised by the generator, including
// the empty/unknown class.
var p9ClassPool = []string{"spec", "plan", "enrich", "dev", "verify", "review", "implement", ""}

// p9UndeclaredNames are names that are never added to cfg.Agents, used for
// DefaultAgent / DefaultEnricher so the resolved agent (when non-empty) is never
// a declared key — driving the undeclared-backend guard (Requirement 4.7).
var p9UndeclaredNames = []string{"", "ghost-agent", "missing-enricher"}

// Generate implements quick.Generator with a smart generator constrained to the
// "no adequate declared backend" input space. It picks one of two guaranteed-
// error scenarios and varies the orthogonal flags and step class.
func (p9Case) Generate(r *rand.Rand, _ int) reflect.Value {
	tc := p9Case{
		stepClass:   p9ClassPool[r.Intn(len(p9ClassPool))],
		preferLocal: r.Intn(2) == 0,
		allowCloud:  r.Intn(2) == 0,
		// DefaultAgent / DefaultEnricher are empty or undeclared on purpose: they
		// must never resolve to a declared key of cfg.Agents.
		defaultAgent:    p9UndeclaredNames[r.Intn(len(p9UndeclaredNames))],
		defaultEnricher: p9UndeclaredNames[r.Intn(len(p9UndeclaredNames))],
	}

	if r.Intn(2) == 0 {
		// Scenario A: empty agent map. Every branch fails to resolve a declared
		// backend regardless of the flags, so noCloud varies freely.
		tc.emptyAgents = true
		tc.noCloud = r.Intn(2) == 0
		return reflect.ValueOf(tc)
	}

	// Scenario B: declare 1-3 cloud-only backends but force the local branch via
	// noCloud, where no local backend is declared -> guided error.
	tc.emptyAgents = false
	tc.noCloud = true
	for n := 1 + r.Intn(3); n > 0; n-- {
		tc.cloudAgents = append(tc.cloudAgents, p9CloudPool[r.Intn(len(p9CloudPool))])
	}
	return reflect.ValueOf(tc)
}

// build assembles a config.Config from the generated case. In scenario B the
// declared backends are cloud-only (Command set, no localhost Endpoint, no local
// model class), so firstLocalDeclared cannot find a local backend.
func (tc p9Case) build() *config.Config {
	cfg := config.NewTestConfig("p9")
	cfg.Agents = map[string]config.Agent{}
	if !tc.emptyAgents {
		for _, name := range tc.cloudAgents {
			cfg.Agents[name] = config.Agent{Command: name}
		}
	}
	cfg.Work.DefaultAgent = tc.defaultAgent
	cfg.Work.DefaultEnricher = tc.defaultEnricher
	cfg.Routing.DefaultStrategy = "cost_aware"
	cfg.Routing.Strategies = map[string]config.RoutingStrategy{"cost_aware": {}}
	return cfg
}

// p9Result captures the outcome of a Route call guarded by recover, so the
// property can prove the absence of any panic (Requirement 4.7).
type p9Result struct {
	decision Decision
	err      error
	panicked bool
}

// p9CallNoPanic invokes Route under a deferred recover. If Route panics, the
// recover records it and the named returns keep their zero values.
func p9CallNoPanic(cfg *config.Config, stepClass string, preferLocal, noCloud, allowCloud bool) (res p9Result) {
	defer func() {
		if r := recover(); r != nil {
			res.panicked = true
		}
	}()
	res.decision, res.err = Route(cfg, stepClass, preferLocal, noCloud, allowCloud)
	return res
}

// TestRouteGuidedErrorNoPanicProperty asserts Property 9: when no adequate
// declared backend exists, Route returns ErrNoDeclaredBackend with the zero
// Decision and never panics.
func TestRouteGuidedErrorNoPanicProperty(t *testing.T) {
	// Feature: audit-coherence-consolidation, Property 9 (design P9).
	// Validates: Requirements 4.7
	property := func(tc p9Case) bool {
		cfg := tc.build()
		res := p9CallNoPanic(cfg, tc.stepClass, tc.preferLocal, tc.noCloud, tc.allowCloud)

		if res.panicked {
			t.Errorf("Route panicked for case %+v (must return a guided error, never panic)", tc)
			return false
		}
		if res.err == nil {
			t.Errorf("expected a guided error for case %+v, got nil (agents=%v)", tc, cfg.Agents)
			return false
		}
		if !errors.Is(res.err, ErrNoDeclaredBackend) {
			t.Errorf("error %v does not wrap ErrNoDeclaredBackend for case %+v", res.err, tc)
			return false
		}
		// The zero Decision proves no undeclared backend was selected (Agent is
		// empty, no field is set) — Requirement 4.7.
		if !reflect.DeepEqual(res.decision, Decision{}) {
			t.Errorf("expected zero Decision on error, got %+v for case %+v", res.decision, tc)
			return false
		}
		return true
	}

	cfg := &quick.Config{
		MaxCount: 200, // >= 100 iterations
		Rand:     rand.New(rand.NewSource(9)),
	}
	if err := quick.Check(property, cfg); err != nil {
		t.Fatalf("Property 9 (erreur guidée sans panic) violated: %v", err)
	}
}
