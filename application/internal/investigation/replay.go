package investigation

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// ReplayPack enables reproducible investigation (spec-my-A §25).
type ReplayPack struct {
	InvestigationID string        `json:"investigation_id"`
	CreatedAt       time.Time     `json:"created_at"`
	Request         Request       `json:"request"`
	Scope           ResolvedScope `json:"scope"`
	CommandHint     string        `json:"command_hint"`
}

// WriteReplayPack saves replay-pack.json for re-running investigation.
func WriteReplayPack(dir string, rep Report) (string, error) {
	pack := ReplayPack{
		InvestigationID: rep.ID,
		CreatedAt:       rep.CreatedAt,
		Request:         rep.Request,
		Scope:           rep.Scope,
		CommandHint:     buildReplayCommand(rep),
	}
	path := filepath.Join(dir, "replay-pack.json")
	raw, err := json.MarshalIndent(pack, "", "  ")
	if err != nil {
		return "", err
	}
	return path, os.WriteFile(path, raw, 0o644)
}

func buildReplayCommand(rep Report) string {
	r := rep.Request
	cmd := "asa investigate \"" + r.Symptom + "\""
	if r.Flow != "" {
		cmd += " --flow " + r.Flow
	}
	if r.TaskID != "" {
		cmd += " --task " + r.TaskID
	}
	if r.Depth != "" {
		cmd += " --depth " + string(r.Depth)
	}
	if r.NoCloud {
		cmd += " --no-cloud"
	}
	return cmd
}
