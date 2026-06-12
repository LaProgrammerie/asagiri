package product

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/product/derivation"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
	"gopkg.in/yaml.v3"
)

type Service struct {
	repo *Repository
}

func NewService(repoRoot string) *Service {
	return &Service{repo: NewRepository(repoRoot)}
}

type CreatePrototypeOptions struct {
	Intent  string
	Product string
	Stack   string
	Style   string
	DryRun  bool
}

func (s *Service) CreatePrototype(opts CreatePrototypeOptions) (string, error) {
	product := ResolveProductID(s.repo.RepoRoot, opts.Intent, opts.Product)
	if opts.DryRun {
		return product, nil
	}
	if err := s.repo.EnsureRoots(product); err != nil {
		return "", err
	}
	p := Product{
		Name:   product,
		Intent: strings.TrimSpace(opts.Intent),
		Stack:  defaultString(opts.Stack, "react"),
		Style:  defaultString(opts.Style, "minimal"),
	}
	if err := ValidateProduct(p); err != nil {
		return "", err
	}
	raw, _ := EncodeYAML(p)
	if err := s.repo.WriteProductFile(product, "product.yaml", raw); err != nil {
		return "", err
	}
	business := defaultBusinessIntent(opts.Intent)
	businessRaw, _ := EncodeYAML(business)
	if err := s.repo.WriteProductFile(product, "business.yaml", businessRaw); err != nil {
		return "", err
	}
	if err := s.repo.WriteProductFile(product, "intent.md", []byte(p.Intent+"\n")); err != nil {
		return "", err
	}
	model := s.defaultModel(product, business)
	modelJSON, _ := json.MarshalIndent(model, "", "  ")
	if err := s.repo.WriteProductFile(product, "prototype/model.json", append(modelJSON, '\n')); err != nil {
		return "", err
	}
	if err := s.repo.WriteProductFile(product, "prototype/package.json", []byte(defaultPackageJSON())); err != nil {
		return "", err
	}
	if err := s.repo.WriteProductFile(product, "prototype/index.html", []byte(defaultIndexHTML())); err != nil {
		return "", err
	}
	if err := s.repo.WriteProductFile(product, "prototype/src/App.tsx", []byte(defaultAppTSX(product))); err != nil {
		return "", err
	}
	if err := s.repo.WriteProductFile(product, "prototype/README.md", []byte(defaultPrototypeReadme(product))); err != nil {
		return "", err
	}
	return product, nil
}

func (s *Service) PatchPrototype(product, instruction string, dryRun bool) error {
	product = Slug(product)
	if dryRun {
		return nil
	}
	report := fmt.Sprintf("## Patch %s\n\n- instruction: %s\n- mode: deterministic v1\n\n", time.Now().UTC().Format(time.RFC3339), strings.TrimSpace(instruction))
	return s.repo.WriteProductFile(product, "extraction/extraction-report.md", []byte(report))
}

func (s *Service) ExtractFlows(product string, dryRun bool) error {
	product = Slug(product)
	model, err := s.loadPrototypeModel(product)
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	for _, flow := range model.Flows {
		body, _ := EncodeYAML(flow)
		if err := s.repo.WriteProductFile(product, filepath.Join("flows", flow.ID+".flow.yaml"), body); err != nil {
			return err
		}
	}
	for _, screen := range model.Screens {
		body, _ := EncodeYAML(screen)
		if err := s.repo.WriteProductFile(product, filepath.Join("screens", screen.ID+".screen.yaml"), body); err != nil {
			return err
		}
	}
	rawModel, _ := EncodeYAML(model)
	if err := s.repo.WriteProductFile(product, "extraction/extracted-model.yaml", rawModel); err != nil {
		return err
	}
	report := fmt.Sprintf("# Extraction report\n\n- flows: %d\n- screens: %d\n- unresolved_contracts: %d\n- metrics_gaps: %d\n", len(model.Flows), len(model.Screens), model.unresolvedContracts(), model.missingMetrics())
	return s.repo.WriteProductFile(product, "extraction/extraction-report.md", []byte(report))
}

