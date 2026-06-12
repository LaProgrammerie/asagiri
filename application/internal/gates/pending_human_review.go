package gates

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
	"github.com/LaProgrammerie/asagiri/application/pkg/asagiri"
)

const HumanReviewGateName = "human_review"
const DefaultHumanReviewVerdictFile = "human_review.verdict.yaml"

// HumanReviewVerdictPath returns the on-disk path for a human review verdict file.
func HumanReviewVerdictPath(repoRoot, taskID, customVerdictFile string) string {
	name := DefaultHumanReviewVerdictFile
	if strings.TrimSpace(customVerdictFile) != "" {
		name = customVerdictFile
	}
	return filepath.Join(repoRoot, ".asagiri", "logs", taskID, "gates", name)
}

func humanReviewPending(repoRoot string, hr config.WorkHumanReviewGateConfig, task sqlite.Task) (PendingGate, bool) {
	if !hr.IsActive() {
		return PendingGate{}, false
	}
	if !taskAwaitingPostDevGate(task.Status) {
		return PendingGate{}, false
	}
	if entry, ok := LastGateEntry(task.PayloadJSON, HumanReviewGateName); ok && GateEntrySatisfied(hr.WarnAdvisory(), entry) {
		return PendingGate{}, false
	}
	path := HumanReviewVerdictPath(repoRoot, task.ID, hr.VerdictFile)
	phase := PendingPhaseResume
	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			phase = PendingPhaseSubmit
		} else {
			return PendingGate{}, false
		}
	}
	return PendingGate{
		Gate:     HumanReviewGateName,
		Scope:    task.ID,
		Blocking: true,
		Phase:    phase,
	}, true
}

func taskAwaitingPostDevGate(status string) bool {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case asagiri.StatusImplemented, sqlite.StatusDone,
		asagiri.StatusVerified, asagiri.StatusReviewFailed:
		return true
	default:
		return false
	}
}

func lastGateEntryFromTask(task asagiri.Task, gateName string) (asagiri.GateHistoryEntry, bool) {
	if task.Gates == nil {
		return asagiri.GateHistoryEntry{}, false
	}
	var last asagiri.GateHistoryEntry
	found := false
	for _, e := range task.Gates.History {
		if strings.EqualFold(strings.TrimSpace(e.Gate), gateName) {
			last = e
			found = true
		}
	}
	return last, found
}

// payload helpers used by LastGateEntry in pending.go
func unmarshalTaskPayload(payloadJSON string) (asagiri.Task, error) {
	var task asagiri.Task
	if payloadJSON == "" {
		return task, nil
	}
	err := json.Unmarshal([]byte(payloadJSON), &task)
	return task, err
}
