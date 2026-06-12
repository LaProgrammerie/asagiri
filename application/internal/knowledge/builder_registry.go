package knowledge

import "context"

// ProductExtractorFunc extracts graph elements for one product.
type ProductExtractorFunc func(ctx context.Context, repoRoot, product string) ([]GraphNode, []GraphEdge, []string, error)

type extractorRegistration struct {
	name     string
	category string
	fn       ProductExtractorFunc
}

var registeredExtractors []extractorRegistration

type repoExtractorRegistration struct {
	name     string
	category string
	fn       RepoExtractorFunc
}

// RepoExtractorFunc extracts repo-wide graph elements (ADR, infra, runtime).
type RepoExtractorFunc func(ctx context.Context, repoRoot string) ([]GraphNode, []GraphEdge, []string, error)

var registeredRepoExtractors []repoExtractorRegistration

// RegisterExtractor adds a product-scoped extractor (called from extractors init).
func RegisterExtractor(name, category string, fn ProductExtractorFunc) {
	registeredExtractors = append(registeredExtractors, extractorRegistration{
		name:     name,
		category: category,
		fn:       fn,
	})
}

// RegisterRepoExtractor adds a repository-scoped extractor.
func RegisterRepoExtractor(name, category string, fn RepoExtractorFunc) {
	registeredRepoExtractors = append(registeredRepoExtractors, repoExtractorRegistration{
		name:     name,
		category: category,
		fn:       fn,
	})
}

func runExtractors(ctx context.Context, repoRoot, product string, categories map[string]bool, repoOnce *repoOnceState, skipCategories map[string]bool) ([]GraphNode, []GraphEdge, []string, error) {
	var nodes []GraphNode
	var edges []GraphEdge
	var warnings []string
	for _, reg := range registeredExtractors {
		if len(categories) > 0 && !categories[reg.category] {
			continue
		}
		if skipCategories[reg.category] {
			continue
		}
		if reg.category == "code" || reg.category == "tests" {
			if repoOnce != nil {
				if reg.category == "code" && repoOnce.codeDone {
					continue
				}
				if reg.category == "tests" && repoOnce.testsDone {
					continue
				}
			}
		}
		n, e, w, err := reg.fn(ctx, repoRoot, product)
		if err != nil {
			return nil, nil, nil, err
		}
		nodes = append(nodes, n...)
		edges = append(edges, e...)
		warnings = append(warnings, w...)
		if repoOnce != nil {
			if reg.category == "code" {
				repoOnce.codeDone = true
			}
			if reg.category == "tests" {
				repoOnce.testsDone = true
			}
		}
	}
	return nodes, edges, warnings, nil
}

func runRepoExtractors(ctx context.Context, repoRoot string, categories map[string]bool, skipCategories map[string]bool) ([]GraphNode, []GraphEdge, []string, error) {
	var nodes []GraphNode
	var edges []GraphEdge
	var warnings []string
	for _, reg := range registeredRepoExtractors {
		if len(categories) > 0 && !categories[reg.category] {
			continue
		}
		if skipCategories[reg.category] {
			continue
		}
		n, e, w, err := reg.fn(ctx, repoRoot)
		if err != nil {
			return nil, nil, nil, err
		}
		nodes = append(nodes, n...)
		edges = append(edges, e...)
		warnings = append(warnings, w...)
	}
	return nodes, edges, warnings, nil
}

type repoOnceState struct {
	codeDone  bool
	testsDone bool
}

type flowCodeLinker func(nodes []GraphNode, edges []GraphEdge) ([]GraphNode, []GraphEdge, []string)

type untestedActionWarner func(nodes []GraphNode, edges []GraphEdge) []string

var (
	flowCodeLinkerHook       flowCodeLinker
	untestedActionWarnerHook untestedActionWarner
)

// RegisterFlowCodeLinker wires flow-to-code linking from the extractors package.
func RegisterFlowCodeLinker(link flowCodeLinker, warn untestedActionWarner) {
	flowCodeLinkerHook = link
	untestedActionWarnerHook = warn
}

func linkFlowToCode(nodes []GraphNode, edges []GraphEdge) ([]GraphNode, []GraphEdge, []string) {
	if flowCodeLinkerHook == nil {
		return nodes, edges, nil
	}
	return flowCodeLinkerHook(nodes, edges)
}

func warnUntestedActions(nodes []GraphNode, edges []GraphEdge) []string {
	if untestedActionWarnerHook == nil {
		return nil
	}
	return untestedActionWarnerHook(nodes, edges)
}
