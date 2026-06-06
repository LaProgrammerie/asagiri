package product

import (
	"fmt"
	"strings"
)

// LayerStep is one product preparation step.
type LayerStep struct {
	Name    string
	Skip    bool
	Reason  string
	RelPath string
}

// LayerPlan is the ordered product pipeline for dry-run display.
func LayerPlan(state ProductArtifactState) []LayerStep {
	steps := []LayerStep{
		{Name: "Create product model", Skip: state.HasProductModel, RelPath: ".asagiri/products/" + state.ProductID + "/product.yaml"},
		{Name: "Generate prototype", Skip: state.HasPrototype, RelPath: ".asagiri/products/" + state.ProductID + "/prototype/"},
		{Name: "Extract flows", Skip: state.HasFlows, RelPath: ".asagiri/products/" + state.ProductID + "/flows/"},
		{Name: "Extract contracts", Skip: state.HasContracts, RelPath: ".asagiri/products/" + state.ProductID + "/contracts/"},
		{Name: "Generate specs", Skip: state.HasGeneratedSpecs, RelPath: ".asagiri/specs/" + state.ProductID + "/"},
		{Name: "Generate tasks", Skip: state.HasTasks, RelPath: ".asagiri/specs/" + state.ProductID + "/tasks.yaml"},
	}
	for i := range steps {
		if steps[i].Skip {
			steps[i].Reason = "already present"
		}
	}
	return steps
}

// LayerResult captures product preparation output.
type LayerResult struct {
	ProductID string
	Generated []string
	Skipped   []string
}

// PrepareLayerOptions configures product layer execution.
type PrepareLayerOptions struct {
	Intent  string
	Product string
	DryRun  bool
}

// PrepareLayer runs missing product pipeline steps via existing services.
func (s *Service) PrepareLayer(opts PrepareLayerOptions) (LayerResult, error) {
	productID := ResolveProductID(s.repo.RepoRoot, opts.Intent, opts.Product)
	res := LayerResult{ProductID: productID}

	state := InspectArtifacts(s.repo.RepoRoot, productID)
	modelPath := ".asagiri/products/" + productID + "/product.yaml"
	protoPath := ".asagiri/products/" + productID + "/prototype/"
	flowsPath := ".asagiri/products/" + productID + "/flows/"
	contractsPath := ".asagiri/products/" + productID + "/contracts/"
	specsPath := ".asagiri/specs/" + productID + "/"

	if !state.HasProductModel || !state.HasPrototype {
		if opts.DryRun {
			res.Generated = append(res.Generated, modelPath, protoPath)
		} else if _, err := s.CreatePrototype(CreatePrototypeOptions{
			Intent:  opts.Intent,
			Product: productID,
		}); err != nil {
			return res, err
		} else {
			res.Generated = append(res.Generated, modelPath, protoPath)
		}
	} else {
		res.Skipped = append(res.Skipped, modelPath, protoPath)
	}

	state = InspectArtifacts(s.repo.RepoRoot, productID)
	if !state.HasFlows {
		if opts.DryRun {
			res.Generated = append(res.Generated, flowsPath)
		} else if err := s.ExtractFlows(productID, false); err != nil {
			return res, err
		} else {
			res.Generated = append(res.Generated, flowsPath)
		}
	} else {
		res.Skipped = append(res.Skipped, flowsPath)
	}

	state = InspectArtifacts(s.repo.RepoRoot, productID)
	if !state.HasContracts {
		if opts.DryRun {
			res.Generated = append(res.Generated, contractsPath)
		} else if err := s.ExtractContracts(productID, false); err != nil {
			return res, err
		} else {
			res.Generated = append(res.Generated, contractsPath)
		}
	} else {
		res.Skipped = append(res.Skipped, contractsPath)
	}

	state = InspectArtifacts(s.repo.RepoRoot, productID)
	if !state.HasGeneratedSpecs || !state.HasTasks {
		if opts.DryRun {
			res.Generated = append(res.Generated, specsPath)
		} else if err := s.GenerateSpecFromProduct(productID, false); err != nil {
			return res, err
		} else {
			res.Generated = append(res.Generated, specsPath)
		}
	} else {
		res.Skipped = append(res.Skipped, specsPath)
	}

	return res, nil
}

// FormatLayerDryRun renders the product-layer dry-run plan.
func FormatLayerDryRun(state ProductArtifactState, planOnly bool) string {
	var b strings.Builder
	b.WriteString("Product-level intent detected\n")
	b.WriteString("\nProduct ID: ")
	b.WriteString(state.ProductID)
	b.WriteString("\n\n")
	if len(state.MissingArtifacts) == 0 {
		b.WriteString("Missing artifacts:\n- none (product layer complete)\n\n")
	} else {
		b.WriteString("Missing artifacts:\n")
		for _, m := range state.MissingArtifacts {
			b.WriteString("- ")
			b.WriteString(m)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Planned workflow:\n")
	for i, step := range LayerPlan(state) {
		line := fmt.Sprintf("%d. %s", i+1, step.Name)
		if step.Skip {
			line += " (skip: already present)"
		}
		b.WriteString(line)
		b.WriteString("\n")
	}
	if planOnly {
		b.WriteString("\n(plan-only: Product Layer plan only — the technical workflow plan is not generated for product-level intents)\n")
	}
	return b.String()
}

// FormatLayerResult renders execution summary.
func FormatLayerResult(res LayerResult) string {
	var b strings.Builder
	b.WriteString("Product preparation complete\n\n")
	if len(res.Generated) > 0 {
		b.WriteString("Generated:\n")
		for _, g := range res.Generated {
			b.WriteString("- ")
			b.WriteString(g)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	if len(res.Skipped) > 0 {
		b.WriteString("Skipped:\n")
		for _, s := range res.Skipped {
			b.WriteString("- ")
			b.WriteString(s)
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}
	b.WriteString("Next:\nasa work ")
	b.WriteString(res.ProductID)
	b.WriteString("\n")
	return b.String()
}
