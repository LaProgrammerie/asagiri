package investigation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// FormatContextPackMarkdown renders agent context per spec-my-A §25.14.
func FormatContextPackMarkdown(rep Report, pack ContextPack) string {
	var b strings.Builder
	b.WriteString("# Context Pack\n\n")
	b.WriteString("## Problem\n\n")
	b.WriteString(rep.Request.Symptom + "\n\n")
	b.WriteString("## Scope\n\n")
	fmt.Fprintf(&b, "- Flow: %s\n", rep.Scope.Flow)
	fmt.Fprintf(&b, "- Step: %s\n", rep.Scope.Step)
	fmt.Fprintf(&b, "- Action: %s\n", rep.Scope.Action)
	if len(rep.Scope.Contracts) > 0 {
		b.WriteString("\n## Related contracts\n\n")
		for _, c := range rep.Scope.Contracts {
			b.WriteString("- " + c + "\n")
		}
	}
	if rep.Scope.Flow != "" {
		b.WriteString("\n## Related flows\n\n")
		b.WriteString("- " + rep.Scope.Flow + "\n")
	}
	b.WriteString("\n## Key files\n\n")
	for _, f := range pack.Files {
		b.WriteString("- `" + f + "`\n")
	}
	if len(pack.Tests) > 0 {
		b.WriteString("\n## Tests\n\n")
		for _, t := range pack.Tests {
			b.WriteString("- `" + t + "`\n")
		}
	}
	if len(pack.APIs) > 0 {
		b.WriteString("\n## API operations\n\n")
		for _, api := range pack.APIs {
			b.WriteString("- " + api + "\n")
		}
	}
	if len(pack.Events) > 0 {
		b.WriteString("\n## Events\n\n")
		for _, ev := range pack.Events {
			b.WriteString("- " + ev + "\n")
		}
	}
	if len(pack.Metrics) > 0 {
		b.WriteString("\n## Metrics\n\n")
		for _, m := range pack.Metrics {
			b.WriteString("- " + m + "\n")
		}
	}
	if len(pack.Risks) > 0 {
		b.WriteString("\n## Risks\n\n")
		for _, r := range pack.Risks {
			b.WriteString("- " + r + "\n")
		}
	}
	if len(rep.Hypotheses) > 0 {
		b.WriteString("\n## Hypotheses (scored)\n\n")
		for _, h := range rep.Hypotheses {
			fmt.Fprintf(&b, "- (%.2f) %s\n", h.Score, h.Statement)
		}
	}
	b.WriteString("\n## Suggested commands\n\n")
	for _, a := range rep.SuggestedActions {
		b.WriteString("- " + a + "\n")
	}
	b.WriteString("\n## Success criteria\n\n")
	b.WriteString("- Reproduce and fix the scoped failure\n")
	b.WriteString("- Run targeted tests from the list above\n")
	b.WriteString("\n## Restrictions\n\n")
	b.WriteString("- No secrets or `.env` files\n")
	b.WriteString("- Stay within scoped modules\n")
	if len(pack.ExcludedSensitive) > 0 {
		for _, p := range pack.ExcludedSensitive {
			b.WriteString("- excluded: `" + p + "`\n")
		}
	}
	return b.String()
}

// WriteContextPackArtifacts saves JSON and Markdown context packs.
func WriteContextPackArtifacts(dir string, rep Report, pack ContextPack) (jsonPath, mdPath string, err error) {
	jsonPath, err = WriteContextPack(dir, pack)
	if err != nil {
		return "", "", err
	}
	mdPath = filepath.Join(dir, "context-pack.md")
	if err := os.WriteFile(mdPath, []byte(FormatContextPackMarkdown(rep, pack)), 0o644); err != nil {
		return jsonPath, "", err
	}
	return jsonPath, mdPath, nil
}
