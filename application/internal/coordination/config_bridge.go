package coordination

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
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

// AssignerFromConfig returns ScoringAssigner when profiles exist, otherwise DefaultAssigner.
func AssignerFromConfig(cfg *config.Config) AgentAssigner {
	ac := AssignerConfigFromConfig(cfg)
	if len(ac.Profiles) > 0 {
		return &ScoringAssigner{Config: ac}
	}
	return &DefaultAssigner{Config: ac}
}

// MergeEvaluatorFromConfig builds merge policy evaluator from config.
func MergeEvaluatorFromConfig(cfg *config.Config) *MergeEvaluator {
	if cfg == nil {
		return &MergeEvaluator{}
	}
	co := cfg.Coordination
	return &MergeEvaluator{
		Require: append([]string(nil), co.Merge.Require...),
		BlockIf: append([]string(nil), co.Merge.BlockIf...),
	}
}

// CoordinatorServices bundles Lot 4 coordination helpers wired from config.
type CoordinatorServices struct {
	Budget     BudgetTracker
	Conflict   ConflictDetector
	Escalation EscalationHandler
	Merge      *MergeEvaluator
}

// CoordinatorServicesFromConfig wires budget, conflict, escalation, and merge from config.
func CoordinatorServicesFromConfig(cfg *config.Config) CoordinatorServices {
	return CoordinatorServices{
		Budget:     NewMemoryBudgetTracker(),
		Conflict:   DefaultConflictDetector{},
		Escalation: EscalationHandlerFromConfig(cfg),
		Merge:      MergeEvaluatorFromConfig(cfg),
	}
}

// NewDefaultCoordinator wires a coordinator from repo config.
func NewDefaultCoordinator(cfg *config.Config, repoRoot string, emitter *CoordinationEmitter) *DefaultCoordinator {
	return &DefaultCoordinator{
		Assigner: AssignerFromConfig(cfg),
		Policies: PoliciesFromConfig(cfg),
		Emitter:  emitter,
		Pipeline: NewDefaultPipeline(cfg),
		RepoRoot: repoRoot,
	}
}

// NewFullCoordinator wires coordinator with handoff persistence and all services.
func NewFullCoordinator(cfg *config.Config, repoRoot string, emitter *CoordinationEmitter) (*DefaultCoordinator, CoordinatorServices) {
	svc := CoordinatorServicesFromConfig(cfg)
	co := NewDefaultCoordinator(cfg, repoRoot, emitter)
	if repoRoot != "" {
		handoffsPath := ""
		if cfg != nil {
			handoffsPath = cfg.Coordination.HandoffsPath
		}
		co.Handoff = &DefaultHandoffBuilder{RepoRoot: repoRoot, HandoffsPath: handoffsPath}
	}
	return co, svc
}

// GraphCoordinator adapts an AgentCoordinator for executiongraph.RunOptions (breaks import cycle in runner).
func GraphCoordinator(coord AgentCoordinator) executiongraph.GraphCoordinator {
	if coord == nil {
		return nil
	}
	return func(ctx context.Context, graph executiongraph.ExecutionGraph) (executiongraph.CoordinationResult, error) {
		res, err := coord.Coordinate(ctx, graph)
		if err != nil {
			return executiongraph.CoordinationResult{}, err
		}
		assignments := make([]executiongraph.CoordinationAssignment, len(res.Assignments))
		for i, a := range res.Assignments {
			assignments[i] = executiongraph.CoordinationAssignment{
				NodeID:    a.NodeID,
				AgentRef:  a.AgentRef,
				Role:      string(a.Role),
				Isolation: string(a.Isolation),
				ProfileID: a.ProfileID,
			}
		}
		return executiongraph.CoordinationResult{
			Graph:       res.Graph,
			Assignments: assignments,
		}, nil
	}
}
