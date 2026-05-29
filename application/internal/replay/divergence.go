package replay

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// DivergenceKind classifies a replay difference (spec §15).
type DivergenceKind string

const (
	DivergenceOutput    DivergenceKind = "output"
	DivergenceTrust     DivergenceKind = "trust"
	DivergenceGraph     DivergenceKind = "graph"
	DivergenceRuntime   DivergenceKind = "runtime"
	DivergenceMissing   DivergenceKind = "missing_validation"
	DivergenceDependency DivergenceKind = "dependency"
	DivergenceKnowledge DivergenceKind = "stale_knowledge"
)

// Divergence is a detected difference between replay packages or sessions.
type Divergence struct {
	Kind     DivergenceKind `json:"kind" yaml:"kind"`
	Severity string         `json:"severity" yaml:"severity"`
	Message  string         `json:"message" yaml:"message"`
	Path     string         `json:"path,omitempty" yaml:"path,omitempty"`
}

// DetectDivergences compares two replay package directories (spec §15).
func DetectDivergences(repoRoot, replayA, replayB string) ([]Divergence, error) {
	pkgA, err := LoadPackage(repoRoot, replayA)
	if err != nil {
		return nil, err
	}
	pkgB, err := LoadPackage(repoRoot, replayB)
	if err != nil {
		return nil, err
	}
	var out []Divergence

	if pkgA.Manifest.Repo.Commit != "" && pkgB.Manifest.Repo.Commit != "" && pkgA.Manifest.Repo.Commit != pkgB.Manifest.Repo.Commit {
		out = append(out, Divergence{
			Kind:     DivergenceDependency,
			Severity: "warning",
			Message:  fmt.Sprintf("repo commit changed: %s -> %s", pkgA.Manifest.Repo.Commit, pkgB.Manifest.Repo.Commit),
		})
	}
	if pkgA.Manifest.Runtime.AsagiriVersion != pkgB.Manifest.Runtime.AsagiriVersion && pkgB.Manifest.Runtime.AsagiriVersion != "" {
		out = append(out, Divergence{
			Kind:     DivergenceRuntime,
			Severity: "warning",
			Message:  fmt.Sprintf("runtime version mismatch: %s vs %s", pkgA.Manifest.Runtime.AsagiriVersion, pkgB.Manifest.Runtime.AsagiriVersion),
		})
	}

	out = append(out, compareArtifactTrees(pkgA.Path, pkgB.Path)...)
	out = append(out, compareTrustScores(pkgA.Path, pkgB.Path)...)
	out = append(out, compareGraphNodes(pkgA.Path, pkgB.Path)...)
	out = append(out, compareMetricsJSON(pkgA.Path, pkgB.Path)...)
	out = append(out, detectStaleKnowledge(pkgA.Manifest, pkgB.Manifest)...)

	return out, nil
}

func compareArtifactTrees(dirA, dirB string) []Divergence {
	var out []Divergence
	seen := map[string]struct{}{}
	hashesA := map[string]string{}
	hashesB := map[string]string{}

	walk := func(root string, hashes map[string]string, cb func(rel string)) {
		_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return err
			}
			rel, _ := filepath.Rel(root, path)
			if rel == ManifestName || strings.HasSuffix(rel, ".gz") ||
				rel == baselineHashesRel || rel == "reports/last-session.json" {
				return nil
			}
			if !isTextArtifact(path) {
				return nil
			}
			rel = filepath.ToSlash(rel)
			if data, err := ReadMaybeCompressed(path); err == nil {
				sum := sha256.Sum256(data)
				hashes[rel] = hex.EncodeToString(sum[:])
			}
			cb(rel)
			return nil
		})
	}
	walk(dirA, hashesA, func(rel string) { seen[rel] = struct{}{} })
	walk(dirB, hashesB, func(rel string) {
		if _, ok := seen[rel]; !ok {
			out = append(out, Divergence{
				Kind:     DivergenceOutput,
				Severity: "info",
				Message:  fmt.Sprintf("replay B added artifact %s", rel),
				Path:     rel,
			})
			return
		}
		if hashesA[rel] != "" && hashesB[rel] != "" && hashesA[rel] != hashesB[rel] {
			out = append(out, Divergence{
				Kind:     DivergenceOutput,
				Severity: "warning",
				Message:  fmt.Sprintf("artefact content differs: %s", rel),
				Path:     rel,
			})
		}
		delete(seen, rel)
	})
	for rel := range seen {
		out = append(out, Divergence{
			Kind:     DivergenceOutput,
			Severity: "info",
			Message:  fmt.Sprintf("replay A has artifact missing in B: %s", rel),
			Path:     rel,
		})
	}
	return out
}

