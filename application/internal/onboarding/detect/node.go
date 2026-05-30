package detect

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// NodeDetector matches package.json at repo root.
type NodeDetector struct{}

func (NodeDetector) ID() string { return "node" }

func (NodeDetector) Detect(repoRoot string) (StackMatch, error) {
	rel := "package.json"
	if _, err := os.Stat(filepath.Join(repoRoot, rel)); err != nil {
		return StackMatch{}, nil
	}
	return StackMatch{
		ID:         "node",
		Confidence: 0.85,
		Signals:    []string{rel},
	}, nil
}

func (NodeDetector) ValidationCommands(repoRoot string) []config.ValidationCommand {
	path := filepath.Join(repoRoot, "package.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return []config.ValidationCommand{
			{Name: "test", Command: "npm test", Required: true},
		}
	}
	var pkg struct {
		Scripts map[string]string `json:"scripts"`
	}
	if err := json.Unmarshal(data, &pkg); err != nil {
		return []config.ValidationCommand{
			{Name: "test", Command: "npm test", Required: true},
		}
	}
	if pkg.Scripts != nil {
		if _, ok := pkg.Scripts["qa:js"]; ok {
			return []config.ValidationCommand{
				{Name: "qa-js", Command: "npm run qa:js", Required: true},
			}
		}
	}
	return []config.ValidationCommand{
		{Name: "test", Command: "npm test", Required: true},
	}
}
