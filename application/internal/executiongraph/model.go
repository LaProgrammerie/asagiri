package executiongraph

import (
	"fmt"
	"sort"
)

// GraphStatus is the lifecycle state of an execution graph (spec §20).
type GraphStatus string

const (
	GraphStatusPlanned    GraphStatus = "planned"
	GraphStatusReady      GraphStatus = "ready"
	GraphStatusRunning    GraphStatus = "running"
	GraphStatusBlocked    GraphStatus = "blocked"
	GraphStatusPaused     GraphStatus = "paused"
	GraphStatusFailed     GraphStatus = "failed"
	GraphStatusCompleted  GraphStatus = "completed"
	GraphStatusAborted    GraphStatus = "aborted"
	GraphStatusRolledBack GraphStatus = "rolled_back"
)

// NodeStatus is the lifecycle state of a graph node (spec §20).
type NodeStatus string

const (
	NodeStatusPending    NodeStatus = "pending"
	NodeStatusReady      NodeStatus = "ready"
	NodeStatusRunning    NodeStatus = "running"
	NodeStatusSucceeded  NodeStatus = "succeeded"
	NodeStatusFailed     NodeStatus = "failed"
	NodeStatusSkipped    NodeStatus = "skipped"
	NodeStatusBlocked    NodeStatus = "blocked"
	NodeStatusRolledBack NodeStatus = "rolled_back"
)

// NodeType classifies executable units in the graph (spec §8).
type NodeType string

const (
	NodeTypeInvestigation          NodeType = "investigation"
	NodeTypeEnrichment             NodeType = "enrichment"
	NodeTypeArchitectureDerivation NodeType = "architecture_derivation"
	NodeTypeContractGeneration     NodeType = "contract_generation"
	NodeTypeImplementation         NodeType = "implementation"
	NodeTypeValidation             NodeType = "validation"
	NodeTypeReview                 NodeType = "review"
	NodeTypeTrustVerification      NodeType = "trust_verification"
	NodeTypeDocumentation          NodeType = "documentation"
	NodeTypeReleaseCheck           NodeType = "release_check"
	NodeTypeManualApproval         NodeType = "manual_approval"
	NodeTypeRollback               NodeType = "rollback"
)

// EdgeType classifies dependency edges (spec §9).
type EdgeType string

const (
	EdgeTypeRequires              EdgeType = "requires"
	EdgeTypeBlocks                EdgeType = "blocks"
	EdgeTypeValidates             EdgeType = "validates"
	EdgeTypeProducesContextFor    EdgeType = "produces_context_for"
	EdgeTypeMustRunAfter          EdgeType = "must_run_after"
	EdgeTypeCanRunAfter           EdgeType = "can_run_after"
	EdgeTypeParallelWith          EdgeType = "parallel_with"
	EdgeTypeRollbackDependsOn     EdgeType = "rollback_depends_on"
	EdgeTypeRequiresHumanApproval EdgeType = "requires_human_approval"
)

// RiskLevel is a coarse risk label for planning (spec §14).
type RiskLevel string

const (
	RiskLevelLow      RiskLevel = "low"
	RiskLevelMedium   RiskLevel = "medium"
	RiskLevelHigh     RiskLevel = "high"
	RiskLevelCritical RiskLevel = "critical"
)

// RollbackStrategy names a rollback approach (spec §16).
type RollbackStrategy string

const (
	RollbackStrategyWorktreeReset      RollbackStrategy = "worktree_reset"
	RollbackStrategyPatchRevert        RollbackStrategy = "patch_revert"
	RollbackStrategyMigrationDown      RollbackStrategy = "migration_down"
	RollbackStrategyFeatureFlagDisable RollbackStrategy = "feature_flag_disable"
	RollbackStrategyManual             RollbackStrategy = "manual"
)

// Strategy holds execution constraints for the graph (spec §7).
type Strategy struct {
	MaxParallel     int       `yaml:"max_parallel" json:"max_parallel"`
	StopOnRisk      RiskLevel `yaml:"stop_on_risk,omitempty" json:"stop_on_risk,omitempty"`
	StrictTrust     bool      `yaml:"strict_trust,omitempty" json:"strict_trust,omitempty"`
	CheckpointEvery string    `yaml:"checkpoint_every,omitempty" json:"checkpoint_every,omitempty"`
	Budget          float64   `yaml:"budget,omitempty" json:"budget,omitempty"`
}