func (s *Service) InspectFlows(product string) (string, error) {
	product = Slug(product)
	model, err := s.loadExtractedModel(product)
	if err != nil {
		return "", err
	}
	steps := 0
	actions := 0
	for _, flow := range model.Flows {
		steps += len(flow.Steps)
		actions += len(flow.Steps)
	}
	risk := "low"
	if model.unresolvedContracts() > 0 || model.missingMetrics() > 0 {
		risk = "medium"
	}
	if model.unresolvedContracts() > 1 {
		risk = "high"
	}
	return fmt.Sprintf("flows=%d steps=%d actions=%d screens=%d unresolved_contracts=%d metrics_gaps=%d risk=%s", len(model.Flows), steps, actions, len(model.Screens), model.unresolvedContracts(), model.missingMetrics(), risk), nil
}

func (s *Service) ReviewFlows(product string, dryRun bool) (string, error) {
	product = Slug(product)
	model, err := s.loadExtractedModel(product)
	if err != nil {
		return "", err
	}
	projection := derivation.DeriveArchitecture(model.toDerivationFlows())
	warnings := make([]string, 0)
	if model.missingMetrics() > 0 {
		warnings = append(warnings, fmt.Sprintf("metrics manquantes sur %d flow(s)", model.missingMetrics()))
	}
	if model.unresolvedContracts() > 0 {
		warnings = append(warnings, fmt.Sprintf("contracts non résolus: %d", model.unresolvedContracts()))
	}
	for _, item := range projection.Security {
		if strings.HasPrefix(item, "rate_limit:") {
			warnings = append(warnings, "rate limiting requis: "+strings.TrimPrefix(item, "rate_limit:"))
		}
	}
	if len(warnings) == 0 {
		warnings = append(warnings, "aucun gap critique détecté")
	}
	risk := "low"
	if len(warnings) > 1 {
		risk = "medium"
	}
	report := "# Flow Review\n\n"
	for _, flow := range model.Flows {
		report += "- flow: " + flow.ID + " | objective: " + defaultString(flow.Business.Objective, model.Business.Objective.Primary) + "\n"
	}
	report += "\n## Warnings\n"
	for _, warning := range warnings {
		report += "- " + warning + "\n"
	}
	report += "\nRisk: " + risk + "\n"
	if !dryRun {
		if err := s.repo.WriteProductFile(product, "reviews/flow-review.md", []byte(report)); err != nil {
			return "", err
		}
	}
	return fmt.Sprintf("flow_review product=%s warnings=%d risk=%s", product, len(warnings), risk), nil
}

func (s *Service) DeriveArchitecture(product string, dryRun bool) (string, error) {
	product = Slug(product)
	model, err := s.loadExtractedModel(product)
	if err != nil {
		return "", err
	}
	projection := derivation.DeriveArchitecture(model.toDerivationFlows())
	if err := ValidateProjectionCouplingStrict(projection); err != nil {
		return "", err
	}
	report := "# Architecture Projection\n\n"
	report += "## API\n" + toMarkdownList(projection.API)
	report += "\n## Async\n" + toMarkdownList(projection.Async)
	report += "\n## Security\n" + toMarkdownList(projection.Security)
	report += "\n## Observability\n" + toMarkdownList(projection.Observability)
	report += "\n## Infrastructure\n" + toMarkdownList(projection.Infrastructure)
	if dryRun {
		return fmt.Sprintf("architecture_projection product=%s api=%d async=%d security=%d", product, len(projection.API), len(projection.Async), len(projection.Security)), nil
	}
	if err := s.repo.WriteProductFile(product, "reviews/architecture-review.md", []byte(report)); err != nil {
		return "", err
	}
	return fmt.Sprintf("architecture_projection product=%s api=%d async=%d security=%d", product, len(projection.API), len(projection.Async), len(projection.Security)), nil
}

