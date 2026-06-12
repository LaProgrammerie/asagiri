package checks

import (
	"context"
	"os"
	"path/filepath"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
	"github.com/LaProgrammerie/asagiri/application/internal/product"
)

const productsRel = ".asagiri/products"

// KnowledgeBlastRadiusFunc maps a knowledge graph impact to blast radius when graph.sqlite exists.
type KnowledgeBlastRadiusFunc func(ctx context.Context, repoRoot, flowID string) (BlastRadiusSummary, bool)

// Dependencies groups injectable collaborators for check runners.
type Dependencies struct {
	Config *config.Config

	KnowledgeBlastRadius KnowledgeBlastRadiusFunc

	Investigate      func(ctx context.Context, repoRoot, feature, taskID string, cfg *config.Config) (investigation.InvestigationResult, error)
	BuildDepGraph    func(repoRoot string, files []string) (analysis.Graph, error)
	LoadBundle       func(repoRoot, productID string) (analysis.Bundle, error)
	ParseFailedTests func(ctx context.Context, repoRoot string) ([]string, error)
	RelatedTests     func(candidates []string) []string
	ReadFile         func(path string) ([]byte, error)
}

// DefaultDependencies returns production wiring for check runners.
func DefaultDependencies() Dependencies {
	return Dependencies{
		Investigate:      investigation.Run,
		BuildDepGraph:    analysis.BuildDependencyGraph,
		LoadBundle:       analysis.LoadBundle,
		ParseFailedTests: investigation.ParseFailedTests,
		RelatedTests:     investigation.RelatedTestPaths,
		ReadFile:         os.ReadFile,
	}
}

// ProductDir returns .asagiri/products/<productID>.
func ProductDir(repoRoot, productID string) string {
	return filepath.Join(repoRoot, productsRel, product.Slug(productID))
}
