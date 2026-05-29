package bus

import (
	"context"
	"fmt"
)

type commandBus struct {
	deps Deps
}

// NewCommandBus builds the command bus with application handlers.
func NewCommandBus(deps Deps) CommandBus {
	return &commandBus{deps: deps.withDefaults()}
}

func (b *commandBus) Dispatch(ctx context.Context, cmd Command) (CommandResult, error) {
	if cmd == nil {
		return CommandResult{}, fmt.Errorf("ui command nil")
	}
	switch typed := cmd.(type) {
	case StartWorkCommand:
		return b.deps.StartWork(ctx, b.deps, typed)
	case RunInvestigationCommand:
		return b.deps.Investigate(ctx, b.deps, typed)
	case VerifyTrustCommand:
		return b.deps.VerifyTrust(ctx, b.deps, typed)
	case BuildKnowledgeGraphCommand:
		return b.deps.BuildKnowledgeGraph(ctx, b.deps, typed)
	case ReplayRunCommand:
		return b.deps.ReplayRun(ctx, b.deps, typed)
	case GraphRollbackCommand:
		return b.deps.GraphRollback(ctx, b.deps, typed)
	case ExportEventsCommand:
		return b.deps.ExportEvents(ctx, b.deps, typed)
	case ExportGraphCommand:
		return b.deps.ExportGraph(ctx, b.deps, typed)
	case GraphResumeCommand:
		return b.deps.GraphResume(ctx, b.deps, typed)
	case AnalyzeKnowledgeImpactCommand:
		return b.deps.AnalyzeKnowledgeImpact(ctx, b.deps, typed)
	case BuildKnowledgeContextCommand:
		return b.deps.BuildKnowledgeContext(ctx, b.deps, typed)
	case CompareReplayCommand:
		return b.deps.CompareReplay(ctx, b.deps, typed)
	case ExplainReplayDivergenceCommand:
		return b.deps.ExplainReplayDivergence(ctx, b.deps, typed)
	case PrototypeCreateCommand:
		return b.deps.PrototypeCreate(ctx, b.deps, typed)
	case FlowsExtractCommand:
		return b.deps.FlowsExtract(ctx, b.deps, typed)
	case ContractsExtractCommand:
		return b.deps.ContractsExtract(ctx, b.deps, typed)
	case SpecGenerateFromProductCommand:
		return b.deps.SpecGenerateFromProduct(ctx, b.deps, typed)
	default:
		return CommandResult{}, fmt.Errorf("ui command not supported: %T", cmd)
	}
}
