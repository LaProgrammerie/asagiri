package bootstrap

import (
	"fmt"
	"os/exec"
	"strings"
)

// GitHead returns the short commit hash at HEAD for repoRoot.
func GitHead(repoRoot string) (string, error) {
	cmd := exec.Command("git", "-C", repoRoot, "rev-parse", "--short", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("git rev-parse HEAD: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}