// GraphNode is one executable unit in the graph.
type GraphNode struct {
	ID                    string           `yaml:"id" json:"id"`
	Type                  NodeType         `yaml:"type" json:"type"`
	Title                 string           `yaml:"title" json:"title"`
	Status                NodeStatus       `yaml:"status,omitempty" json:"status,omitempty"`
	Task                  string           `yaml:"task,omitempty" json:"task,omitempty"`
	Agent                 string           `yaml:"agent,omitempty" json:"agent,omitempty"`
	Risk                  RiskLevel        `yaml:"risk,omitempty" json:"risk,omitempty"`
	BlastRadius           int              `yaml:"blast_radius,omitempty" json:"blast_radius,omitempty"`
	RequiresHumanApproval bool             `yaml:"requires_human_approval,omitempty" json:"requires_human_approval,omitempty"`
	RollbackStrategy      RollbackStrategy `yaml:"rollback_strategy,omitempty" json:"rollback_strategy,omitempty"`
	EstimatedCost         float64          `yaml:"estimated_cost,omitempty" json:"estimated_cost,omitempty"`
	EstimatedDuration     string           `yaml:"estimated_duration,omitempty" json:"estimated_duration,omitempty"`
	RequiredChecks        []string         `yaml:"required_checks,omitempty" json:"required_checks,omitempty"`
	Outputs               []string         `yaml:"outputs,omitempty" json:"outputs,omitempty"`
}

// GraphEdge connects two nodes with a typed dependency.
type GraphEdge struct {
	From   string   `yaml:"from" json:"from"`
	To     string   `yaml:"to" json:"to"`
	Type   EdgeType `yaml:"type" json:"type"`
	Reason string   `yaml:"reason,omitempty" json:"reason,omitempty"`
}

// Checkpoint marks a resumable point after a node completes (spec §15).
type Checkpoint struct {
	After string `yaml:"after" json:"after"`
}

// RollbackPlan describes graph-level rollback defaults (spec §16).
type RollbackPlan struct {
	Strategy        RollbackStrategy `yaml:"strategy" json:"strategy"`
	PreserveReports bool             `yaml:"preserve_reports,omitempty" json:"preserve_reports,omitempty"`
}

// ExecutionGraph is the persisted execution plan artefact (spec §7).
type ExecutionGraph struct {
	ID          string        `yaml:"id" json:"id"`
	Product     string        `yaml:"product" json:"product"`
	Flow        string        `yaml:"flow,omitempty" json:"flow,omitempty"`
	Status      GraphStatus   `yaml:"status" json:"status"`
	CreatedAt   string        `yaml:"created_at" json:"created_at"`
	Strategy    Strategy      `yaml:"strategy" json:"strategy"`
	Nodes       []GraphNode   `yaml:"nodes" json:"nodes"`
	Edges       []GraphEdge   `yaml:"edges" json:"edges"`
	Checkpoints []Checkpoint  `yaml:"checkpoints,omitempty" json:"checkpoints,omitempty"`
	Rollback    *RollbackPlan `yaml:"rollback,omitempty" json:"rollback,omitempty"`
}

