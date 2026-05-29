package coordination

import (
	"github.com/LaProgrammerie/asagiri/application/internal/executiongraph"
)

// RunnerCoordinator adapts DefaultCoordinator for executiongraph.RunOptions.
func RunnerCoordinator(c *DefaultCoordinator) executiongraph.GraphCoordinator {
	return GraphCoordinator(c)
}
