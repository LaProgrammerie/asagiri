package replay

import "errors"

var (
	ErrInvalidReplayID  = errors.New("invalid replay id")
	ErrReplayNotFound   = errors.New("replay package not found")
	ErrInvalidSource    = errors.New("replay source required: --from-run, --from-graph, or --from-investigation")
	ErrInvalidMode      = errors.New("invalid replay mode")
	ErrStrictDivergence = errors.New("strict replay: divergence detected")
	ErrOfflineViolation = errors.New("offline replay: external call blocked")
	ErrSnapshotName     = errors.New("replay snapshot: name required")
)
