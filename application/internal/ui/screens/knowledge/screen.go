package knowledge

import (
	"fmt"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/ui/bus"
)

// ViewModel contains knowledge explorer data.
type ViewModel struct {
	Search  bus.KnowledgeSearchResult
	ShowCLI bool
}

// Render returns knowledge explorer content.
func Render(vm ViewModel) string {
	query := strings.TrimSpace(vm.Search.Query)
	if query == "" {
		query = "onboarding"
	}
	var b strings.Builder
	b.WriteString(fmt.Sprintf("Knowledge search: %s\n", query))
	if vm.Search.Warning != "" {
		b.WriteString("Warning: " + vm.Search.Warning + "\n")
	}
	if len(vm.Search.Matches) == 0 {
		b.WriteString("- no match")
		return b.String()
	}
	for i, match := range vm.Search.Matches {
		if i >= 8 {
			break
		}
		b.WriteString(fmt.Sprintf("- %s [%s] %s\n", match.ID, match.Type, match.Name))
		b.WriteString(fmt.Sprintf("  Path: %s  Score: %.0f%%\n", match.Path, match.Score*100))
		if vm.ShowCLI && match.CLIEquivalent != "" {
			b.WriteString("  CLI: " + match.CLIEquivalent + "\n")
		}
	}
	return strings.TrimRight(b.String(), "\n")
}
