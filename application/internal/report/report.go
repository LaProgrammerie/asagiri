package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

type Step struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
}

type RunReport struct {
	RunID      string           `json:"run_id"`
	Feature    string           `json:"feature"`
	Status     string           `json:"status"`
	Generated  string           `json:"generated_at"`
	Steps      []Step           `json:"steps"`
	Tasks      []sqlite.Task    `json:"tasks"`
	Repository string           `json:"repository"`
	Cost       *CostPerformance `json:"cost_performance,omitempty"`
}

// CostPerformance is an optional V3 section (specv3 §15).
type CostPerformance struct {
	EstimatedInputTokens    int     `json:"estimated_input_tokens"`
	EstimatedOutputTokens   int     `json:"estimated_output_tokens"`
	ActualInputTokens       int     `json:"actual_input_tokens"`
	ActualOutputTokens      int     `json:"actual_output_tokens"`
	EstimatedCost           string  `json:"estimated_cost"`
	ActualCost              string  `json:"actual_cost"`
	EstimatedDuration       string  `json:"estimated_duration"`
	ActualDuration          string  `json:"actual_duration"`
	FilesScanned            int     `json:"files_scanned_local"`
	CandidateFiles          int     `json:"candidate_files"`
	LargeFilesSummarized    int     `json:"large_files_summarized"`
	CloudContextReducedFrom int     `json:"cloud_context_reduced_from_tokens"`
	TokenSavingsPercent     float64 `json:"token_savings_percent"`
}

type Writer struct {
	RepoRoot string
}

func NewWriter(repoRoot string) *Writer {
	return &Writer{RepoRoot: repoRoot}
}

// Write persists run report artefacts; optional cost section when cost is non-nil.
func (w *Writer) Write(run sqlite.Run, tasks []sqlite.Task, steps []Step, cost *CostPerformance) (string, string, error) {
	runDir := filepath.Join(w.RepoRoot, ".asagiri", "runs", run.ID)
	if err := os.MkdirAll(runDir, 0o755); err != nil {
		return "", "", fmt.Errorf("create run report dir: %w", err)
	}

	payload := RunReport{
		RunID:      run.ID,
		Feature:    run.Feature,
		Status:     run.Status,
		Generated:  time.Now().UTC().Format(time.RFC3339Nano),
		Steps:      steps,
		Tasks:      tasks,
		Repository: w.RepoRoot,
		Cost:       cost,
	}

	jsonPath := filepath.Join(runDir, "report.json")
	body, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("marshal report json: %w", err)
	}
	if err := os.WriteFile(jsonPath, body, 0o644); err != nil {
		return "", "", fmt.Errorf("write report json: %w", err)
	}

	mdPath := filepath.Join(runDir, "report.md")
	md := toMarkdown(payload)
	if err := os.WriteFile(mdPath, []byte(md), 0o644); err != nil {
		return "", "", fmt.Errorf("write report markdown: %w", err)
	}

	return mdPath, jsonPath, nil
}

