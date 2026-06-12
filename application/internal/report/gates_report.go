package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/gates"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

type gateReportRow struct {
	Gate       string
	Scope      string
	Status     string
	Confidence float64
	Notes      string
}

func gatesMarkdown(repoRoot, runID string, tasks []sqlite.Task) string {
	rows := collectGateReportRows(repoRoot, runID, tasks)
	if len(rows) == 0 {
		return ""
	}
	var sb strings.Builder
	sb.WriteString("## Gates\n\n")
	sb.WriteString("| Gate | Scope | Status | Confidence | Notes |\n")
	sb.WriteString("|---|---|---|---:|---|\n")
	for _, row := range rows {
		fmt.Fprintf(&sb, "| `%s` | `%s` | %s | %.2f | %s |\n",
			row.Gate, row.Scope, row.Status, row.Confidence, row.Notes)
	}
	return sb.String()
}

func collectGateReportRows(repoRoot, runID string, tasks []sqlite.Task) []gateReportRow {
	var rows []gateReportRow
	seen := make(map[string]bool)

	for _, task := range tasks {
		for _, row := range gateRowsFromTaskPayload(task.ID, task.PayloadJSON) {
			key := row.Gate + "\x00" + row.Scope
			if seen[key] {
				continue
			}
			seen[key] = true
			rows = append(rows, row)
		}
	}

	if runID != "" {
		if row, ok := gateRowFromLogFile(repoRoot, runID, "plan"); ok {
			key := row.Gate + "\x00" + row.Scope
			if !seen[key] {
				rows = append(rows, row)
			}
		}
	}

	for _, task := range tasks {
		for _, gateName := range []string{"enrich", "governance", "human_review", "verify_evidence", "trust"} {
			key := gateName + "\x00" + task.ID
			if seen[key] {
				continue
			}
			if row, ok := gateRowFromLogFile(repoRoot, task.ID, gateName); ok {
				rows = append(rows, row)
			}
		}
	}

	return rows
}

func gateRowsFromTaskPayload(taskID, payloadJSON string) []gateReportRow {
	if payloadJSON == "" {
		return nil
	}
	var payload struct {
		Gates *struct {
			History []gateHistoryPayload `json:"history"`
		} `json:"gates"`
		Governance *struct {
			History []gateHistoryPayload `json:"history"`
		} `json:"governance"`
	}
	if err := json.Unmarshal([]byte(payloadJSON), &payload); err != nil {
		return nil
	}

	var rows []gateReportRow
	if payload.Gates != nil {
		byGate := lastHistoryByGate(payload.Gates.History)
		for gateName, entry := range byGate {
			rows = append(rows, snapshotToRow(gateName, taskID, entry))
		}
	}

	if payload.Governance != nil && len(payload.Governance.History) > 0 {
		hasGovernanceInGates := false
		if payload.Gates != nil {
			for _, h := range payload.Gates.History {
				if strings.EqualFold(strings.TrimSpace(h.Gate), "governance") || h.Gate == "" {
					hasGovernanceInGates = true
					break
				}
			}
		}
		if !hasGovernanceInGates {
			last := payload.Governance.History[len(payload.Governance.History)-1]
			last.Gate = "governance"
			rows = append(rows, snapshotToRow("governance", taskID, last))
		}
	}
	return rows
}

type gateHistoryPayload struct {
	Gate       string   `json:"gate"`
	Status     string   `json:"status"`
	Confidence float64  `json:"confidence"`
	Notes      []string `json:"notes"`
	ParseError string   `json:"parse_error"`
}

func lastHistoryByGate(history []gateHistoryPayload) map[string]gateHistoryPayload {
	out := make(map[string]gateHistoryPayload)
	for _, h := range history {
		name := strings.TrimSpace(h.Gate)
		if name == "" {
			name = "governance"
		}
		out[name] = h
	}
	return out
}

func snapshotToRow(gateName, scope string, h gateHistoryPayload) gateReportRow {
	notes := strings.Join(h.Notes, "; ")
	if notes == "" && h.ParseError != "" {
		notes = h.ParseError
	}
	return gateReportRow{
		Gate:       gateName,
		Scope:      scope,
		Status:     h.Status,
		Confidence: h.Confidence,
		Notes:      notes,
	}
}

func gateRowFromLogFile(repoRoot, scopeID, gateName string) (gateReportRow, bool) {
	path := filepath.Join(repoRoot, ".asagiri", "logs", scopeID, "gates", gateName+".json")
	body, err := os.ReadFile(path)
	if err != nil {
		return gateReportRow{}, false
	}
	var doc gates.LogDocument
	if err := json.Unmarshal(body, &doc); err != nil {
		return gateReportRow{}, false
	}
	notes := strings.Join(doc.Notes, "; ")
	if notes == "" && doc.ParseError != "" {
		notes = doc.ParseError
	}
	scope := scopeID
	if doc.TaskID != "" {
		scope = doc.TaskID
	} else if doc.RunID != "" {
		scope = doc.RunID
	}
	name := strings.TrimSpace(doc.GateName)
	if name == "" {
		name = gateName
	}
	return gateReportRow{
		Gate:       name,
		Scope:      scope,
		Status:     doc.Status,
		Confidence: doc.Confidence,
		Notes:      notes,
	}, true
}
