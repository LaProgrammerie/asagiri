package bus

import (
	"context"
	"fmt"
)

type queryBus struct {
	deps Deps
}

// NewQueryBus builds the read-only query bus.
func NewQueryBus(deps Deps) QueryBus {
	return &queryBus{deps: deps.withDefaults()}
}

func (b *queryBus) Query(ctx context.Context, query Query) (QueryResult, error) {
	switch q := query.(type) {
	case GetRuntimeStatusQuery:
		return b.handleGetRuntimeStatus(ctx, q)
	case ListRunsQuery:
		return b.handleListRuns(ctx, q)
	case GetRecentEventsQuery:
		return b.handleGetRecentEvents(ctx, q)
	case GetTrustSummaryQuery:
		return b.handleGetTrustSummary(ctx, q)
	case ListActiveAgentsQuery:
		return b.handleListActiveAgents(ctx, q)
	case GetFlowGraphQuery:
		return b.handleGetFlowGraph(ctx, q)
	case GetFlowExplorerQuery:
		return b.handleGetFlowExplorer(ctx, q)
	case GetGraphExplorerQuery:
		return b.handleGetGraphExplorer(ctx, q)
	case SearchKnowledgeQuery:
		return b.handleSearchKnowledge(ctx, q)
	case GetTrustExplorerQuery:
		return b.handleGetTrustExplorer(ctx, q)
	case GetExplainQuery:
		return b.handleGetExplain(ctx, q)
	case GetAgentTheatreQuery:
		return b.handleGetAgentTheatre(ctx, q)
	case GetReplayPackageQuery:
		return b.handleGetReplayPackage(ctx, q)
	case GetPrototypePipelineQuery:
		return b.handleGetPrototypePipeline(ctx, q)
	case GetMissionControlSnapshotQuery:
		return b.handleGetMissionControlSnapshot(ctx, q)
	default:
		return nil, fmt.Errorf("ui query not supported: %T", query)
	}
}
