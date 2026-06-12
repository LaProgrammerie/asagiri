package components

import (
	"strings"
)

// TableRow is one selectable table row.
type TableRow struct {
	Cells []string
}

// TableViewModel configures table rendering.
type TableViewModel struct {
	Title   string
	Headers []string
	Rows    []TableRow
	Cursor  int
	Focused bool
}

// RenderTableView renders a simple fixed-width table with selection.
func RenderTableView(vm TableViewModel) string {
	var b strings.Builder
	if vm.Title != "" {
		b.WriteString(vm.Title + "\n")
	}
	if len(vm.Headers) > 0 {
		b.WriteString(strings.Join(vm.Headers, "  ") + "\n")
		b.WriteString(strings.Repeat("─", len(strings.Join(vm.Headers, "  "))) + "\n")
	}
	if len(vm.Rows) == 0 {
		b.WriteString("- none")
		return strings.TrimRight(b.String(), "\n")
	}
	for i, row := range vm.Rows {
		prefix := " "
		if vm.Focused && i == vm.Cursor {
			prefix = ">"
		}
		b.WriteString(prefix + " " + strings.Join(row.Cells, "  ") + "\n")
	}
	return strings.TrimRight(b.String(), "\n")
}
