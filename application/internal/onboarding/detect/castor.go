package detect

import (
	"os"
	"path/filepath"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
)

// CastorDetector matches castor.php at repo root.
type CastorDetector struct{}

func (CastorDetector) ID() string { return "castor" }

func (CastorDetector) Detect(repoRoot string) (StackMatch, error) {
	rel := "castor.php"
	if _, err := os.Stat(filepath.Join(repoRoot, rel)); err != nil {
		return StackMatch{}, nil
	}
	return StackMatch{
		ID:         "castor",
		Confidence: 0.9,
		Signals:    []string{rel},
	}, nil
}

func (CastorDetector) ValidationCommands(_ string) []config.ValidationCommand {
	return []config.ValidationCommand{
		{Name: "static-checks", Command: "castor qa:static-checks", Required: true},
		{Name: "phpunit", Command: "castor qa:phpunit", Required: true},
	}
}