// Validate checks structural integrity of the graph model.
func (g ExecutionGraph) Validate() error {
	if err := ValidateGraphID(g.ID); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidGraph, err)
	}
	if g.Product == "" {
		return fmt.Errorf("%w: product required", ErrInvalidGraph)
	}
	if !isGraphStatus(g.Status) {
		return fmt.Errorf("%w: invalid graph status %q", ErrInvalidGraph, g.Status)
	}
	if g.Strategy.MaxParallel < 1 {
		return fmt.Errorf("%w: strategy.max_parallel must be >= 1", ErrInvalidGraph)
	}
	if g.Strategy.StopOnRisk != "" && !isRiskLevel(g.Strategy.StopOnRisk) {
		return fmt.Errorf("%w: invalid stop_on_risk %q", ErrInvalidGraph, g.Strategy.StopOnRisk)
	}
	if g.Strategy.CheckpointEvery != "" {
		if err := ValidateCheckpointEvery(g.Strategy.CheckpointEvery); err != nil {
			return fmt.Errorf("%w: %v", ErrInvalidGraph, err)
		}
	}
	if g.Rollback != nil && !isRollbackStrategy(g.Rollback.Strategy) {
		return fmt.Errorf("%w: invalid rollback strategy %q", ErrInvalidGraph, g.Rollback.Strategy)
	}

	nodeIDs := make(map[string]struct{}, len(g.Nodes))
	for _, n := range g.Nodes {
		if n.ID == "" {
			return fmt.Errorf("%w: node id required", ErrInvalidGraph)
		}
		if _, dup := nodeIDs[n.ID]; dup {
			return fmt.Errorf("%w: duplicate node id %q", ErrInvalidGraph, n.ID)
		}
		nodeIDs[n.ID] = struct{}{}
		if !isNodeType(n.Type) {
			return fmt.Errorf("%w: invalid node type %q on node %q", ErrInvalidGraph, n.Type, n.ID)
		}
		if n.Status != "" && !isNodeStatus(n.Status) {
			return fmt.Errorf("%w: invalid node status %q on node %q", ErrInvalidGraph, n.Status, n.ID)
		}
		if n.Risk != "" && !isRiskLevel(n.Risk) {
			return fmt.Errorf("%w: invalid risk %q on node %q", ErrInvalidGraph, n.Risk, n.ID)
		}
		if n.RollbackStrategy != "" && !isRollbackStrategy(n.RollbackStrategy) {
			return fmt.Errorf("%w: invalid rollback strategy %q on node %q", ErrInvalidGraph, n.RollbackStrategy, n.ID)
		}
	}

	for i, e := range g.Edges {
		if !isEdgeType(e.Type) {
			return fmt.Errorf("%w: invalid edge type %q at edge %d", ErrInvalidGraph, e.Type, i)
		}
		if _, ok := nodeIDs[e.From]; !ok {
			return fmt.Errorf("%w: edge from unknown node %q", ErrInvalidGraph, e.From)
		}
		if _, ok := nodeIDs[e.To]; !ok {
			return fmt.Errorf("%w: edge to unknown node %q", ErrInvalidGraph, e.To)
		}
	}

	for i, cp := range g.Checkpoints {
		if _, ok := nodeIDs[cp.After]; !ok {
			return fmt.Errorf("%w: checkpoint %d references unknown node %q", ErrInvalidGraph, i, cp.After)
		}
	}

	return nil
}

// SortedNodes returns nodes ordered by ID for deterministic rendering.
func (g ExecutionGraph) SortedNodes() []GraphNode {
	out := append([]GraphNode(nil), g.Nodes...)
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

// SortedEdges returns edges ordered by from, to, then type.
func (g ExecutionGraph) SortedEdges() []GraphEdge {
	out := append([]GraphEdge(nil), g.Edges...)
	sort.Slice(out, func(i, j int) bool {
		if out[i].From != out[j].From {
			return out[i].From < out[j].From
		}
		if out[i].To != out[j].To {
			return out[i].To < out[j].To
		}
		return out[i].Type < out[j].Type
	})
	return out
}

func isGraphStatus(s GraphStatus) bool {
	switch s {
	case GraphStatusPlanned, GraphStatusReady, GraphStatusRunning, GraphStatusBlocked,
		GraphStatusPaused, GraphStatusFailed, GraphStatusCompleted, GraphStatusAborted,
		GraphStatusRolledBack:
		return true
	default:
		return false
	}
}

func isNodeStatus(s NodeStatus) bool {
	switch s {
	case NodeStatusPending, NodeStatusReady, NodeStatusRunning, NodeStatusSucceeded,
		NodeStatusFailed, NodeStatusSkipped, NodeStatusBlocked, NodeStatusRolledBack:
		return true
	default:
		return false
	}
}

func isNodeType(t NodeType) bool {
	switch t {
	case NodeTypeInvestigation, NodeTypeEnrichment, NodeTypeArchitectureDerivation,
		NodeTypeContractGeneration, NodeTypeImplementation, NodeTypeValidation,
		NodeTypeReview, NodeTypeTrustVerification, NodeTypeDocumentation,
		NodeTypeReleaseCheck, NodeTypeManualApproval, NodeTypeRollback:
		return true
	default:
		return false
	}
}

func isEdgeType(t EdgeType) bool {
	switch t {
	case EdgeTypeRequires, EdgeTypeBlocks, EdgeTypeValidates, EdgeTypeProducesContextFor,
		EdgeTypeMustRunAfter, EdgeTypeCanRunAfter, EdgeTypeParallelWith,
		EdgeTypeRollbackDependsOn, EdgeTypeRequiresHumanApproval:
		return true
	default:
		return false
	}
}

func isRiskLevel(r RiskLevel) bool {
	switch r {
	case RiskLevelLow, RiskLevelMedium, RiskLevelHigh, RiskLevelCritical:
		return true
	default:
		return false
	}
}

func isRollbackStrategy(s RollbackStrategy) bool {
	switch s {
	case RollbackStrategyWorktreeReset, RollbackStrategyPatchRevert,
		RollbackStrategyMigrationDown, RollbackStrategyFeatureFlagDisable,
		RollbackStrategyManual:
		return true
	default:
		return false
	}
}
