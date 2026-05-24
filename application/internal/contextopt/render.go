package contextopt

import (
	"fmt"
	"strings"
)

// RenderPackMarkdown renders a ContextPack as markdown for CLI --show.
func RenderPackMarkdown(p ContextPack) string {
	var sb strings.Builder
	sb.WriteString(p.TaskObjective)
	sb.WriteString("\n\n")
	if strings.TrimSpace(p.AcceptanceCriteria) != "" {
		sb.WriteString(p.AcceptanceCriteria)
		sb.WriteString("\n\n")
	}
	sb.WriteString(p.FileHints)
	sb.WriteString("\n\n")
	sb.WriteString(p.Investigation)
	sb.WriteString("\n\n## Fichiers\n")
	for _, d := range p.FileExcerpts {
		sb.WriteString(fmt.Sprintf("### %s\n```\n%s\n```\n", d.Path, strings.TrimSpace(d.Excerpt)))
	}
	if len(p.ValidationLines) > 0 {
		sb.WriteString("\n## Validation\n")
		for _, v := range p.ValidationLines {
			sb.WriteString("- ")
			sb.WriteString(v)
			sb.WriteString("\n")
		}
	}
	if p.OutputFormat != "" {
		sb.WriteString("\n## Sortie\n")
		sb.WriteString(p.OutputFormat)
	}
	return sb.String()
}
