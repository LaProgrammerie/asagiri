package bus

import (
	"context"
	"fmt"
)

type commandBus struct {
	deps Deps
}

// NewCommandBus builds the command bus with lot-1 command stubs.
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
	case BuildKnowledgeGraphCommand, ReplayRunCommand:
		return CommandResult{
			Accepted:      false,
			Message:       "not implemented in lot 3",
			CLIEquivalent: cmd.CLIEquivalent(),
		}, nil
	default:
		return CommandResult{}, fmt.Errorf("ui command not supported: %T", cmd)
	}
}
