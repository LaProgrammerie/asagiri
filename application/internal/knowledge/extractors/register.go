package extractors

import (
	"context"

	"github.com/LaProgrammerie/asagiri/application/internal/knowledge"
)

func init() {
	knowledge.RegisterExtractor("flows", "flows", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return FlowExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("contracts", "contracts", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return ContractExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("events", "contracts", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return EventExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("permissions", "contracts", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return PermissionExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("code", "code", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return CodeExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("tests", "tests", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return TestExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("observability", "contracts", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return ObservabilityExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterExtractor("analytics", "contracts", func(ctx context.Context, repoRoot, product string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return AnalyticsExtractor{}.Extract(ctx, repoRoot, product)
	})
	knowledge.RegisterRepoExtractor("adr", "adr", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return ADRExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("infra", "infra", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return InfraExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("runtime", "runtime", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return RuntimeExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("specs", "specs", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return SpecExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("tasks", "tasks", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return TaskExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("trust", "trust", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return TrustReportExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("investigation", "investigation", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return InvestigationReportExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterRepoExtractor("config", "config", func(ctx context.Context, repoRoot string) ([]knowledge.GraphNode, []knowledge.GraphEdge, []string, error) {
		return ConfigExtractor{}.Extract(ctx, repoRoot, "")
	})
	knowledge.RegisterFlowCodeLinker(LinkFlowToCode, WarnUntestedActions)
}
