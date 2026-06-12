package knowledge

import "errors"

var (
	ErrInvalidGraph      = errors.New("invalid knowledge graph")
	ErrInvalidNodeID     = errors.New("invalid node id")
	ErrInvalidEdgeID     = errors.New("invalid edge id")
	ErrInvalidNode       = errors.New("invalid graph node")
	ErrInvalidEdge       = errors.New("invalid graph edge")
	ErrInvalidProvenance = errors.New("invalid provenance")
	ErrNotFound          = errors.New("not found")
	ErrNotImplemented    = errors.New("not implemented")
)
