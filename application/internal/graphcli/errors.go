package graphcli

import "errors"

var (
	// ErrCIFailed is returned when graph plan/run fails CI policy.
	ErrCIFailed = errors.New("execution graph failed CI policy")
	// ErrNotEnabled is returned when execution_graph is disabled in config.
	ErrNotEnabled = errors.New("execution graph disabled in config")
	// ErrFlowRequired is returned when --flow is missing.
	ErrFlowRequired = errors.New("flow required: use --flow")
)
