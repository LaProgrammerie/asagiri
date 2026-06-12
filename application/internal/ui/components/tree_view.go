package components

import (
	"fmt"
	"strings"
)

// TreeItem is one selectable row in a tree view.
type TreeItem struct {
	ID       string
	Label    string
	Meta     string
	Children []TreeItem
}

// TreeViewModel configures tree rendering.
type TreeViewModel struct {
	Title    string
	Items    []TreeItem
	Cursor   int
	Focused  bool
	Expanded map[string]bool
}

// RenderTreeView renders a flat tree with cursor selection.
func RenderTreeView(vm TreeViewModel) string {
	if vm.Title != "" {
		return vm.Title + "\n" + renderTreeItems(vm.Items, vm.Cursor, vm.Focused, 0)
	}
	return strings.TrimRight(renderTreeItems(vm.Items, vm.Cursor, vm.Focused, 0), "\n")
}

func renderTreeItems(items []TreeItem, cursor int, focused bool, depth int) string {
	var b strings.Builder
	idx := 0
	for _, item := range items {
		prefix := strings.Repeat("  ", depth)
		marker := " "
		if focused && idx == cursor {
			marker = ">"
		}
		line := fmt.Sprintf("%s%s %s", prefix, marker, item.Label)
		if item.Meta != "" {
			line += " " + item.Meta
		}
		b.WriteString(line + "\n")
		idx++
		if len(item.Children) > 0 {
			b.WriteString(renderTreeItems(item.Children, cursor, focused, depth+1))
		}
	}
	return b.String()
}
