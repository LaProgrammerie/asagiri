package knowledge

import (
	"fmt"
	"regexp"
	"strings"
	"time"
)

// NodeType classifies knowledge graph nodes (spec-my-E §4.1).
type NodeType string

const (
	NodeTypeProduct        NodeType = "product"
	NodeTypeFlow           NodeType = "flow"
	NodeTypeFlowStep       NodeType = "flow_step"
	NodeTypeAction         NodeType = "action"
	NodeTypeScreen         NodeType = "screen"
	NodeTypeAPIOperation   NodeType = "api_operation"
	NodeTypeEvent          NodeType = "event"
	NodeTypePermission     NodeType = "permission"
	NodeTypeMetric         NodeType = "metric"
	NodeTypeTrace          NodeType = "trace"
	NodeTypeLog            NodeType = "log"
	NodeTypeModule         NodeType = "module"
	NodeTypeFile           NodeType = "file"
	NodeTypeSymbol         NodeType = "symbol"
	NodeTypeTest           NodeType = "test"
	NodeTypeMigration      NodeType = "migration"
	NodeTypeInfraResource  NodeType = "infra_resource"
	NodeTypeConfig         NodeType = "config"
	NodeTypeSecretBoundary NodeType = "secret_boundary"
	NodeTypeADR            NodeType = "adr"
	NodeTypeReview         NodeType = "review"
	NodeTypeIncident       NodeType = "incident"
	NodeTypeCostCenter     NodeType = "cost_center"
	NodeTypeAgent          NodeType = "agent"
)

var hybridIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]*:[^:+/\\]+$`)

// GraphNode is one vertex in the engineering knowledge graph (spec-my-E §8).
type GraphNode struct {
	ID         string         `json:"id"`
	Type       NodeType       `json:"type"`
	Name       string         `json:"name,omitempty"`
	Path       string         `json:"path,omitempty"`
	Properties map[string]any `json:"properties,omitempty"`
	Source     GraphSource    `json:"source"`
	Confidence float64        `json:"confidence"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// NodeID builds a hybrid id type:stable_key.
func NodeID(nodeType NodeType, stableKey string) string {
	return string(nodeType) + ":" + stableKey
}

// ValidateNodeID checks hybrid id shape and type prefix.
func ValidateNodeID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: id required", ErrInvalidNodeID)
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("%w: path segments not allowed", ErrInvalidNodeID)
	}
	if !hybridIDPattern.MatchString(id) {
		return fmt.Errorf("%w: expected type:stable_key", ErrInvalidNodeID)
	}
	prefix, _, ok := strings.Cut(id, ":")
	if !ok || !isNodeType(NodeType(prefix)) {
		return fmt.Errorf("%w: unknown node type prefix %q", ErrInvalidNodeID, prefix)
	}
	return nil
}

// Validate checks node fields.
func (n GraphNode) Validate() error {
	if err := ValidateNodeID(n.ID); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidNode, err)
	}
	prefix, _, _ := strings.Cut(n.ID, ":")
	if NodeType(prefix) != n.Type {
		return fmt.Errorf("%w: id prefix %q does not match type %q", ErrInvalidNode, prefix, n.Type)
	}
	if !isNodeType(n.Type) {
		return fmt.Errorf("%w: invalid type %q", ErrInvalidNode, n.Type)
	}
	if err := ValidateUpsertProvenance(n.Source, n.Confidence); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidNode, err)
	}
	return nil
}

func isNodeType(t NodeType) bool {
	switch t {
	case NodeTypeProduct, NodeTypeFlow, NodeTypeFlowStep, NodeTypeAction, NodeTypeScreen,
		NodeTypeAPIOperation, NodeTypeEvent, NodeTypePermission, NodeTypeMetric, NodeTypeTrace,
		NodeTypeLog, NodeTypeModule, NodeTypeFile, NodeTypeSymbol, NodeTypeTest,
		NodeTypeMigration, NodeTypeInfraResource, NodeTypeConfig, NodeTypeSecretBoundary,
		NodeTypeADR, NodeTypeReview, NodeTypeIncident, NodeTypeCostCenter, NodeTypeAgent:
		return true
	default:
		return false
	}
}