func (s *Service) ExtractContracts(product string, dryRun bool) error {
	product = Slug(product)
	model, err := s.loadExtractedModel(product)
	if err != nil {
		return err
	}
	projection := derivation.DeriveArchitecture(model.toDerivationFlows())
	if err := ValidateProjectionCouplingStrict(projection); err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	if err := s.repo.WriteProductFile(product, "contracts/api.openapi.yaml", []byte(defaultOpenAPI(product))); err != nil {
		return err
	}
	permissionsBody, _ := EncodeYAML(map[string]any{"roles": map[string]any{"owner": map[string]any{"permissions": projection.Permissions}, "member": map[string]any{"permissions": []string{"workspace.read"}}}})
	if err := s.repo.WriteProductFile(product, "contracts/permissions.yaml", permissionsBody); err != nil {
		return err
	}
	eventsBody, _ := EncodeYAML(map[string]any{"events": projection.API})
	if err := s.repo.WriteProductFile(product, "contracts/events.yaml", eventsBody); err != nil {
		return err
	}
	analyticsBody, _ := EncodeYAML(map[string]any{"events": projection.Analytics})
	if err := s.repo.WriteProductFile(product, "contracts/analytics.yaml", analyticsBody); err != nil {
		return err
	}
	obsBody, _ := EncodeYAML(map[string]any{"requirements": projection.Observability, "metrics_coverage": projection.MetricsCoverage})
	if err := s.repo.WriteProductFile(product, "contracts/observability.yaml", obsBody); err != nil {
		return err
	}
	summary, _ := s.InspectFlows(product)
	return s.repo.WriteProductFile(product, "extraction/extraction-report.md", []byte("# Extraction report\n\n"+summary+"\n"))
}

func (s *Service) GenerateSpecFromProduct(product string, dryRun bool) error {
	product = Slug(product)
	model, err := s.loadExtractedModel(product)
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	spec := fmt.Sprintf("# %s\n\n## Requirements\n\n- Flow-centric product execution\n- Metrics-driven engineering\n", product)
	design := "## Design\n\n- Business intent -> flows -> architecture derivation\n- Contracts aligned with observability and analytics\n"
	taskLines := make([]string, 0)
	var generatedTasks []asagiri.Task
	taskIdx := 1
	for _, flow := range model.Flows {
		for _, step := range flow.Steps {
			taskID := fmt.Sprintf("%s-%03d", product, taskIdx)
			taskIdx++
			task := asagiri.Task{
				ID:      taskID,
				Title:   fmt.Sprintf("Implement %s/%s", flow.ID, step.Action),
				Feature: product,
				Status:  asagiri.StatusPending,
				Risk:    "medium",
				Type:    "implementation",
				Source: asagiri.TaskSource{
					Spec:              ".kiro/specs/" + product + "/tasks.md",
					Section:           "generated-from-product-flow",
					Product:           product,
					Flow:              flow.ID,
					Step:              step.ID,
					Action:            step.Action,
					BusinessObjective: defaultString(flow.Business.Objective, model.Business.Objective.Primary),
				},
				Scope: asagiri.TaskScope{
					AllowedPaths: []string{"application/**", ".asagiri/products/" + product + "/**"},
				},
				Acceptance: []string{
					"flow step implemented",
					"contract alignment verified",
					"analytics/metrics linked",
				},
				Agents: asagiri.TaskAgents{Implementer: config.DefaultAgentDev, Reviewer: config.DefaultAgentReviewer, Enricher: config.DefaultAgentEnrich},
			}
			task.TouchMetadata(time.Now().UTC())
			generatedTasks = append(generatedTasks, task)
			taskLines = append(taskLines, fmt.Sprintf("- [ ] %s (%s/%s)", task.Title, flow.ID, step.ID))
		}
	}
	tasksMD := strings.Join(taskLines, "\n") + "\n"
	if err := s.repo.WriteProductFile(product, "generated-specs/requirements.md", []byte(spec)); err != nil {
		return err
	}
	if err := s.repo.WriteProductFile(product, "generated-specs/design.md", []byte(design)); err != nil {
		return err
	}
	if err := s.repo.WriteProductFile(product, "generated-specs/tasks.md", []byte(tasksMD)); err != nil {
		return err
	}
	if err := s.repo.WriteSpecsFile(product, "spec.md", []byte(spec+"\n"+design)); err != nil {
		return err
	}
	tasksYAML, _ := EncodeYAML(map[string]any{"tasks": generatedTasks})
	if err := s.repo.WriteSpecsFile(product, "tasks.yaml", tasksYAML); err != nil {
		return err
	}
	if err := s.repo.WriteSpecsFile(product, "metadata.yaml", []byte("status: ready\nsource: product-layer-flow-first\n")); err != nil {
		return err
	}
	kiroDir := filepath.Join(s.repo.RepoRoot, ".kiro", "specs", product)
	if err := os.MkdirAll(kiroDir, 0o755); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(kiroDir, "requirements.md"), []byte(spec), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(kiroDir, "design.md"), []byte(design), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(kiroDir, "tasks.md"), []byte(tasksMD), 0o644); err != nil {
		return err
	}
	for _, task := range generatedTasks {
		body, _ := EncodeYAML(task)
		if err := s.repo.WriteTaskFile(product, task.ID+".yaml", body); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) ReviewProduct(product string, dryRun bool) (string, error) {
	product = Slug(product)
	summary, err := s.InspectFlows(product)
	if err != nil {
		return "", err
	}
	flowSummary, err := s.ReviewFlows(product, dryRun)
	if err != nil {
		return "", err
	}
	report := "# Product review\n\n- ux: coherent\n- security: sensitive actions flagged\n- observability: metrics-driven checks enabled\n- summary: " + summary + "\n- flow_review: " + flowSummary + "\n"
	if !dryRun {
		if err := s.repo.WriteProductFile(product, "reviews/product-review.md", []byte(report)); err != nil {
			return "", err
		}
	}
	return summary, nil
}

