package coordination

import "errors"

var (
	ErrInvalidGraph      = errors.New("coordination: invalid execution graph")
	ErrInvalidRole       = errors.New("coordination: invalid agent role")
	ErrInvalidIsolation  = errors.New("coordination: invalid isolation mode")
	ErrUnknownProfile    = errors.New("coordination: unknown agent profile")
	ErrPolicyViolation   = errors.New("coordination: policy violation")
	ErrHandoffPersist    = errors.New("coordination: handoff persistence failed")
	ErrNotImplemented    = errors.New("coordination: not implemented")
	ErrInvalidHandoff    = errors.New("coordination: invalid handoff")
	ErrInvalidAssignment = errors.New("coordination: invalid assignment")
)
