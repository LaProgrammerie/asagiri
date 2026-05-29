package coordination

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// PoliciesFromConfig builds a PolicyEvaluator from the coordination block.
func PoliciesFromConfig(cfg *config.Config) PolicyEvaluator {
	if cfg == nil {
		return PolicyEvaluator{}
	}
	co := cfg.Coordination
	return PolicyEvaluator{
		Policies: CoordinationPolicies{
			MaxParallelAgents:        co.MaxParallelAgents,
			RequireIndependentReview: co.RequireIndependentReview,
			AllowSelfReview:          co.AllowSelfReview,
			RequireSecurityReviewFor: append([]string(nil), co.RequireSecurityReviewFor...),
			DefaultIsolation:         IsolationMode(co.DefaultIsolation),
		},
	}
}

// AssignerConfigFromConfig builds assigner settings from config.
func AssignerConfigFromConfig(cfg *config.Config) AssignerConfig {
	if cfg == nil {
		return AssignerConfig{DefaultIsolation: IsolationIsolatedWorktree}
	}
	co := cfg.Coordination
	profiles := make(map[string]AgentProfile, len(co.Profiles))
	for id, p := range co.Profiles {
		profiles[id] = AgentProfile{
			ID:               id,
			Role:             AgentRole(p.Role),
			Agent:            p.Agent,
			Capabilities:     append([]string(nil), p.Capabilities...),
			Restrictions:     append([]string(nil), p.Restrictions...),
			MaxContextTokens: p.MaxContextTokens,
			Isolation:        IsolationMode(p.Isolation),
		}
	}
	return AssignerConfig{
		DefaultIsolation: IsolationMode(co.DefaultIsolation),
		Assignment:       co.Assignment,
		Profiles:         profiles,
	}
}

// NewDefaultCoordinator wires a coordinator from repo config.
func NewDefaultCoordinator(cfg *config.Config, repoRoot string, emitter *CoordinationEmitter) *DefaultCoordinator {
	return &DefaultCoordinator{
		Assigner: &DefaultAssigner{Config: AssignerConfigFromConfig(cfg)},
		Policies: PoliciesFromConfig(cfg),
		Emitter:  emitter,
		RepoRoot: repoRoot,
	}
}
