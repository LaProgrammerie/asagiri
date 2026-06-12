// Package static wraps local static analysis collectors (spec-my-A §25.8).
package static

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
)

// Collect runs grep, symbols, and dependency hints for a feature scope.
func Collect(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (investigation.InvestigationResult, error) {
	return investigation.Run(ctx, repoRoot, feature, taskID, cfg)
}
