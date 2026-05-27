package asagiri

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// RuntimeClient is the in-process runtime SDK entry (spec-my-A §24.18).
type RuntimeClient struct {
	store *runtime.Store
}

// Connect opens the runtime database for a repository root (in-process embedding).
func Connect(repoRoot string) (*RuntimeClient, error) {
	store, err := runtime.Open(repoRoot)
	if err != nil {
		return nil, err
	}
	return &RuntimeClient{store: store}, nil
}

// Close releases the database handle.
func (c *RuntimeClient) Close() error {
	if c == nil || c.store == nil {
		return nil
	}
	return c.store.Close()
}

// Store exposes the underlying runtime store for advanced embedding.
func (c *RuntimeClient) Store() *runtime.Store {
	if c == nil {
		return nil
	}
	return c.store
}

// StartSession creates an engineering session.
func (c *RuntimeClient) StartSession(name, productID, flowID string) (runtime.Session, error) {
	return c.store.CreateSession(name, productID, flowID)
}

// EmitEvent publishes a runtime bus event.
func (c *RuntimeClient) EmitEvent(ctx context.Context, eventType, sessionID, flowID string, payload map[string]any) (runtime.RuntimeEvent, error) {
	return c.store.EmitEvent(eventType, "sdk", sessionID, flowID, payload)
}

// RunFlow records flow.started / flow.completed events for a session.
func (c *RuntimeClient) RunFlow(sessionID, flowID string) error {
	_, err := c.store.EmitEvent("flow.started", "sdk", sessionID, flowID, nil)
	if err != nil {
		return err
	}
	_, err = c.store.EmitEvent("flow.completed", "sdk", sessionID, flowID, nil)
	return err
}
