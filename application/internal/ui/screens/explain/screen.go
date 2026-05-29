package explain

import (
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains explain screen data.
type ViewModel struct {
	Explain bus.ExplainResult
	ShowCLI bool
}

// Render returns explainability panel content.
func Render(vm ViewModel) string {
	var b strings.Builder
	b.WriteString("Explain\n")
	b.WriteString("Question: " + value(vm.Explain.Question, value(vm.Explain.Subject, "current decision")) + "\n")
	b.WriteString("Subject: " + value(vm.Explain.Subject, "current decision") + "\n")
	if vm.Explain.Warning != "" {
		b.WriteString("Warning: " + vm.Explain.Warning + "\n")
	}
	b.WriteString("\nSupported questions\n")
	if len(vm.Explain.SupportedQuestions) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, question := range vm.Explain.SupportedQuestions {
			b.WriteString("- " + question + "\n")
		}
	}
	b.WriteString("\nReasons\n")
	if len(vm.Explain.Reasons) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, reason := range vm.Explain.Reasons {
			b.WriteString("- " + reason + "\n")
		}
	}
	b.WriteString("\nEvidence\n")
	if len(vm.Explain.Evidence) == 0 {
		b.WriteString("- none\n")
	} else {
		for _, evidence := range vm.Explain.Evidence {
			b.WriteString("- " + evidence + "\n")
		}
	}
	b.WriteString("\nSource\n")
	b.WriteString("- " + value(vm.Explain.Source, "query-bus read-only") + "\n")
	b.WriteString("\nAlternatives\n")
	if len(vm.Explain.Alternatives) == 0 {
		b.WriteString("- none")
	} else {
		for _, alternative := range vm.Explain.Alternatives {
			b.WriteString("- " + alternative + "\n")
		}
	}
	if vm.ShowCLI && vm.Explain.CLIEquivalent != "" {
		b.WriteString("\nCLI equivalent\n")
		b.WriteString(vm.Explain.CLIEquivalent)
	}
	return strings.TrimRight(b.String(), "\n")
}

func value(v string, fallback string) string {
	if strings.TrimSpace(v) == "" {
		return fallback
	}
	return v
}
