package prototype

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains prototype split view data.
type ViewModel struct {
	Pipeline bus.PrototypePipelineResult
	ShowCLI  bool
}

// Render returns prototype wireframe + flow extraction split.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Prototype Mode\n")
	b.WriteString("Product: " + value(vm.Pipeline.Product, "-") + "\n")
	b.WriteString("Stage: " + value(vm.Pipeline.PipelineStage, "wireframe") + "\n")
	if len(vm.Pipeline.StagesDone) > 0 {
		b.WriteString("Done: " + strings.Join(vm.Pipeline.StagesDone, " -> ") + "\n")
	}
	if vm.Pipeline.Warning != "" {
		b.WriteString("Warning: " + vm.Pipeline.Warning + "\n")
	}
	b.WriteString("\nWireframe\n")
	b.WriteString(renderWireframe(vm.Pipeline))
	b.WriteString("\n\nFlow extraction\n")
	if len(vm.Pipeline.FlowExtraction) == 0 {
		b.WriteString("- none")
	} else {
		for i, step := range vm.Pipeline.FlowExtraction {
			if i >= 5 {
				break
			}
			b.WriteString(fmt.Sprintf("- %s/%s action=%s screen=%s next=%s\n", value(step.FlowID, "-"), value(step.StepID, "-"), value(step.Action, "-"), value(step.Screen, "-"), value(step.Next, "-")))
			b.WriteString(fmt.Sprintf("  contract=%s trust=%s metric=%s\n", value(step.Contract, "TODO"), value(step.Trust, "pending"), value(step.Metric, "-")))
		}
	}
	b.WriteString("\nPipeline actions\n")
	for _, action := range vm.Pipeline.SuggestedActions {
		b.WriteString("- " + action + "\n")
	}
	if vm.ShowCLI {
		b.WriteString("- open prototype: asa prototype")
	}
	return strings.TrimRight(b.String(), "\n")
}

func renderWireframe(p bus.PrototypePipelineResult) string {
	title := value(p.WireframeTitle, p.Product)
	path := value(p.WireframePath, ".asagiri/products/<product>/prototype/src/App.tsx")
	return strings.Join([]string{
		"Title: " + title,
		"Path: " + path,
		"Email    [____________]",
		"Password [____________]",
		"[ Sign in ]",
	}, "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
