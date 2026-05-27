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
	ID      string     `yaml:"id"`
	Title   string     `yaml:"title"`
	Entry   string     `yaml:"entry_screen"`
	Steps   []FlowStep `yaml:"steps"`
	Outcome string     `yaml:"outcome"`
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

type Screen struct {
	ID      string   `yaml:"id"`
	Title   string   `yaml:"title"`
	Route   string   `yaml:"route"`
	States  []string `yaml:"states,omitempty"`
	Actions []string `yaml:"actions,omitempty"`
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

