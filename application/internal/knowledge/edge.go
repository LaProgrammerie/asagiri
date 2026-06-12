package knowledge

import (
	"fmt"
	"strings"
	"time"
)

// EdgeType classifies knowledge graph relations (spec-my-E §4.2).
type EdgeType string

const (
	EdgeTypeImplements EdgeType = "implements"
	EdgeTypeCalls      EdgeType = "calls"
	EdgeTypeEmits      EdgeType = "emits"
	EdgeTypeRequires   EdgeType = "requires"
	EdgeTypeValidates  EdgeType = "validates"
	EdgeTypeTests      EdgeType = "tests"
	EdgeTypeObserves   EdgeType = "observes"
	EdgeTypeConfigures EdgeType = "configures"
	EdgeTypeDependsOn  EdgeType = "depends_on"
	EdgeTypeOwns       EdgeType = "owns"
	EdgeTypeImpacts    EdgeType = "impacts"
	EdgeTypeBreaks     EdgeType = "breaks"
	EdgeTypeProduces   EdgeType = "produces"
	EdgeTypeConsumes   EdgeType = "consumes"
	EdgeTypeReviewedBy EdgeType = "reviewed_by"
	EdgeTypeFailedIn   EdgeType = "failed_in"
	EdgeTypeCosts      EdgeType = "costs"
)

// GraphEdge connects two nodes (spec-my-E §9).
type GraphEdge struct {
	ID         string         `json:"id"`
	From       string         `json:"from"`
	To         string         `json:"to"`
	Type       EdgeType       `json:"type"`
	Properties map[string]any `json:"properties,omitempty"`
	Source     GraphSource    `json:"source"`
	Confidence float64        `json:"confidence"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// EdgeID builds a hybrid id for an edge (type:from>to with node colons escaped).
func EdgeID(edgeType EdgeType, from, to string) string {
	return string(edgeType) + ":" + escapeNodeRef(from) + ">" + escapeNodeRef(to)
}

func escapeNodeRef(id string) string {
	return strings.ReplaceAll(id, ":", "_")
}

// ValidateEdgeID checks hybrid id shape and edge type prefix.
func ValidateEdgeID(id string) error {
	if id == "" {
		return fmt.Errorf("%w: id required", ErrInvalidEdgeID)
	}
	if strings.Contains(id, "..") || strings.ContainsAny(id, `/\`) {
		return fmt.Errorf("%w: path segments not allowed", ErrInvalidEdgeID)
	}
	if !hybridIDPattern.MatchString(id) {
		return fmt.Errorf("%w: expected type:stable_key", ErrInvalidEdgeID)
	}
	prefix, _, ok := strings.Cut(id, ":")
	if !ok || !isEdgeType(EdgeType(prefix)) {
		return fmt.Errorf("%w: unknown edge type prefix %q", ErrInvalidEdgeID, prefix)
	}
	return nil
}

// Validate checks edge fields.
func (e GraphEdge) Validate() error {
	if err := ValidateEdgeID(e.ID); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEdge, err)
	}
	prefix, _, _ := strings.Cut(e.ID, ":")
	if EdgeType(prefix) != e.Type {
		return fmt.Errorf("%w: id prefix %q does not match type %q", ErrInvalidEdge, prefix, e.Type)
	}
	if !isEdgeType(e.Type) {
		return fmt.Errorf("%w: invalid type %q", ErrInvalidEdge, e.Type)
	}
	if e.From == "" || e.To == "" {
		return fmt.Errorf("%w: from and to required", ErrInvalidEdge)
	}
	if err := ValidateNodeID(e.From); err != nil {
		return fmt.Errorf("%w: invalid from node: %v", ErrInvalidEdge, err)
	}
	if err := ValidateNodeID(e.To); err != nil {
		return fmt.Errorf("%w: invalid to node: %v", ErrInvalidEdge, err)
	}
	if err := ValidateUpsertProvenance(e.Source, e.Confidence); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidEdge, err)
	}
	return nil
}

func isEdgeType(t EdgeType) bool {
	switch t {
	case EdgeTypeImplements, EdgeTypeCalls, EdgeTypeEmits, EdgeTypeRequires,
		EdgeTypeValidates, EdgeTypeTests, EdgeTypeObserves, EdgeTypeConfigures,
		EdgeTypeDependsOn, EdgeTypeOwns, EdgeTypeImpacts, EdgeTypeBreaks,
		EdgeTypeProduces, EdgeTypeConsumes, EdgeTypeReviewedBy, EdgeTypeFailedIn,
		EdgeTypeCosts:
		return true
	default:
		return false
	}
}
