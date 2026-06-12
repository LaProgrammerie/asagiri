package product

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

type Product struct {
	Name   string `yaml:"name"`
	Intent string `yaml:"intent"`
	Stack  string `yaml:"stack"`
	Style  string `yaml:"style"`
}

type Flow struct {
	ID                       string          `yaml:"id"`
	Title                    string          `yaml:"title"`
	Entry                    string          `yaml:"entry_screen"`
	Steps                    []FlowStep      `yaml:"steps"`
	Outcome                  string          `yaml:"outcome"`
	Business                 FlowBusiness    `yaml:"business,omitempty"`
	Metrics                  []string        `yaml:"metrics,omitempty"`
	ArchitectureImplications []string        `yaml:"architecture_implications,omitempty"`
	Observability            FlowTelemetry   `yaml:"observability,omitempty"`
	Security                 FlowSecurity    `yaml:"security,omitempty"`
	CostProfile              FlowCostProfile `yaml:"cost_profile,omitempty"`
}

type FlowStep struct {
	ID          string   `yaml:"id"`
	Screen      string   `yaml:"screen"`
	Action      string   `yaml:"action"`
	Next        string   `yaml:"next"`
	ContractRef string   `yaml:"contract_ref,omitempty"`
	Sensitive   bool     `yaml:"sensitive,omitempty"`
	Errors      []string `yaml:"errors,omitempty"`
}

type FlowBusiness struct {
	Objective          string `yaml:"objective,omitempty"`
	Criticality        string `yaml:"criticality,omitempty"`
	MonetizationImpact string `yaml:"monetization_impact,omitempty"`
}

type FlowTelemetry struct {
	Traces  []string `yaml:"traces,omitempty"`
	Metrics []string `yaml:"metrics,omitempty"`
	Logs    []string `yaml:"logs,omitempty"`
}

type FlowSecurity struct {
	RequiresAuthentication bool     `yaml:"requires_authentication,omitempty"`
	SensitiveActions       []string `yaml:"sensitive_actions,omitempty"`
}

type FlowCostProfile struct {
	ExpectedComplexity     string `yaml:"expected_complexity,omitempty"`
	InfrastructureCostRisk string `yaml:"infrastructure_cost_risk,omitempty"`
}

type Screen struct {
	ID      string   `yaml:"id"`
	Title   string   `yaml:"title"`
	Route   string   `yaml:"route"`
	States  []string `yaml:"states,omitempty"`
	Actions []string `yaml:"actions,omitempty"`
}

type BusinessIntent struct {
	Objective struct {
		Primary   string   `yaml:"primary"`
		Secondary []string `yaml:"secondary,omitempty"`
	} `yaml:"objective"`
	TargetUsers    []string         `yaml:"target_users,omitempty"`
	SuccessMetrics []BusinessMetric `yaml:"success_metrics,omitempty"`
	Constraints    []string         `yaml:"constraints,omitempty"`
	BusinessRisk   struct {
		Level   string   `yaml:"level,omitempty"`
		Reasons []string `yaml:"reasons,omitempty"`
	} `yaml:"business_risk,omitempty"`
	Monetization struct {
		Model           string `yaml:"model,omitempty"`
		ActivationEvent string `yaml:"activation_event,omitempty"`
	} `yaml:"monetization,omitempty"`
	ObservabilityRequirements []string `yaml:"observability_requirements,omitempty"`
}

type BusinessMetric struct {
	ID     string `yaml:"id"`
	Target string `yaml:"target"`
}

func ParseProductYAML(data []byte) (Product, error) {
	var p Product
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Product{}, fmt.Errorf("parse product.yaml: %w", err)
	}
	return p, ValidateProduct(p)
}

func ParseFlowYAML(data []byte) (Flow, error) {
	var f Flow
	if err := yaml.Unmarshal(data, &f); err != nil {
		return Flow{}, fmt.Errorf("parse flow yaml: %w", err)
	}
	return f, ValidateFlow(f)
}

func ParseScreenYAML(data []byte) (Screen, error) {
	var s Screen
	if err := yaml.Unmarshal(data, &s); err != nil {
		return Screen{}, fmt.Errorf("parse screen yaml: %w", err)
	}
	return s, ValidateScreen(s)
}

func ParseBusinessYAML(data []byte) (BusinessIntent, error) {
	var b BusinessIntent
	if err := yaml.Unmarshal(data, &b); err != nil {
		return BusinessIntent{}, fmt.Errorf("parse business yaml: %w", err)
	}
	return b, ValidateBusiness(b)
}

func ValidateProduct(p Product) error {
	if strings.TrimSpace(p.Name) == "" {
		return fmt.Errorf("product.name is required")
	}
	if strings.TrimSpace(p.Intent) == "" {
		return fmt.Errorf("product.intent is required")
	}
	if p.Stack == "" {
		p.Stack = "react"
	}
	if p.Stack != "react" {
		return fmt.Errorf("product.stack must be react (v1)")
	}
	return nil
}

func ValidateFlow(f Flow) error {
	if strings.TrimSpace(f.ID) == "" {
		return fmt.Errorf("flow.id is required")
	}
	if strings.TrimSpace(f.Entry) == "" {
		return fmt.Errorf("flow.entry_screen is required")
	}
	if len(f.Steps) == 0 {
		return fmt.Errorf("flow.steps must not be empty")
	}
	for i, step := range f.Steps {
		if strings.TrimSpace(step.ID) == "" {
			return fmt.Errorf("flow.steps[%d].id is required", i)
		}
		if strings.TrimSpace(step.Screen) == "" {
			return fmt.Errorf("flow.steps[%d].screen is required", i)
		}
		if strings.TrimSpace(step.Action) == "" {
			return fmt.Errorf("flow.steps[%d].action is required", i)
		}
		if step.Sensitive && len(step.Errors) == 0 {
			return fmt.Errorf("flow.steps[%d] sensitive action requires errors", i)
		}
	}
	if f.Business.Criticality == "high" && len(f.Metrics) == 0 {
		return fmt.Errorf("flow.metrics is required when business.criticality is high")
	}
	return nil
}

func ValidateScreen(s Screen) error {
	if strings.TrimSpace(s.ID) == "" {
		return fmt.Errorf("screen.id is required")
	}
	if strings.TrimSpace(s.Title) == "" {
		return fmt.Errorf("screen.title is required")
	}
	if strings.TrimSpace(s.Route) == "" {
		return fmt.Errorf("screen.route is required")
	}
	return nil
}

func ValidateBusiness(b BusinessIntent) error {
	if strings.TrimSpace(b.Objective.Primary) == "" {
		return fmt.Errorf("business.objective.primary is required")
	}
	for i, metric := range b.SuccessMetrics {
		if strings.TrimSpace(metric.ID) == "" {
			return fmt.Errorf("business.success_metrics[%d].id is required", i)
		}
	}
	return nil
}