func toMarkdown(r RunReport) string {
	var sb strings.Builder
	sb.WriteString("# Asagiri Report\n\n")
	fmt.Fprintf(&sb, "- Run: `%s`\n", r.RunID)
	fmt.Fprintf(&sb, "- Feature: `%s`\n", r.Feature)
	fmt.Fprintf(&sb, "- Status: `%s`\n", r.Status)
	fmt.Fprintf(&sb, "- Generated: `%s`\n\n", r.Generated)

	sb.WriteString("## Steps\n\n")
	if len(r.Steps) == 0 {
		sb.WriteString("- Aucun step enregistré\n\n")
	} else {
		for _, step := range r.Steps {
			fmt.Fprintf(&sb, "- `%s`: %s", step.Name, step.Status)
			if step.Message != "" {
				sb.WriteString(" — " + step.Message)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	sb.WriteString("## Tasks\n\n")
	if len(r.Tasks) == 0 {
		sb.WriteString("- Aucune task\n")
	} else {
		for _, task := range r.Tasks {
			fmt.Fprintf(&sb, "- `%s` [%s] %s\n", task.ID, task.Status, extractTaskTitle(task.PayloadJSON))
		}
	}
	if gov := governanceMarkdown(r.Tasks); gov != "" {
		sb.WriteString("\n")
		sb.WriteString(gov)
	}
	if r.Cost != nil {
		sb.WriteString("\n")
		sb.WriteString(CostPerformanceMarkdown(*r.Cost))
	}
	return sb.String()
}

// CostPerformanceMarkdown renders the Cost & Performance table (specv3 §15).
func CostPerformanceMarkdown(c CostPerformance) string {
	var sb strings.Builder
	sb.WriteString("## Cost & Performance\n\n")
	sb.WriteString("| Metric | Estimated | Actual |\n|---|---:|---:|\n")
	fmt.Fprintf(&sb, "| Input tokens | %s | %s |\n", formatInt(c.EstimatedInputTokens), formatInt(c.ActualInputTokens))
	fmt.Fprintf(&sb, "| Output tokens | %s | %s |\n", formatInt(c.EstimatedOutputTokens), formatInt(c.ActualOutputTokens))
	fmt.Fprintf(&sb, "| Cost | %s | %s |\n", c.EstimatedCost, c.ActualCost)
	fmt.Fprintf(&sb, "| Duration | %s | %s |\n", c.EstimatedDuration, c.ActualDuration)
	sb.WriteString("\n## Local Work Saved\n\n")
	fmt.Fprintf(&sb, "- %s files scanned locally\n", formatInt(c.FilesScanned))
	fmt.Fprintf(&sb, "- %s candidate files selected\n", formatInt(c.CandidateFiles))
	if c.LargeFilesSummarized > 0 {
		fmt.Fprintf(&sb, "- %s large files summarized locally\n", formatInt(c.LargeFilesSummarized))
	}
	if c.CloudContextReducedFrom > 0 {
		fmt.Fprintf(&sb, "- estimated cloud context reduced from %s to %s tokens\n",
			formatInt(c.CloudContextReducedFrom), formatInt(c.ActualInputTokens))
	}
	if c.TokenSavingsPercent > 0 {
		fmt.Fprintf(&sb, "- estimated token savings: %.1f%%\n", c.TokenSavingsPercent)
	}
	return sb.String()
}

func formatInt(n int) string {
	return fmt.Sprintf("%d", n)
}

func extractTaskTitle(payloadJSON string) string {
	if payloadJSON == "" {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return ""
	}
	title, _ := payload["title"].(string)
	return title
}

type governanceSnapshot struct {
	TaskID     string  `json:"-"`
	Status     string  `json:"status"`
	Confidence float64 `json:"confidence"`
	Notes      string  `json:"notes"`
	AgentHint  string  `json:"-"`
}

func governanceMarkdown(tasks []sqlite.Task) string {
	var rows []governanceSnapshot
	for _, task := range tasks {
		if snap, ok := lastGovernanceSnapshot(task.PayloadJSON); ok {
			snap.TaskID = task.ID
			rows = append(rows, snap)
		}
	}
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Governance\n\n")
	sb.WriteString("| Task | Status | Confidence | Notes |\n")
	sb.WriteString("|---|---|---:|---|\n")
	for _, row := range rows {
		fmt.Fprintf(&sb, "| `%s` | %s | %.2f | %s |\n", row.TaskID, row.Status, row.Confidence, row.Notes)
	}
	return sb.String()
}

func lastGovernanceSnapshot(payloadJSON string) (governanceSnapshot, bool) {
	if payloadJSON == "" {
		return governanceSnapshot{}, false
	}
	var payload struct {
		Governance *struct {
			History []struct {
				Status     string   `json:"status"`
				Confidence float64  `json:"confidence"`
				Notes      []string `json:"notes"`
				ParseError string   `json:"parse_error"`
			} `json:"history"`
		} `json:"governance"`
	}
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil || payload.Governance == nil || len(payload.Governance.History) == 0 {
		return governanceSnapshot{}, false
	}
	last := payload.Governance.History[len(payload.Governance.History)-1]
	notes := strings.Join(last.Notes, "; ")
	if notes == "" && last.ParseError != "" {
		notes = last.ParseError
	}
	return governanceSnapshot{
		Status:     last.Status,
		Confidence: last.Confidence,
		Notes:      notes,
	}, true
}
