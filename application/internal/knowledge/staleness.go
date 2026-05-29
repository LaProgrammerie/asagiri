package knowledge

import (
	"context"
	"errors"
	"fmt"
)

const recommendIncrementalBuild = "asa knowledge build --incremental"

// StalenessReport summarizes outdated graph elements (spec-my-E §21).
type StalenessReport struct {
	Stale            bool     `json:"stale"`
	FilesChanged     int      `json:"files_changed"`
	EdgesOutdated    int      `json:"edges_outdated"`
	RecommendCommand string   `json:"recommend_command,omitempty"`
	Warnings         []string `json:"warnings,omitempty"`
}

// StalenessDetector checks whether the graph needs rebuild.
type StalenessDetector interface {
	Check(ctx context.Context, repoRoot string) (StalenessReport, error)
}

// GraphStalenessDetector compares source mtimes to index_metadata last build.
type GraphStalenessDetector struct {
	OpenStore func(string) (GraphStore, error)
}

// DefaultStalenessDetector returns a detector using the registered SQLite store.
func DefaultStalenessDetector() *GraphStalenessDetector {
	return &GraphStalenessDetector{OpenStore: OpenStore}
}

// Check reports staleness when inputs changed after the last successful build.
func (d *GraphStalenessDetector) Check(ctx context.Context, repoRoot string) (StalenessReport, error) {
	if repoRoot == "" {
		return StalenessReport{}, fmt.Errorf("staleness check: repo root required")
	}
	open := d.OpenStore
	if open == nil {
		open = OpenStore
	}
	store, err := open(repoRoot)
	if err != nil {
		return StalenessReport{}, err
	}
	defer store.Close()

	meta, err := store.GetIndexMetadata(ctx, "build")
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return StalenessReport{
				Stale:            true,
				RecommendCommand: "asa knowledge build",
				Warnings:         []string{"no prior knowledge build metadata"},
			}, nil
		}
		return StalenessReport{}, err
	}

	builtAt, ok := parseBuiltAt(meta)
	if !ok {
		return StalenessReport{
			Stale:            true,
			RecommendCommand: "asa knowledge build",
			Warnings:         []string{"build metadata missing built_at"},
		}, nil
	}

	products, err := resolveProducts(repoRoot, "")
	if err != nil {
		return StalenessReport{}, err
	}

	req := buildRequestFromMeta(meta)
	currentIdx, err := ScanSourceIndex(repoRoot, products, req)
	if err != nil {
		return StalenessReport{}, err
	}

	stored := storedSourceMTimes(meta)
	var changed int
	if len(stored) > 0 {
		for cat, unix := range currentIdx.CategoryMTimes {
			if !categoryUnchanged(stored, cat, unix) {
				changed++
			}
		}
	} else {
		changed, err = CountFilesChangedSince(repoRoot, products, req, builtAt)
		if err != nil {
			return StalenessReport{}, err
		}
	}

	report := StalenessReport{FilesChanged: changed}
	if changed > 0 {
		report.Stale = true
		report.EdgesOutdated = estimateOutdatedEdges(changed)
		report.RecommendCommand = recommendIncrementalBuild
	}
	return report, nil
}

func estimateOutdatedEdges(filesChanged int) int {
	if filesChanged <= 1 {
		return filesChanged
	}
	return filesChanged / 2
}

func buildRequestFromMeta(meta map[string]any) BuildRequest {
	req := BuildRequest{
		IncludeFlows:     true,
		IncludeContracts: true,
	}
	if scope, ok := meta["scope"].(string); ok {
		req.Scope = scope
	}
	if v, ok := meta["include_flows"].(bool); ok {
		req.IncludeFlows = v
	}
	if v, ok := meta["include_contracts"].(bool); ok {
		req.IncludeContracts = v
	}
	if v, ok := meta["include_code"].(bool); ok {
		req.IncludeCode = v
	}
	if v, ok := meta["include_tests"].(bool); ok {
		req.IncludeTests = v
	}
	if v, ok := meta["include_infra"].(bool); ok {
		req.IncludeInfra = v
	}
	if v, ok := meta["include_adr"].(bool); ok {
		req.IncludeADR = v
	}
	if v, ok := meta["include_runtime"].(bool); ok {
		req.IncludeRuntime = v
	}
	return req
}

// StubStalenessDetector is a deprecated placeholder.
type StubStalenessDetector struct{}

func (StubStalenessDetector) Check(_ context.Context, _ string) (StalenessReport, error) {
	return StalenessReport{}, ErrNotImplemented
}
