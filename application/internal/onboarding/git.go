package onboarding

import (
	"fmt"
	"os/exec"
	"strings"
)

func gitRoot(startDir string) (string, error) {
	cmd := exec.Command("git", "-C", startDir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("dépôt Git requis : exécutez cette commande depuis un clone Git (git init si besoin)")
	}
	return strings.TrimSpace(string(out)), nil
}
