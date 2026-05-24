package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/store/sqlite"
)

type Step struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	Message   string `json:"message,omitempty"`
	StartedAt string `json:"started_at,omitempty"`
	EndedAt   string `json:"ended_at,omitempty"`
}

type RunReport struct {
	RunID      string        `json:"run_id"`
	Feature    string        `json:"feature"`
	Status     string        `json:"status"`
	Generated  string        `json:"generated_at"`
	Steps      []Step        `json:"steps"`
	Tasks      []sqlite.Task `json:"tasks"`
	Repository string        `json:"repository"`
	Cost       *CostPerformance `json:"cost_performance,omitempty"`
}

// CostPerformance is an optional V3 section (specv3 §15).
type CostPerformance struct {
	EstimatedInputTokens  int    `json:"estimated_input_tokens"`
	EstimatedOutputTokens int    `json:"estimated_output_tokens"`
	ActualInputTokens     int    `json:"actual_input_tokens"`
	ActualOutputTokens    int    `json:"actual_output_tokens"`
	EstimatedCost         string `json:"estimated_cost"`
	ActualCost            string `json:"actual_cost"`
	EstimatedDuration     string `json:"estimated_duration"`
	ActualDuration        string `json:"actual_duration"`
	FilesScanned          int    `json:"files_scanned_local"`
	CandidateFiles        int    `json:"candidate_files"`
}

type Writer struct {
	RepoRoot string
}

func NewWriter(repoRoot string) *Writer {
	return &Writer{RepoRoot: repoRoot}
}

func (w *Writer) Write(run sqlite.Run, tasks []sqlite.Task, steps []Step) (string, string, error) {
	runDir := filepath.Join(w.RepoRoot, ".agentflow", "runs", run.ID)
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
	sb.WriteString("# AgentFlow Report\n\n")
	sb.WriteString(fmt.Sprintf("- Run: `%s`\n", r.RunID))
	sb.WriteString(fmt.Sprintf("- Feature: `%s`\n", r.Feature))
	sb.WriteString(fmt.Sprintf("- Status: `%s`\n", r.Status))
	sb.WriteString(fmt.Sprintf("- Generated: `%s`\n\n", r.Generated))

	sb.WriteString("## Steps\n\n")
	if len(r.Steps) == 0 {
		sb.WriteString("- Aucun step enregistré\n\n")
	} else {
		for _, step := range r.Steps {
			sb.WriteString(fmt.Sprintf("- `%s`: %s", step.Name, step.Status))
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
			sb.WriteString(fmt.Sprintf("- `%s` [%s] %s\n", task.ID, task.Status, extractTaskTitle(task.PayloadJSON)))
		}
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
	sb.WriteString(fmt.Sprintf("| Input tokens | %s | %s |\n", formatInt(c.EstimatedInputTokens), formatInt(c.ActualInputTokens)))
	sb.WriteString(fmt.Sprintf("| Output tokens | %s | %s |\n", formatInt(c.EstimatedOutputTokens), formatInt(c.ActualOutputTokens)))
	sb.WriteString(fmt.Sprintf("| Cost | %s | %s |\n", c.EstimatedCost, c.ActualCost))
	sb.WriteString(fmt.Sprintf("| Duration | %s | %s |\n", c.EstimatedDuration, c.ActualDuration))
	sb.WriteString("\n## Local Work Saved\n\n")
	sb.WriteString(fmt.Sprintf("- %s files scanned locally\n", formatInt(c.FilesScanned)))
	sb.WriteString(fmt.Sprintf("- %s candidate files selected\n", formatInt(c.CandidateFiles)))
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
