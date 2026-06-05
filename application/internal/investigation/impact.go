package investigation

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/analysis"
	"github.com/google/uuid"
)

// ImpactRequest configures change impact analysis (spec-my-A §25.3).
type ImpactRequest struct {
	Flow      string
	Change    string
	ProductID string
	RepoRoot  string
}

// ImpactReport summarises blast radius for a proposed change.
type ImpactReport struct {
	ID        string        `json:"id"`
	CreatedAt time.Time     `json:"created_at"`
	Request   ImpactRequest `json:"request"`
	Affected  []string      `json:"affected_files"`
	Flows     []string      `json:"related_flows"`
	Contracts []string      `json:"related_contracts"`
	Risks     []string      `json:"risks"`
}

// RunImpact performs local impact analysis without cloud calls.
func RunImpact(ctx context.Context, req ImpactRequest) (ImpactReport, error) {
	if req.RepoRoot == "" {
		return ImpactReport{}, fmt.Errorf("impact: repo root required")
	}
	if req.ProductID == "" {
		req.ProductID = "workspace-saas"
	}
	rep := ImpactReport{
		ID:        uuid.NewString(),
		CreatedAt: time.Now().UTC(),
		Request:   req,
	}
	patterns := tokenizeSymptom(req.Change)
	if req.Flow != "" {
		rep.Flows = append(rep.Flows, req.Flow)
		rep.Contracts = append(rep.Contracts, req.Flow+".*")
	}
	local, err := Run(ctx, req.RepoRoot, req.Flow, "", nil)
	if err == nil {
		rep.Affected = append(rep.Affected, local.CandidateFiles...)
		rep.Affected = append(rep.Affected, local.RelatedTests...)
	}
	rep.Affected = dedupeKeepOrder(rep.Affected)

	if len(local.CandidateFiles) > 0 {
		g, _ := analysis.BuildDependencyGraph(req.RepoRoot, local.CandidateFiles)
		for _, n := range g.Nodes {
			if n.Kind == "file" {
				rep.Affected = append(rep.Affected, n.Name)
			}
		}
	}
	rep.Affected = dedupeKeepOrder(rep.Affected)

	contractsDir := filepath.Join(req.RepoRoot, ".asagiri", "products", req.ProductID, "contracts")
	_ = filepath.WalkDir(contractsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		content := strings.ToLower(string(raw))
		for _, p := range patterns {
			if strings.Contains(content, strings.ToLower(p)) {
				rel, _ := filepath.Rel(req.RepoRoot, path)
				rep.Contracts = append(rep.Contracts, filepath.ToSlash(rel))
				break
			}
		}
		return nil
	})

	if len(rep.Affected) > 20 {
		rep.Risks = append(rep.Risks, "large blast radius (>20 files)")
	}
	if req.Flow == "" {
		rep.Risks = append(rep.Risks, "flow not specified — scope may be incomplete")
	}
	rep.Contracts = dedupeKeepOrder(rep.Contracts)
	return rep, nil
}

// WriteImpactReport persists impact analysis under .asagiri/investigations/.
func WriteImpactReport(repoRoot string, rep ImpactReport) (string, error) {
	dir := filepath.Join(repoRoot, ".asagiri", "investigations", "impact-"+rep.ID)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	path := filepath.Join(dir, "impact-report.md")
	var b strings.Builder
	b.WriteString("# Impact Analysis\n\n")
	fmt.Fprintf(&b, "- Flow: %s\n", rep.Request.Flow)
	fmt.Fprintf(&b, "- Change: %s\n", rep.Request.Change)
	b.WriteString("\n## Affected files\n\n")
	for _, f := range rep.Affected {
		b.WriteString("- " + f + "\n")
	}
	b.WriteString("\n## Related contracts\n\n")
	for _, c := range rep.Contracts {
		b.WriteString("- " + c + "\n")
	}
	b.WriteString("\n## Risks\n\n")
	for _, r := range rep.Risks {
		b.WriteString("- " + r + "\n")
	}
	return path, os.WriteFile(path, []byte(b.String()), 0o644)
}
