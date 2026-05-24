package plan

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/spec"
	"github.com/LaProgrammerie/hyper-fast-builder/application/internal/store/sqlite"
)

var (
	markdownTaskPattern = regexp.MustCompile(`^[-*]\s+\[( |x|X)\]\s+(.+)$`)
	numberedTaskPattern = regexp.MustCompile(`^\d+[.)]\s+(.+)$`)
)

// Task is one normalized plan item.
type Task struct {
	ID      string   `json:"id"`
	Title   string   `json:"title"`
	Status  string   `json:"status"`
	Scope   []string `json:"scope,omitempty"`
	Checks  []string `json:"checks,omitempty"`
	Details string   `json:"details,omitempty"`
}

// Normalize extracts task lines from spec markdown.
func Normalize(feature string, doc *spec.Document) ([]Task, error) {
	if doc == nil {
		return nil, fmt.Errorf("document spec nil")
	}
	source := doc.Tasks
	if strings.TrimSpace(source) == "" {
		source = doc.CombinedText()
	}
	lines := strings.Split(source, "\n")
	tasks := make([]Task, 0)
	for _, raw := range lines {
		line := strings.TrimSpace(raw)
		if line == "" {
			continue
		}
		title := ""
		if match := markdownTaskPattern.FindStringSubmatch(line); len(match) == 3 {
			title = strings.TrimSpace(match[2])
		} else if match := numberedTaskPattern.FindStringSubmatch(line); len(match) == 2 {
			title = strings.TrimSpace(match[1])
		}
		if title == "" {
			continue
		}
		id := fmt.Sprintf("%s-%03d", sanitizeID(feature), len(tasks)+1)
		tasks = append(tasks, Task{
			ID:     id,
			Title:  title,
			Status: sqlite.StatusPending,
		})
	}
	if len(tasks) == 0 {
		tasks = append(tasks, Task{
			ID:      fmt.Sprintf("%s-001", sanitizeID(feature)),
			Title:   "Implémenter la feature selon la spec",
			Status:  sqlite.StatusPending,
			Details: "Fallback task generated from active spec",
		})
	}
	return tasks, nil
}

func sanitizeID(feature string) string {
	s := strings.ToLower(strings.TrimSpace(feature))
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "/", "-")
	s = regexp.MustCompile(`[^a-z0-9-]+`).ReplaceAllString(s, "")
	if s == "" {
		return "task"
	}
	return s
}

func ToPayloadJSON(task Task) (string, error) {
	body, err := json.Marshal(task)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
