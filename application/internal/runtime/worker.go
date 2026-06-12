package runtime

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"
)

const workerTick = 500 * time.Millisecond

// RunWorker is the long-running daemon loop (spec-my-A §24.3).
func RunWorker(ctx context.Context, repoRoot string) error {
	store, err := Open(repoRoot)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	workerID := uuid.NewString()
	if err := store.TouchWorkerHeartbeat(workerID); err != nil {
		return err
	}
	_, _ = store.EmitEvent("worker.started", "daemon", "", "", map[string]any{"worker_id": workerID})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	ticker := time.NewTicker(workerTick)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			_, _ = store.EmitEvent("worker.stopped", "daemon", "", "", nil)
			return ctx.Err()
		case <-sigCh:
			_, _ = store.EmitEvent("worker.stopped", "daemon", "", "", map[string]any{"signal": "interrupt"})
			return nil
		case <-ticker.C:
			_ = store.TouchWorkerHeartbeat(workerID)
			_, _ = store.CollectMetrics()
			_, _ = store.ProcessHookQueue(ctx, 5)
		}
	}
}
