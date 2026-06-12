package agentexternal

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentcontext"
	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
)

// RenderProviderMarkdown builds the minimal provider profile document (Markdown + frontmatter).
func RenderProviderMarkdown(spec agentspec.Spec, externalKind string) string {
	prompt := renderExportPrompt(spec)
	contentHash := spec.ContentHash
	kind := strings.TrimSpace(externalKind)
	if kind == "" && spec.External != nil {
		kind = strings.TrimSpace(spec.External.Kind)
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("asagiri: true\n")
	_, _ = fmt.Fprintf(&b, "agent_id: %s\n", spec.ID)
	_, _ = fmt.Fprintf(&b, "agent_version: %q\n", spec.Version)
	_, _ = fmt.Fprintf(&b, "content_hash: %s\n", contentHash)
	if kind != "" {
		_, _ = fmt.Fprintf(&b, "external_kind: %s\n", kind)
	}
	b.WriteString("generated_by: asagiri\n")
	b.WriteString("---\n\n")

	_, _ = fmt.Fprintf(&b, "# Agent %s v%s\n\n", spec.ID, spec.Version)

	b.WriteString("## Output contract\n\n")
	_, _ = fmt.Fprintf(&b, "- format: %s\n", spec.OutputContract.Format)
	if len(spec.OutputContract.RequiredFields) > 0 {
		fields := append([]string(nil), spec.OutputContract.RequiredFields...)
		sort.Strings(fields)
		_, _ = fmt.Fprintf(&b, "- required_fields: %s\n", strings.Join(fields, ", "))
	}
	b.WriteString("\n")

	b.WriteString("## Orchestrated prompt\n\n")
	b.WriteString(strings.TrimSpace(prompt))
	b.WriteString("\n")
	return b.String()
}

func renderExportPrompt(spec agentspec.Spec) string {
	ctx := agentcontext.Build(agentcontext.Input{
		Spec:           spec,
		Feature:        "external-export",
		TaskID:         "external-export",
		RunID:          "external-export",
		Phase:          "export",
		UserTaskPrompt: "Export provider profile — aucune tâche workflow associée.",
	})
	return agentcontext.RenderPrompt(ctx)
}

func contentHashBytes(data []byte) string {
	return hashBytes(data)
}
