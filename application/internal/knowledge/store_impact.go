package knowledge

import (
	"context"
	"fmt"
)

// StoreImpactAnalyzer analyzes impact using a persisted knowledge graph.
type StoreImpactAnalyzer struct {
	RepoRoot  string
	OpenStore func(string) (GraphStore, error)
}

// NewStoreImpactAnalyzer returns an analyzer for repoRoot.
func NewStoreImpactAnalyzer(repoRoot string) *StoreImpactAnalyzer {
	return &StoreImpactAnalyzer{
		RepoRoot:  repoRoot,
		OpenStore: OpenStore,
	}
}

// Analyze implements ImpactAnalyzer.
func (a *StoreImpactAnalyzer) Analyze(ctx context.Context, req ImpactRequest) (ImpactResult, error) {
	if a.RepoRoot == "" {
		return ImpactResult{}, fmt.Errorf("impact analyze: repo root required")
	}
	open := a.OpenStore
	if open == nil {
		open = OpenStore
	}
	store, err := open(a.RepoRoot)
	if err != nil {
		return ImpactResult{}, err
	}
	defer func() { _ = store.Close() }()
	return NewImpactAnalyzer(store).Analyze(ctx, req)
}
