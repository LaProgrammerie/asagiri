package contextopt

import (
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/investigation"
)

// PackInput carries investigation and task hints for ordering.
type PackInput struct {
	Feature            string
	TaskID             string
	TaskDescription    string
	Inv                investigation.InvestigationResult
	ReducedFiles       []FileEntry
	AcceptanceCriteria string
	OutputFormat       string
}

// BuildPack orders context sections per specv3 §8.3.
func BuildPack(cfg *config.Config, in PackInput) ContextPack {
	var hints strings.Builder
	for _, p := range in.Inv.CandidateFiles {
		hints.WriteString("- ")
		hints.WriteString(p)
		hints.WriteString("\n")
	}
	ordered := append([]FileEntry{}, in.ReducedFiles...)
	sort.SliceStable(ordered, func(i, j int) bool {
		if ordered[i].Score != ordered[j].Score {
			return ordered[i].Score > ordered[j].Score
		}
		return ordered[i].RelPath < ordered[j].RelPath
	})
	digests := make([]FileDigest, 0, len(ordered))
	for _, f := range ordered {
		digests = append(digests, FileDigest{Path: f.RelPath, Excerpt: f.Content})
	}
	invLines := []string{
		"### Investigation",
		"candidate_files: " + strings.Join(in.Inv.CandidateFiles, ", "),
		"tests: " + strings.Join(in.Inv.RelatedTests, ", "),
		"sensitive: " + strings.Join(in.Inv.SensitivePaths, ", "),
		"large_files: " + strings.Join(in.Inv.LargeFiles, ", "),
	}
	validation := cfg.ValidationCommandLines()
	purpose := in.TaskDescription
	if purpose == "" && in.TaskID != "" {
		purpose = "task " + in.TaskID
	}
	if purpose == "" {
		purpose = "feature " + in.Feature
	}
	return ContextPack{
		TaskObjective:      "### Objective\n" + purpose,
		AcceptanceCriteria: in.AcceptanceCriteria,
		FileHints:          "### File hints\n" + hints.String(),
		Investigation:      strings.Join(invLines, "\n"),
		FileExcerpts:       digests,
		ValidationLines:    validation,
		OutputFormat:       in.OutputFormat,
	}
}

// PackTokens estimates token footprint of a pack using cfg ratios (delegates to cost estimators in Optimize).
func PackApproxTokens(pack ContextPack, te config.TokenEstimationConfig) int {
	var sb strings.Builder
	sb.WriteString(pack.TaskObjective)
	sb.WriteString(pack.AcceptanceCriteria)
	sb.WriteString(pack.FileHints)
	sb.WriteString(pack.Investigation)
	for _, d := range pack.FileExcerpts {
		sb.WriteString(d.Path)
		sb.WriteString(d.Excerpt)
	}
	for _, v := range pack.ValidationLines {
		sb.WriteString(v)
	}
	sb.WriteString(pack.OutputFormat)
	// inline token formula: chars/default ratio
	chars := sb.Len()
	if te.DefaultCharsPerToken <= 0 {
		te.DefaultCharsPerToken = 4
	}
	return int(float64(chars)/te.DefaultCharsPerToken + 0.999)
}
