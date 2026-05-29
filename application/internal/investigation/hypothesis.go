package investigation

import (
	"fmt"
	"strings"
)

// Hypothesis is a scored root-cause candidate (spec-my-A §25, deterministic V1).
type Hypothesis struct {
	ID          string   `json:"id"`
	Statement   string   `json:"statement"`
	Score       float64  `json:"score"`
	EvidenceIDs []string `json:"evidence_ids,omitempty"`
	Category    string   `json:"category"`
}

// Evidence is a traceable finding.
type Evidence struct {
	ID       string `json:"id"`
	Kind     string `json:"kind"` // grep, file, test, config, flow, contract
	Summary  string `json:"summary"`
	Location string `json:"location,omitempty"`
}

// GenerateHypotheses applies deterministic rules from scope and local findings.
func GenerateHypotheses(scope ResolvedScope, res InvestigationResult, graph ContextPack) ([]Hypothesis, []Evidence) {
	var evidence []Evidence
	eid := 0
	nextID := func() string {
		eid++
		return fmt.Sprintf("ev-%03d", eid)
	}
	addEv := func(kind, summary, loc string) string {
		id := nextID()
		evidence = append(evidence, Evidence{ID: id, Kind: kind, Summary: summary, Location: loc})
		return id
	}
	for _, h := range res.GrepHits {
		if len(evidence) >= 30 {
			break
		}
		addEv("grep", h, "")
	}
	for _, f := range res.CandidateFiles {
		addEv("file", "candidate source file", f)
	}
	for _, t := range res.RelatedTests {
		addEv("test", "related test", t)
	}
	if scope.Flow != "" {
		addEv("flow", "resolved flow scope: "+scope.Flow, scope.Flow)
	}
	for _, c := range scope.Contracts {
		addEv("contract", "related contract: "+c, c)
	}
	for _, api := range graph.APIs {
		addEv("contract", "knowledge graph API: "+api, api)
	}
	for _, ev := range graph.Events {
		addEv("flow", "knowledge graph event: "+ev, ev)
	}
	for _, m := range graph.Metrics {
		addEv("config", "knowledge graph metric: "+m, m)
	}
	for _, logHint := range LinkLogsToFlowEvents(scope.Instruction, graph.Events) {
		addEv("flow", logHint, scope.Flow)
	}

	var hyps []Hypothesis
	score := func(base float64, boosts ...bool) float64 {
		s := base
		for _, b := range boosts {
			if b {
				s += 0.1
			}
		}
		if s > 0.95 {
			return 0.95
		}
		return s
	}
	hasTests := len(res.RelatedTests) > 0
	hasGrep := len(res.GrepHits) > 0
	symptomLower := strings.ToLower(scope.Instruction)

	if strings.Contains(symptomLower, "invite") || scope.Action == "invite_member" {
		evIDs := filterEvidenceIDs(evidence, "flow", "contract", "grep")
		hyps = append(hyps, Hypothesis{
			ID: "hyp-invite-flow", Statement: "Invitation flow step fails (permissions or async email)",
			Score: score(0.72, hasTests, hasGrep), EvidenceIDs: evIDs, Category: "flow",
		})
	}
	if len(res.SensitivePaths) > 0 {
		id := addEv("security", "sensitive paths excluded from context", strings.Join(res.SensitivePaths, ", "))
		hyps = append(hyps, Hypothesis{
			ID: "hyp-config-secret", Statement: "Misconfiguration or missing secret (paths redacted)",
			Score: 0.45, EvidenceIDs: []string{id}, Category: "config",
		})
	}
	if len(res.Errors) > 0 {
		id := addEv("collector", "collector errors: "+strings.Join(res.Errors, "; "), "")
		hyps = append(hyps, Hypothesis{
			ID: "hyp-partial-discovery", Statement: "Investigation incomplete — some collectors failed",
			Score: 0.55, EvidenceIDs: []string{id}, Category: "meta",
		})
	}
	if len(graph.Risks) > 0 {
		evIDs := filterEvidenceIDs(evidence, "contract", "flow", "test")
		hyps = append(hyps, Hypothesis{
			ID: "hyp-graph-risk", Statement: "Knowledge graph risk: " + graph.Risks[0],
			Score: score(0.68, len(graph.APIs) > 0), EvidenceIDs: evIDs, Category: "graph",
		})
	}
	if len(hyps) == 0 && len(res.CandidateFiles) > 0 {
		evIDs := filterEvidenceIDs(evidence, "file", "grep")
		hyps = append(hyps, Hypothesis{
			ID: "hyp-general-code", Statement: "Issue likely in candidate files from local discovery",
			Score: score(0.5, hasGrep), EvidenceIDs: evIDs, Category: "code",
		})
	}
	if scope.TaskID != "" {
		hyps = append(hyps, Hypothesis{
			ID: "hyp-task-scope", Statement: "Failure scoped to task " + scope.TaskID,
			Score: 0.6, EvidenceIDs: filterEvidenceIDs(evidence, "file", "test"), Category: "task",
		})
	}
	return hyps, evidence
}

func filterEvidenceIDs(evidence []Evidence, kinds ...string) []string {
	allowed := map[string]struct{}{}
	for _, k := range kinds {
		allowed[k] = struct{}{}
	}
	var ids []string
	for _, e := range evidence {
		if _, ok := allowed[e.Kind]; ok || len(kinds) == 0 {
			ids = append(ids, e.ID)
		}
		if len(ids) >= 8 {
			break
		}
	}
	return ids
}
