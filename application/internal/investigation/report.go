package investigation

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/json"
)

// Report is the full investigation output (spec-my-A §25).
type Report struct {
	ID                  string              `json:"id"`
	CreatedAt           time.Time           `json:"created_at"`
	Request             Request             `json:"request"`
	Scope               ResolvedScope       `json:"scope"`
	Evidence            []Evidence          `json:"evidence"`
	Hypotheses          []Hypothesis        `json:"hypotheses"`
	RootCauseCandidates []Hypothesis        `json:"root_cause_candidates"`
	SuggestedActions    []string            `json:"suggested_actions"`
	ContextPackPath     string              `json:"context_pack_path,omitempty"`
	ReplayPackPath      string              `json:"replay_pack_path,omitempty"`
	LocalResult         InvestigationResult `json:"local_result"`
	EstimateTokens      int                 `json:"estimate_tokens,omitempty"`
	EstimateCostEUR     float64             `json:"estimate_cost_eur,omitempty"`
	Limits              []string            `json:"limits,omitempty"`
	Risks               []string            `json:"risks,omitempty"`
}

// WriteReport persists markdown and JSON artefacts under .asagiri/investigations/.
func WriteReport(repoRoot string, rep Report) (string, error) {
	dir := filepath.Join(repoRoot, ".asagiri", "investigations", rep.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	mdPath := filepath.Join(dir, "report.md")
	jsonPath := filepath.Join(dir, "report.json")
	if err := os.WriteFile(mdPath, []byte(formatReportMarkdown(rep)), 0o644); err != nil {
		return "", err
	}
	raw, _ := json.MarshalIndent(rep, "", "  ")
	if err := os.WriteFile(jsonPath, raw, 0o644); err != nil {
		return "", err
	}
	return mdPath, nil
}

func formatReportMarkdown(rep Report) string {
	var b strings.Builder
	b.WriteString("# Investigation Report\n\n")
	b.WriteString("## Input\n\n")
	fmt.Fprintf(&b, "- Symptom: %s\n", rep.Request.Symptom)
	fmt.Fprintf(&b, "- Feature: %s\n", rep.Request.Feature)
	fmt.Fprintf(&b, "- Depth: %s\n", rep.Request.Depth)
	b.WriteString("\n## Resolved Scope\n\n")
	fmt.Fprintf(&b, "- Flow: %s\n", rep.Scope.Flow)
	fmt.Fprintf(&b, "- Step: %s\n", rep.Scope.Step)
	fmt.Fprintf(&b, "- Action: %s\n", rep.Scope.Action)
	if len(rep.Scope.LikelyModules) > 0 {
		b.WriteString("- Likely modules:\n")
		for _, m := range rep.Scope.LikelyModules {
			fmt.Fprintf(&b, "  - %s\n", m)
		}
	}
	if len(rep.Scope.Contracts) > 0 {
		b.WriteString("- Contracts:\n")
		for _, c := range rep.Scope.Contracts {
			fmt.Fprintf(&b, "  - %s\n", c)
		}
	}
	b.WriteString("\n## Evidence Collected\n\n")
	for _, e := range rep.Evidence {
		fmt.Fprintf(&b, "- **%s** (%s): %s", e.ID, e.Kind, e.Summary)
		if e.Location != "" {
			fmt.Fprintf(&b, " — `%s`", e.Location)
		}
		b.WriteString("\n")
	}
	b.WriteString("\n## Related Flows\n\n")
	if rep.Scope.Flow != "" {
		fmt.Fprintf(&b, "- %s\n", rep.Scope.Flow)
	} else {
		b.WriteString("- (none resolved)\n")
	}
	b.WriteString("\n## Related Contracts\n\n")
	if len(rep.Scope.Contracts) == 0 {
		b.WriteString("- (none)\n")
	} else {
		for _, c := range rep.Scope.Contracts {
			fmt.Fprintf(&b, "- %s\n", c)
		}
	}
	b.WriteString("\n## Hypotheses\n\n")
	for _, h := range rep.Hypotheses {
		fmt.Fprintf(&b, "- **%s** (score %.2f, %s): %s\n", h.ID, h.Score, h.Category, h.Statement)
	}
	b.WriteString("\n## Root Cause Candidates\n\n")
	for _, h := range rep.RootCauseCandidates {
		fmt.Fprintf(&b, "- **%s** (score %.2f): %s\n", h.ID, h.Score, h.Statement)
	}
	b.WriteString("\n## Suggested Next Actions\n\n")
	for _, a := range rep.SuggestedActions {
		fmt.Fprintf(&b, "- %s\n", a)
	}
	b.WriteString("\n## Context Pack\n\n")
	if rep.ContextPackPath != "" {
		fmt.Fprintf(&b, "Path: `%s`\n", rep.ContextPackPath)
	} else {
		b.WriteString("- (not generated)\n")
	}
	b.WriteString("\n## Risks\n\n")
	for _, r := range rep.Risks {
		fmt.Fprintf(&b, "- %s\n", r)
	}
	b.WriteString("\n## Limits\n\n")
	for _, l := range rep.Limits {
		fmt.Fprintf(&b, "- %s\n", l)
	}
	return b.String()
}
