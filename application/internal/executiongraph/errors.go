package executiongraph

import "errors"

var (
	ErrInvalidGraph            = errors.New("invalid execution graph")
	ErrInvalidGraphID          = errors.New("invalid graph id")
	ErrInvalidTransition       = errors.New("invalid graph state transition")
	ErrInvalidNodeTransition   = errors.New("invalid node state transition")
	ErrGraphAlreadyInState     = errors.New("graph already in target state")
	ErrNodeAlreadyInState      = errors.New("node already in target state")
	ErrNotImplemented          = errors.New("not implemented")
	ErrCycleDetected           = errors.New("dependency cycle detected")
	ErrMissingRollbackStrategy = errors.New("high-risk node missing rollback strategy")
	ErrTrustGateBlocked        = errors.New("trust gate blocked execution")
)