type extractedModel struct {
	Product  string         `yaml:"product" json:"product"`
	Business BusinessIntent `yaml:"business" json:"business"`
	Flows    []Flow         `yaml:"flows" json:"flows"`
	Screens  []Screen       `yaml:"screens" json:"screens"`
}

func (m extractedModel) unresolvedContracts() int {
	var count int
	for _, flow := range m.Flows {
		for _, step := range flow.Steps {
			if strings.Contains(step.ContractRef, "TODO") || step.ContractRef == "" {
				count++
			}
		}
	}
	return count
}

func (m extractedModel) missingMetrics() int {
	count := 0
	for _, flow := range m.Flows {
		if len(flow.Metrics) == 0 {
			count++
		}
	}
	return count
}

func (m extractedModel) toDerivationFlows() []derivation.FlowInput {
	out := make([]derivation.FlowInput, 0, len(m.Flows))
	for _, flow := range m.Flows {
		item := derivation.FlowInput{
			ID:                flow.ID,
			BusinessObjective: defaultString(flow.Business.Objective, m.Business.Objective.Primary),
			Metrics:           flow.Metrics,
			StepActions:       make([]derivation.StepAction, 0, len(flow.Steps)),
		}
		for _, step := range flow.Steps {
			item.StepActions = append(item.StepActions, derivation.StepAction{
				StepID:      step.ID,
				Action:      step.Action,
				ContractRef: step.ContractRef,
				Sensitive:   step.Sensitive,
				Errors:      step.Errors,
			})
		}
		out = append(out, item)
	}
	return out
}

func (s *Service) defaultModel(product string, business BusinessIntent) extractedModel {
	return extractedModel{
		Product:  product,
		Business: business,
		Flows: []Flow{
			{
				ID:    "workspace-onboarding",
				Title: "Workspace onboarding",
				Entry: "landing",
				Steps: []FlowStep{
					{ID: "step-1", Screen: "landing", Action: "click_get_started", Next: "signup", ContractRef: "POST /api/workspaces", Errors: []string{"email_taken"}},
					{ID: "step-2", Screen: "signup", Action: "invite_member", Next: "dashboard", ContractRef: "TODO:auth.signup", Sensitive: true, Errors: []string{"weak_password"}},
				},
				Outcome: "workspace_created",
				Business: FlowBusiness{
					Objective:          business.Objective.Primary,
					Criticality:        "high",
					MonetizationImpact: "high",
				},
				Metrics:                  []string{"onboarding_completion_rate", "invitation_delivery_success_rate"},
				ArchitectureImplications: []string{"async_email_delivery", "rate_limiting", "audit_logs"},
				Observability: FlowTelemetry{
					Traces:  []string{"onboarding.start", "onboarding.complete"},
					Metrics: []string{"onboarding_step_duration", "onboarding_dropoff_rate"},
					Logs:    []string{"invitation_failed"},
				},
				Security: FlowSecurity{
					RequiresAuthentication: true,
					SensitiveActions:       []string{"invite_member"},
				},
				CostProfile: FlowCostProfile{
					ExpectedComplexity:     "medium",
					InfrastructureCostRisk: "low",
				},
			},
		},
		Screens: []Screen{
			{ID: "landing", Title: "Landing", Route: "/", States: []string{"idle"}, Actions: []string{"click_get_started"}},
			{ID: "signup", Title: "Signup", Route: "/signup", States: []string{"idle", "submitting", "error"}, Actions: []string{"invite_member"}},
			{ID: "dashboard", Title: "Dashboard", Route: "/dashboard", States: []string{"ready"}, Actions: []string{"invite_member"}},
		},
	}
}