func compareMetricsJSON(dirA, dirB string) []Divergence {
	pathA := filepath.Join(dirA, "graph", "metrics.json")
	pathB := filepath.Join(dirB, "graph", "metrics.json")
	bodyA, errA := ReadMaybeCompressed(pathA)
	bodyB, errB := ReadMaybeCompressed(pathB)
	if errA != nil || errB != nil {
		return nil
	}
	if sha256.Sum256(bodyA) != sha256.Sum256(bodyB) {
		return []Divergence{{
			Kind:     DivergenceRuntime,
			Severity: "warning",
			Message:  "graph metrics.json content differs",
			Path:     "graph/metrics.json",
		}}
	}
	return nil
}

func detectStaleKnowledge(mA, mB Manifest) []Divergence {
	if mA.Source.Graph == "" || mB.Source.Graph == "" {
		return nil
	}
	if mA.Source.Graph != mB.Source.Graph {
		return []Divergence{{
			Kind:     DivergenceKnowledge,
			Severity: "warning",
			Message:  fmt.Sprintf("knowledge/graph scope changed: %s vs %s", mA.Source.Graph, mB.Source.Graph),
		}}
	}
	return nil
}

func compareTrustScores(dirA, dirB string) []Divergence {
	scoresA := trustScoresFromDir(dirA)
	scoresB := trustScoresFromDir(dirB)
	var out []Divergence
	all := map[string]struct{}{}
	for k := range scoresA {
		all[k] = struct{}{}
	}
	for k := range scoresB {
		all[k] = struct{}{}
	}
	for dim := range all {
		a, okA := scoresA[dim]
		b, okB := scoresB[dim]
		if !okA {
			out = append(out, Divergence{Kind: DivergenceMissing, Severity: "warning", Message: fmt.Sprintf("replay A skipped %s validation", dim)})
			continue
		}
		if !okB {
			out = append(out, Divergence{Kind: DivergenceMissing, Severity: "warning", Message: fmt.Sprintf("replay B skipped %s validation", dim)})
			continue
		}
		delta := b - a
		if delta < -0.01 || delta > 0.01 {
			out = append(out, Divergence{
				Kind:     DivergenceTrust,
				Severity: "warning",
				Message:  fmt.Sprintf("trust score diff %s: %+.2f", dim, delta),
			})
		}
	}
	return out
}

func trustScoresFromDir(replayDir string) map[string]float64 {
	out := map[string]float64{}
	trustRoot := filepath.Join(replayDir, "trust")
	entries, err := os.ReadDir(trustRoot)
	if err != nil {
		return out
	}
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		reportPath := filepath.Join(trustRoot, e.Name(), "report.json")
		body, err := os.ReadFile(reportPath)
		if err != nil {
			continue
		}
		var report struct {
			Checks []struct {
				Type       string  `json:"type"`
				Confidence float64 `json:"confidence"`
				Status     string  `json:"status"`
			} `json:"checks"`
		}
		if err := json.Unmarshal(body, &report); err != nil {
			continue
		}
		for _, c := range report.Checks {
			if c.Type != "" {
				out[c.Type] = c.Confidence
			}
		}
	}
	return out
}

func compareGraphNodes(dirA, dirB string) []Divergence {
	nodesA := graphNodeIDs(filepath.Join(dirA, "graph", "execution-graph.json"))
	nodesB := graphNodeIDs(filepath.Join(dirB, "graph", "execution-graph.json"))
	if len(nodesA) == 0 && len(nodesB) == 0 {
		return nil
	}
	var out []Divergence
	for id := range nodesB {
		if _, ok := nodesA[id]; !ok {
			out = append(out, Divergence{
				Kind:     DivergenceGraph,
				Severity: "info",
				Message:  fmt.Sprintf("replay B inserted node %s", id),
			})
		}
	}
	for id := range nodesA {
		if _, ok := nodesB[id]; !ok {
			out = append(out, Divergence{
				Kind:     DivergenceGraph,
				Severity: "info",
				Message:  fmt.Sprintf("replay A node %s missing in B", id),
			})
		}
	}
	return out
}

func graphNodeIDs(path string) map[string]struct{} {
	out := map[string]struct{}{}
	body, err := os.ReadFile(path)
	if err != nil {
		body, err = ReadMaybeCompressed(path)
		if err != nil {
			return out
		}
	}
	var graph struct {
		Nodes []struct {
			ID string `json:"id"`
		} `json:"nodes"`
	}
	if err := json.Unmarshal(body, &graph); err != nil {
		return out
	}
	for _, n := range graph.Nodes {
		if n.ID != "" {
			out[n.ID] = struct{}{}
		}
	}
	return out
}

// ExplainDivergences renders human-readable divergence lines (spec §6.4).
func ExplainDivergences(divs []Divergence) []string {
	lines := make([]string, 0, len(divs))
	for _, d := range divs {
		lines = append(lines, d.Message)
	}
	return lines
}
