package executiongraph

import "context"

// NodeExecutor runs agent work for one graph node after coordination assignment (spec-my-D §7).
type NodeExecutor func(ctx context.Context, graph ExecutionGraph, node GraphNode, assignment CoordinationAssignment) error