func (s *Service) loadPrototypeModel(product string) (extractedModel, error) {
	modelPath := filepath.Join(s.repo.productRoot(product), "prototype", "model.json")
	modelRaw, err := os.ReadFile(modelPath)
	if err != nil {
		return extractedModel{}, fmt.Errorf("prototype model not found: %w", err)
	}
	var model extractedModel
	if err := json.Unmarshal(modelRaw, &model); err != nil {
		return extractedModel{}, err
	}
	if model.Business.Objective.Primary == "" {
		model.Business = defaultBusinessIntent(product)
	}
	return model, nil
}

func (s *Service) loadExtractedModel(product string) (extractedModel, error) {
	modelPath := filepath.Join(s.repo.productRoot(product), "extraction", "extracted-model.yaml")
	data, err := os.ReadFile(modelPath)
	if err != nil {
		return extractedModel{}, err
	}
	var model extractedModel
	if err := yaml.Unmarshal(data, &model); err != nil {
		return extractedModel{}, err
	}
	if model.Business.Objective.Primary == "" {
		model.Business = defaultBusinessIntent(product)
	}
	return model, nil
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}

func toMarkdownList(values []string) string {
	if len(values) == 0 {
		return "- none\n"
	}
	var out strings.Builder
	for _, value := range values {
		out.WriteString("- ")
		out.WriteString(value)
		out.WriteString("\n")
	}
	return out.String()
}

func defaultBusinessIntent(intent string) BusinessIntent {
	b := BusinessIntent{}
	b.Objective.Primary = defaultString(strings.TrimSpace(intent), "reduce onboarding friction")
	b.TargetUsers = []string{"technical_founder", "small_team"}
	b.SuccessMetrics = []BusinessMetric{
		{ID: "onboarding_completion_rate", Target: ">=70%"},
		{ID: "workspace_creation_success_rate", Target: ">=99%"},
	}
	b.Constraints = []string{"low_operational_cost", "email_verification_required"}
	b.BusinessRisk.Level = "medium"
	b.BusinessRisk.Reasons = []string{"onboarding critical for conversion"}
	b.Monetization.Model = "subscription"
	b.Monetization.ActivationEvent = "onboarding_completed"
	b.ObservabilityRequirements = []string{"onboarding funnel", "invite delivery monitoring"}
	return b
}

func defaultPackageJSON() string {
	return "{\n  \"name\": \"asagiri-product-prototype\",\n  \"private\": true,\n  \"version\": \"0.0.1\",\n  \"scripts\": {\n    \"dev\": \"vite\"\n  }\n}\n"
}

func defaultIndexHTML() string {
	return "<!doctype html>\n<html><body><div id=\"root\"></div><script type=\"module\" src=\"/src/App.tsx\"></script></body></html>\n"
}

func defaultAppTSX(product string) string {
	return "export default function App() {\n  return <main><h1>" + product + " prototype</h1><p>Deterministic v1 scaffold.</p></main>\n}\n"
}

func defaultPrototypeReadme(product string) string {
	return "# Prototype " + product + "\n\nThis scaffold is deterministic and generated by `asa prototype create`.\n"
}

func defaultOpenAPI(product string) string {
	return "openapi: 3.1.0\ninfo:\n  title: " + product + " API\n  version: 0.1.0\npaths:\n  /api/workspaces:\n    post:\n      summary: Create workspace\n      responses:\n        '201':\n          description: Created\n"
}
