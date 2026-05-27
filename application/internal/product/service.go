package product

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

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
	product := Slug(opts.Product)
	if product == "product" {
		product = Slug(opts.Intent)
	}
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
	if err := s.repo.WriteProductFile(product, "intent.md", []byte(p.Intent+"\n")); err != nil {
		return "", err
	}
	model := s.defaultModel(product)
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
	modelPath := filepath.Join(s.repo.productRoot(product), "prototype", "model.json")
	modelRaw, err := os.ReadFile(modelPath)
	if err != nil {
		return fmt.Errorf("prototype model not found: %w", err)
	}
	var model extractedModel
	if err := json.Unmarshal(modelRaw, &model); err != nil {
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
	report := fmt.Sprintf("# Extraction report\n\n- flows: %d\n- screens: %d\n- unresolved_contracts: %d\n", len(model.Flows), len(model.Screens), model.unresolvedContracts())
	return s.repo.WriteProductFile(product, "extraction/extraction-report.md", []byte(report))
}

func (s *Service) InspectFlows(product string) (string, error) {
	product = Slug(product)
	modelPath := filepath.Join(s.repo.productRoot(product), "extraction", "extracted-model.yaml")
	data, err := os.ReadFile(modelPath)
	if err != nil {
		return "", err
	}
	var model extractedModel
	if err := yaml.Unmarshal(data, &model); err != nil {
		return "", err
	}
	steps := 0
	for _, flow := range model.Flows {
		steps += len(flow.Steps)
	}
	risk := "low"
	if model.unresolvedContracts() > 0 {
		risk = "medium"
	}
	return fmt.Sprintf("flows=%d steps=%d screens=%d unresolved_contracts=%d risk=%s", len(model.Flows), steps, len(model.Screens), model.unresolvedContracts(), risk), nil
}

func (s *Service) ExtractContracts(product string, dryRun bool) error {
	product = Slug(product)
	summary, err := s.InspectFlows(product)
	if err != nil {
		return err
	}
	if dryRun {
		return nil
	}
	if err := s.repo.WriteProductFile(product, "contracts/api.openapi.yaml", []byte(defaultOpenAPI(product))); err != nil {
		return err
	}
	for _, file := range []string{"permissions.yaml", "events.yaml", "analytics.yaml", "observability.yaml"} {
		if err := s.repo.WriteProductFile(product, filepath.Join("contracts", file), []byte("version: v1\nstatus: draft\n")); err != nil {
			return err
		}
	}
	return s.repo.WriteProductFile(product, "extraction/extraction-report.md", []byte("# Extraction report\n\n"+summary+"\n"))
}

func (s *Service) GenerateSpecFromProduct(product string, dryRun bool) error {
	product = Slug(product)
	if dryRun {
		return nil
	}
	spec := fmt.Sprintf("# %s\n\n## Requirements\n\n- Provide deterministic executable product artifacts.\n", product)
	design := "## Design\n\n- Flow-driven extraction\n- Contracts derived from actions\n"
	tasksMD := "- [ ] Extract and validate flows\n- [ ] Generate contracts and specs\n"
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
	if err := s.repo.WriteSpecsFile(product, "tasks.yaml", []byte("tasks:\n  - id: "+product+"-001\n    title: Implement core flow\n")); err != nil {
		return err
	}
	if err := s.repo.WriteSpecsFile(product, "metadata.yaml", []byte("status: ready\nsource: product-layer\n")); err != nil {
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
	task := asagiri.Task{
		ID:      product + "-001",
		Title:   "Implement " + product + " first flow",
		Feature: product,
		Status:  asagiri.StatusPending,
		Type:    "implementation",
		Source:  asagiri.TaskSource{Spec: ".kiro/specs/" + product + "/tasks.md", Section: "generated-from-product"},
		Scope:   asagiri.TaskScope{AllowedPaths: []string{"application/**", ".asagiri/products/" + product + "/**"}},
		Acceptance: []string{
			"flows extracted",
			"contracts generated",
		},
		Agents: asagiri.TaskAgents{Implementer: "cursor", Reviewer: "codex", Enricher: "ollama"},
	}
	task.TouchMetadata(time.Now().UTC())
	body, _ := EncodeYAML(task)
	if err := s.repo.WriteTaskFile(product, task.ID+".yaml", body); err != nil {
		return err
	}
	return nil
}

func (s *Service) ReviewProduct(product string, dryRun bool) (string, error) {
	product = Slug(product)
	summary, err := s.InspectFlows(product)
	if err != nil {
		return "", err
	}
	report := "# Product review\n\n- ux: coherent\n- security: verify sensitive actions\n- observability: draft only\n- summary: " + summary + "\n"
	if !dryRun {
		if err := s.repo.WriteProductFile(product, "reviews/product-review.md", []byte(report)); err != nil {
			return "", err
		}
	}
	return summary, nil
}

type extractedModel struct {
	Product string   `yaml:"product" json:"product"`
	Flows   []Flow   `yaml:"flows" json:"flows"`
	Screens []Screen `yaml:"screens" json:"screens"`
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

func (s *Service) defaultModel(product string) extractedModel {
	return extractedModel{
		Product: product,
		Flows: []Flow{
			{
				ID:    "workspace-onboarding",
				Title: "Workspace onboarding",
				Entry: "landing",
				Steps: []FlowStep{
					{ID: "step-1", Screen: "landing", Action: "click_get_started", Next: "signup", ContractRef: "POST /api/workspaces", Errors: []string{"email_taken"}},
					{ID: "step-2", Screen: "signup", Action: "submit_signup", Next: "dashboard", ContractRef: "TODO:auth.signup", Sensitive: true, Errors: []string{"weak_password"}},
				},
				Outcome: "workspace_created",
			},
		},
		Screens: []Screen{
			{ID: "landing", Title: "Landing", Route: "/", States: []string{"idle"}, Actions: []string{"click_get_started"}},
			{ID: "signup", Title: "Signup", Route: "/signup", States: []string{"idle", "submitting", "error"}, Actions: []string{"submit_signup"}},
			{ID: "dashboard", Title: "Dashboard", Route: "/dashboard", States: []string{"ready"}, Actions: []string{"invite_member"}},
		},
	}
}

func defaultString(v, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
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

