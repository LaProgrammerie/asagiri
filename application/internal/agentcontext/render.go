package agentcontext

import (
	"fmt"
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/agentspec"
)

const orchestratedPreamble = "Tu es exécuté par Asagiri en mode orchestré. Suis strictement le scope, le handoff et les gates actives."

// RenderPrompt produces a deterministic final prompt for the agent subprocess.
func RenderPrompt(ctx ExecutionContext) string {
	var b strings.Builder

	b.WriteString(orchestratedPreamble)
	b.WriteString("\n\n")

	writeSection(&b, "Agent", []string{
		fmt.Sprintf("id: %s", ctx.AgentID),
		fmt.Sprintf("role: %s", ctx.AgentRole),
		fmt.Sprintf("version: %s", ctx.AgentVersion),
		fmt.Sprintf("mode: %s", ctx.Mode),
	})
	writeSection(&b, "Run", []string{
		fmt.Sprintf("feature: %s", ctx.Feature),
		fmt.Sprintf("task_id: %s", ctx.TaskID),
		fmt.Sprintf("run_id: %s", ctx.RunID),
		fmt.Sprintf("phase: %s", ctx.Phase),
	})

	if ctx.SystemPrompt != "" {
		b.WriteString("## System prompt\n\n")
		b.WriteString(strings.TrimSpace(ctx.SystemPrompt))
		b.WriteString("\n\n")
	}

	writeBulletSection(&b, "Instructions", ctx.Instructions)
	writeBulletSection(&b, "Constraints", ctx.Constraints)

	writeSection(&b, "Scope strict", scopeLines(ctx))
	writeSection(&b, "Gates et handoff", []string{
		"Respecte les gates work déjà évaluées sur la tâche.",
		"Ne contourne pas human_review, verify_evidence, enrich, governance ni trust.",
		"Aligne-toi sur docs/ai/active/handoff.md et la spec Kiro active.",
	})

	writeOutputContract(&b, ctx.OutputContract)

	if len(ctx.ContextFiles) > 0 {
		writeBulletSection(&b, "Context files", sortedCopy(ctx.ContextFiles))
	}
	if len(ctx.References) > 0 {
		writeBulletSection(&b, "References", sortedCopy(ctx.References))
	}

	if ctx.UserTaskPrompt != "" {
		b.WriteString("## Task prompt\n\n")
		b.WriteString(strings.TrimSpace(ctx.UserTaskPrompt))
		b.WriteString("\n")
	}

	return strings.TrimRight(b.String(), "\n") + "\n"
}

func scopeLines(ctx ExecutionContext) []string {
	lines := []string{
		"Ne modifie aucun fichier hors périmètre autorisé.",
	}
	if len(ctx.AllowedPaths) > 0 {
		lines = append(lines, "Allowed paths:")
		for _, p := range sortedCopy(ctx.AllowedPaths) {
			lines = append(lines, "  - "+p)
		}
	}
	if len(ctx.ForbiddenPaths) > 0 {
		lines = append(lines, "Forbidden paths:")
		for _, p := range sortedCopy(ctx.ForbiddenPaths) {
			lines = append(lines, "  - "+p)
		}
	}
	return lines
}

func writeOutputContract(b *strings.Builder, contract agentspec.OutputContract) {
	b.WriteString("## Output contract\n\n")
	_, _ = fmt.Fprintf(b, "format: %s\n", contract.Format)
	if len(contract.RequiredFields) > 0 {
		b.WriteString("required_fields:\n")
		for _, f := range sortedCopy(contract.RequiredFields) {
			b.WriteString("  - " + f + "\n")
		}
	}
	b.WriteString("\n")
	b.WriteString("Produis uniquement la sortie attendue ; pas de prose hors format si un format structuré est requis.\n\n")
}

func writeSection(b *strings.Builder, title string, lines []string) {
	if len(lines) == 0 {
		return
	}
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, line := range lines {
		b.WriteString(line)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func writeBulletSection(b *strings.Builder, title string, items []string) {
	if len(items) == 0 {
		return
	}
	b.WriteString("## ")
	b.WriteString(title)
	b.WriteString("\n\n")
	for _, item := range sortedCopy(items) {
		b.WriteString("- ")
		b.WriteString(item)
		b.WriteString("\n")
	}
	b.WriteString("\n")
}

func sortedCopy(items []string) []string {
	out := append([]string(nil), items...)
	sortStrings(out)
	return out
}

func sortStrings(items []string) {
	sort.Strings(items)
}
