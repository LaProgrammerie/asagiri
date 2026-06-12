package detect

import (
	"sort"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// StackMatch is one detector hit.
type StackMatch struct {
	ID         string   `json:"id"`
	Confidence float64  `json:"confidence"`
	Signals    []string `json:"signals"`
}

// Detector finds stack signals and proposes validation commands.
type Detector interface {
	ID() string
	Detect(repoRoot string) (StackMatch, error)
	ValidationCommands(repoRoot string) []config.ValidationCommand
}

// DefaultDetectors is the V1 registry (Go, Castor, Node).
var DefaultDetectors = []Detector{
	&GoDetector{},
	&CastorDetector{},
	&NodeDetector{},
}

// DetectAll runs detectors and merges validation commands deduplicated by command string.
func DetectAll(repoRoot string, override string) ([]StackMatch, []config.ValidationCommand) {
	override = strings.ToLower(strings.TrimSpace(override))
	if override != "" && override != "auto" {
		for _, d := range DefaultDetectors {
			if d.ID() == override || (override == "php" && d.ID() == "castor") {
				match, _ := d.Detect(repoRoot)
				if match.ID == "" {
					match = StackMatch{ID: d.ID(), Confidence: 1, Signals: []string{"override:" + override}}
				}
				return []StackMatch{match}, d.ValidationCommands(repoRoot)
			}
		}
		return nil, nil
	}

	var matches []StackMatch
	cmdIndex := map[string]config.ValidationCommand{}
	for _, d := range DefaultDetectors {
		match, err := d.Detect(repoRoot)
		if err != nil || match.ID == "" {
			continue
		}
		matches = append(matches, match)
		for _, cmd := range d.ValidationCommands(repoRoot) {
			key := strings.TrimSpace(cmd.Command)
			if key == "" {
				continue
			}
			if _, ok := cmdIndex[key]; !ok {
				cmdIndex[key] = cmd
			}
		}
	}
	sort.Slice(matches, func(i, j int) bool {
		return matches[i].Confidence > matches[j].Confidence
	})
	cmds := make([]config.ValidationCommand, 0, len(cmdIndex))
	for _, cmd := range cmdIndex {
		cmds = append(cmds, cmd)
	}
	sort.Slice(cmds, func(i, j int) bool {
		return cmds[i].Name < cmds[j].Name
	})
	return matches, cmds
}

// PrimaryStack returns the highest-confidence stack id or empty.
func PrimaryStack(matches []StackMatch) string {
	if len(matches) == 0 {
		return ""
	}
	return matches[0].ID
}
