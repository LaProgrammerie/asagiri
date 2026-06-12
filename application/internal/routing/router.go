package routing

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// Decision is the model/agent routing outcome (specv3 §11).
type Decision struct {
	Agent     string
	Model     string
	Local     bool
	StepClass string
	Reason    string
}

// ErrNoDeclaredBackend is returned when no Agent_Backend declared in
// config.agents matches the computed routing decision. Callers handle it as a
// value (errors are returned, never panicked, at CLI boundaries — see
// docs/ai/03-standards.md) and no undeclared backend is ever selected
// (Requirements 4.2, 4.7).
var ErrNoDeclaredBackend = errors.New(
	"routing: aucun Agent_Backend déclaré dans config.agents ne correspond à la décision")

// Route picks the agent/model for a step class. It is a pure function: identical
// inputs yield identical Decisions with no observable side effect (Requirement
// 4.1).
//
// Priority order (Requirements 4.3, 4.4, 4.5): no_cloud > prefer_local >
// cloud_heavy > cloud_fast > default. no_cloud prevails over allowCloud and
// preferLocal (Requirement 4.4). The selected Agent is always a declared key of
// cfg.Agents (Requirement 4.2); otherwise Route returns the zero Decision and an
// error wrapping ErrNoDeclaredBackend (Requirement 4.7), never a panic. The
// Reason always belongs to {prefer_local, no_cloud, cloud_heavy, cloud_fast,
// default} (Requirement 4.6).
func Route(cfg *config.Config, stepClass string, preferLocal, noCloud, allowCloud bool) (Decision, error) {
	cls := strings.ToLower(strings.TrimSpace(stepClass))
	if cfg == nil {
		return Decision{}, fmt.Errorf("%w: config nil (classe=%q)", ErrNoDeclaredBackend, cls)
	}

	strategy := cfg.Routing.Strategies["cost_aware"]
	if cfg.Routing.DefaultStrategy != "" {
		if s, ok := cfg.Routing.Strategies[cfg.Routing.DefaultStrategy]; ok {
			strategy = s
		}
	}

	var (
		agent  string
		reason string
		local  bool
	)

	switch {
	case noCloud:
		// Priorité 1 : no_cloud prévaut sur allowCloud et preferLocal (4.4).
		local, reason = true, "no_cloud"
		agent = firstLocalDeclared(cfg)
	case preferLocal || contains(strategy.PreferLocalFor, cls):
		// Priorité 2 : preferLocal explicite ou classe locale par config (4.3).
		local, reason = true, "prefer_local"
		agent = firstLocalDeclared(cfg)
	case contains(strategy.UseCloudHeavyFor, cls):
		// Priorité 3 : classe cloud lourde (4.5).
		reason = "cloud_heavy"
		agent = cloudHeavyDeclared(cfg)
	case contains(strategy.UseCloudFastFor, cls):
		// Priorité 4 : classe cloud rapide (4.5).
		reason = "cloud_fast"
		agent = cfg.Work.DefaultAgent
	default:
		// Priorité 5 : stratégie par défaut (4.5).
		reason = "default"
		agent = cfg.Work.DefaultAgent
	}

	// allowCloud n'influence plus la décision : no_cloud prévaut (4.4) et les
	// autres branches sont pilotées par la classe d'étape et les flags locaux.
	_ = allowCloud

	if agent == "" {
		return Decision{}, fmt.Errorf("%w: classe=%q raison=%q local=%v", ErrNoDeclaredBackend, cls, reason, local)
	}
	if _, ok := cfg.Agents[agent]; !ok {
		return Decision{}, fmt.Errorf("%w: agent=%q non déclaré (classe=%q raison=%q)", ErrNoDeclaredBackend, agent, cls, reason)
	}

	return Decision{
		Agent:     agent,
		Model:     cfg.AgentModel(agent),
		Local:     local,
		StepClass: cls,
		Reason:    reason,
	}, nil
}

// firstLocalDeclared returns a declared local Agent_Backend: the default
// enricher when it is declared in cfg.Agents, otherwise the first declared agent
// that looks local, scanning agent keys in sorted order for determinism
// (Requirement 4.1). Returns "" when no local backend is declared.
func firstLocalDeclared(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if enr := cfg.Work.DefaultEnricher; enr != "" {
		if _, ok := cfg.Agents[enr]; ok {
			return enr
		}
	}
	for _, name := range sortedAgentKeys(cfg) {
		if isLocalAgent(cfg, name) {
			return name
		}
	}
	return ""
}

// cloudHeavyDeclared returns a declared cloud (non-local) Agent_Backend: the
// default agent when it is declared and cloud, otherwise the first declared
// cloud agent in sorted order (Requirement 4.1). Returns "" when no cloud
// backend is declared.
func cloudHeavyDeclared(cfg *config.Config) string {
	if cfg == nil {
		return ""
	}
	if def := cfg.Work.DefaultAgent; def != "" {
		if _, ok := cfg.Agents[def]; ok && !isLocalAgent(cfg, def) {
			return def
		}
	}
	for _, name := range sortedAgentKeys(cfg) {
		if !isLocalAgent(cfg, name) {
			return name
		}
	}
	return ""
}

// sortedAgentKeys returns the declared agent names sorted, for deterministic
// iteration (Go map iteration order is randomized).
func sortedAgentKeys(cfg *config.Config) []string {
	if cfg == nil || len(cfg.Agents) == 0 {
		return nil
	}
	keys := make([]string, 0, len(cfg.Agents))
	for name := range cfg.Agents {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	return keys
}

// isLocalAgent reports whether a declared agent is local, using config signals
// only (no hard-coded agent name): a localhost endpoint or a model profile whose
// class is local.
func isLocalAgent(cfg *config.Config, name string) bool {
	if cfg == nil || name == "" {
		return false
	}
	if a, ok := cfg.Agents[name]; ok {
		ep := strings.ToLower(a.Endpoint)
		if strings.Contains(ep, "localhost") || strings.Contains(ep, "127.0.0.1") {
			return true
		}
	}
	if p, ok := cfg.Models[name]; ok && strings.Contains(strings.ToLower(p.Class), "local") {
		return true
	}
	return false
}

func contains(list []string, item string) bool {
	for _, x := range list {
		if strings.EqualFold(x, item) {
			return true
		}
	}
	return false
}
