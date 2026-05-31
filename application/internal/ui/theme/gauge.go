package theme

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ringTemplate is a rounded (octagonal) donut outline. Each non-space cell is
// part of the ring perimeter; the percentage is drawn in the hollow center.
var ringTemplate = []string{
	"   ╭─────╮   ",
	"  ╱       ╲  ",
	" ╱         ╲ ",
	"│           │",
	" ╲         ╱ ",
	"  ╲       ╱  ",
	"   ╰─────╯   ",
}

// ringPerimeter lists ring cells (row, col) in clockwise order from the top.
var ringPerimeter = [][2]int{
	{0, 3}, {0, 4}, {0, 5}, {0, 6}, {0, 7}, {0, 8}, {0, 9},
	{1, 10}, {2, 11},
	{3, 12},
	{4, 11}, {5, 10},
	{6, 9}, {6, 8}, {6, 7}, {6, 6}, {6, 5}, {6, 4}, {6, 3},
	{5, 2}, {4, 1},
	{3, 0},
	{2, 1}, {1, 2},
}

// RenderRingGauge draws a circular donut gauge with the percentage centered.
// The arc fills clockwise from the top, proportional to pct.
func (s Styles) RenderRingGauge(pct int) string {
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}

	grid := make([][]rune, len(ringTemplate))
	for r, line := range ringTemplate {
		grid[r] = []rune(line)
	}

	filled := (pct * len(ringPerimeter)) / 100
	filledSet := make(map[[2]int]bool, filled)
	for i := 0; i < filled && i < len(ringPerimeter); i++ {
		filledSet[ringPerimeter[i]] = true
	}

	const centerRow = 3
	txt := fmt.Sprintf("%d%%", pct)
	tr := []rune(txt)
	width := len(grid[centerRow])
	start := (width - len(tr)) / 2
	center := make(map[int]rune, len(tr))
	for i, ch := range tr {
		center[start+i] = ch
	}

	primary := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(s.Theme.Palette.Primary))
	muted := lipgloss.NewStyle().Foreground(lipgloss.Color(s.Theme.Palette.Border))

	var b strings.Builder
	for r, row := range grid {
		for c, ch := range row {
			if r == centerRow {
				if tc, ok := center[c]; ok {
					b.WriteString(s.PanelTitle.Render(string(tc)))
					continue
				}
			}
			if ch == ' ' {
				b.WriteString(" ")
				continue
			}
			if filledSet[[2]int{r, c}] {
				b.WriteString(primary.Render(string(ch)))
			} else {
				b.WriteString(muted.Render(string(ch)))
			}
		}
		if r < len(grid)-1 {
			b.WriteByte('\n')
		}
	}
	return b.String()
}
