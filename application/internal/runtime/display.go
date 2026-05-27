package runtime

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// StatusView is enriched status for rich terminal UX (spec-my-A §24.20).
type StatusView struct {
	DaemonStatus
	Metrics       MetricsSnapshot
	SessionName   string
	BranchName    string
	FlowID        string
	FlowSteps     []FlowStepView
	RecentEvents  []RuntimeEvent
}

// FlowStepView is one step line in the terminal dashboard.
type FlowStepView struct {
	ID     string
	Label  string
	Status string // done, active, pending
}

// BuildStatusView collects data for rich display.
func (s *Store) BuildStatusView() (StatusView, error) {
	st, err := s.Status()
	if err != nil {
		return StatusView{}, err
	}
	metrics, _ := s.CollectMetrics()
	view := StatusView{DaemonStatus: st, Metrics: metrics}
	sessions, _ := s.ListSessions()
	if len(sessions) > 0 {
		view.SessionName = sessions[0].Name
		view.FlowID = sessions[0].FlowID
		view.BranchName = sessions[0].BranchID
		if view.FlowID != "" {
			view.FlowSteps = loadFlowSteps(s.repoRoot, sessions[0].ProductID, view.FlowID)
		}
	}
	view.RecentEvents, _ = s.ListEvents(8)
	return view, nil
}

// FormatStatusRich renders spec-my-A §24.20 terminal target.
func FormatStatusRich(v StatusView) string {
	var b strings.Builder
	b.WriteString("Asagiri Runtime\n")
	b.WriteString("════════════════\n")
	if v.SessionName != "" {
		b.WriteString(fmt.Sprintf("Session: %s\n", v.SessionName))
	}
	if v.BranchName != "" {
		b.WriteString(fmt.Sprintf("Branch:  %s\n", v.BranchName))
	}
	if v.FlowID != "" {
		b.WriteString(fmt.Sprintf("Flow:    %s\n", v.FlowID))
	}
	b.WriteString("\nRuntime\n")
	b.WriteString("───────\n")
	b.WriteString(fmt.Sprintf("Workers active:        %d\n", v.Metrics.WorkersActive))
	b.WriteString(fmt.Sprintf("Queued events:         %d\n", v.QueuedEvents))
	b.WriteString(fmt.Sprintf("Memory hits:           %.0f%%\n", v.Metrics.MemoryHits*100))
	b.WriteString(fmt.Sprintf("Context reduction:     %.0f%%\n", v.Metrics.ContextReductionRatio*100))
	if len(v.FlowSteps) > 0 {
		b.WriteString("\nFlows\n")
		b.WriteString("─────\n")
		for _, step := range v.FlowSteps {
			icon := "○"
			switch step.Status {
			case "done":
				icon = "✓"
			case "active":
				icon = "⠋"
			}
			b.WriteString(fmt.Sprintf("%s %s\n", icon, step.Label))
		}
	}
	if len(v.RecentEvents) > 0 {
		b.WriteString("\nRecent events\n")
		b.WriteString("─────────────\n")
		for _, e := range v.RecentEvents {
			b.WriteString(e.Type + "\n")
		}
	}
	return b.String()
}

func loadFlowSteps(repoRoot, productID, flowID string) []FlowStepView {
	if productID == "" || flowID == "" {
		return nil
	}
	dir := filepath.Join(repoRoot, ".asagiri", "products", productID, "flows")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}
	var raw struct {
		Steps []struct {
			ID     string `yaml:"id"`
			Action string `yaml:"action"`
		} `yaml:"steps"`
	}
	for _, ent := range entries {
		if ent.IsDir() || !strings.HasSuffix(ent.Name(), ".flow.yaml") {
			continue
		}
		if !strings.Contains(ent.Name(), flowID) && flowID != "" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(dir, ent.Name()))
		if err != nil {
			continue
		}
		if yaml.Unmarshal(b, &raw) != nil {
			continue
		}
		break
	}
	if len(raw.Steps) == 0 {
		return []FlowStepView{{ID: "start", Label: flowID + ".start", Status: "active"}}
	}
	out := make([]FlowStepView, 0, len(raw.Steps))
	for i, s := range raw.Steps {
		label := s.Action
		if label == "" {
			label = s.ID
		}
		st := "pending"
		if i == 0 {
			st = "done"
		} else if i == 1 {
			st = "active"
		}
		out = append(out, FlowStepView{ID: s.ID, Label: label, Status: st})
	}
	return out
}
