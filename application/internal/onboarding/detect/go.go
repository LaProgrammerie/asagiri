package detect

import (
	"os"
	"path/filepath"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// GoDetector matches go.mod at repo root or under application/.
type GoDetector struct{}

func (GoDetector) ID() string { return "go" }

func (GoDetector) Detect(repoRoot string) (StackMatch, error) {
	for _, rel := range []string{"go.mod", filepath.Join("application", "go.mod")} {
		if _, err := os.Stat(filepath.Join(repoRoot, rel)); err == nil {
			return StackMatch{
				ID:         "go",
				Confidence: 0.95,
				Signals:    []string{rel},
			}, nil
		}
	}
	return StackMatch{}, nil
}

func (GoDetector) ValidationCommands(repoRoot string) []config.ValidationCommand {
	if _, err := os.Stat(filepath.Join(repoRoot, "go.mod")); err != nil {
		if _, err2 := os.Stat(filepath.Join(repoRoot, "application", "go.mod")); err2 != nil {
			return nil
		}
	}
	return config.DefaultGoValidationCommands(filepath.Base(repoRoot))
}
