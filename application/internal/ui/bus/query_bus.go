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
	case GetRunDetailQuery:
		return b.handleGetRunDetail(ctx, q)
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
	case GetRecommendedActionsQuery:
		return b.handleGetRecommendedActions(ctx, q)
	case GetAgentTheatreQuery:
		return b.handleGetAgentTheatre(ctx, q)
	case GetReplayPackageQuery:
		return b.handleGetReplayPackage(ctx, q)
	case GetPrototypePipelineQuery:
		return b.handleGetPrototypePipeline(ctx, q)
	case GetMissionControlSnapshotQuery:
		return b.handleGetMissionControlSnapshot(ctx, q)
	case GetGraphRollbackImpactQuery:
		return b.handleGetGraphRollbackImpact(ctx, q)
	case GetPaletteEntriesQuery:
		return b.handleGetPaletteEntries(ctx, q)
	case GetGraphViewQuery:
		return b.handleGetGraphView(ctx, q)
	case GetGraphNodeDetailQuery:
		return b.handleGetGraphNodeDetail(ctx, q)
	case GetFlowStepDetailQuery:
		return b.handleGetFlowStepDetail(ctx, q)
	case GetKnowledgeMatchDetailQuery:
		return b.handleGetKnowledgeMatchDetail(ctx, q)
	case GetTrustDimensionDetailQuery:
		return b.handleGetTrustDimensionDetail(ctx, q)
	case GetReplayEventDetailQuery:
		return b.handleGetReplayEventDetail(ctx, q)
	case GetReplayCompareQuery:
		return b.handleGetReplayCompare(ctx, q)
	case GetReadinessQuery:
		return b.handleGetReadiness(ctx, q)
	case GetOnboardingStateQuery:
		return b.handleGetOnboardingState(ctx, q)
	case GetOnboardingWizardQuery:
		return b.handleGetOnboardingWizard(ctx, q)
	case ValidateOnboardingStepQuery:
		return b.handleValidateOnboardingStep(ctx, q)
	default:
		return nil, fmt.Errorf("ui query not supported: %T", query)
	}
}
