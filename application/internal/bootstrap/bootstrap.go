package bootstrap

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/store/sqlite"
)

const runtimeDir = ".asagiri"

var runtimeDirs = []string{"runs", "tasks", "logs", "worktrees"}

// GitRoot returns the repository root for startDir (typically cwd).
func GitRoot(startDir string) (string, error) {
	cmd := exec.Command("git", "-C", startDir, "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("dépôt Git requis : exécutez cette commande depuis un clone Git (git init si besoin)")
	}
	return strings.TrimSpace(string(out)), nil
}

// Init bootstraps .asagiri/ in the Git repository containing startDir.
func Init(startDir string) error {
	repoRoot, err := GitRoot(startDir)
	if err != nil {
		return err
	}

	base := filepath.Join(repoRoot, runtimeDir)
	for _, sub := range runtimeDirs {
		if err := os.MkdirAll(filepath.Join(base, sub), 0o755); err != nil {
			return fmt.Errorf("créer %s/%s: %w", runtimeDir, sub, err)
		}
	}

	cfgPath := config.ConfigPath(repoRoot)
	if err := ensureConfig(repoRoot, cfgPath); err != nil {
		return err
	}

	cfg, err := config.Load(cfgPath, repoRoot)
	if err != nil {
		return err
	}

	store, err := sqlite.Open(cfg.StateDBPath(repoRoot))
	if err != nil {
		return err
	}
	defer store.Close()

	if err := store.Migrate(); err != nil {
		return fmt.Errorf("migrations SQLite: %w", err)
	}

	return nil
}

func ensureConfig(repoRoot, cfgPath string) error {
	if _, statErr := os.Stat(cfgPath); statErr == nil {
		return nil
	} else if !os.IsNotExist(statErr) {
		return fmt.Errorf("stat config: %w", statErr)
	}

	examplePath := config.ExamplePath(repoRoot)
	if _, err := os.Stat(examplePath); err == nil {
		return copyFile(examplePath, cfgPath)
	}

	return fmt.Errorf("config manquant : copiez %s vers %s", config.DefaultExampleRel, config.DefaultConfigRel)
}

func copyFile(src, dst string) error {
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("ouvrir example config: %w", err)
	}
	defer in.Close()

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		return fmt.Errorf("créer config: %w", err)
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return fmt.Errorf("copier config: %w", err)
	}
	return nil
}

// DoctorCheck describes one diagnostic step.
type DoctorCheck struct {
	Name string
	Err  error
}

// Doctor runs environment checks from startDir.
func Doctor(startDir string) ([]DoctorCheck, error) {
	var checks []DoctorCheck

	repoRoot, gitErr := GitRoot(startDir)
	checks = append(checks, DoctorCheck{Name: "git", Err: gitErr})
	if gitErr != nil {
		return checks, nil
	}

	cfgPath := config.ConfigPath(repoRoot)
	cfg, cfgErr := config.Load(cfgPath, repoRoot)
	checks = append(checks, DoctorCheck{Name: "config", Err: cfgErr})
	if cfgErr != nil {
		return checks, nil
	}

	for _, sub := range runtimeDirs {
		p := filepath.Join(repoRoot, runtimeDir, sub)
		if _, err := os.Stat(p); err != nil {
			checks = append(checks, DoctorCheck{
				Name: "dir:" + sub,
				Err:  fmt.Errorf("%s manquant — lancez asa init", filepath.Join(runtimeDir, sub)),
			})
		}
	}

	store, dbErr := sqlite.Open(cfg.StateDBPath(repoRoot))
	if dbErr != nil {
		checks = append(checks, DoctorCheck{Name: "sqlite", Err: dbErr})
		return checks, nil
	}
	defer store.Close()

	if err := store.Ping(); err != nil {
		checks = append(checks, DoctorCheck{Name: "sqlite", Err: err})
		return checks, nil
	}

	v, err := store.SchemaVersion()
	if err != nil {
		checks = append(checks, DoctorCheck{Name: "schema", Err: err})
		return checks, nil
	}
	if v < 1 {
		checks = append(checks, DoctorCheck{
			Name: "schema",
			Err:  fmt.Errorf("schéma non migré (version %d) — lancez asa init", v),
		})
		return checks, nil
	}

	checks = append(checks, DoctorCheck{Name: "schema", Err: nil})
	return checks, nil
}

// FormatDoctor writes human-readable doctor output.
func FormatDoctor(w io.Writer, checks []DoctorCheck) error {
	ok := true
	for _, c := range checks {
		if c.Err != nil {
			ok = false
			if _, err := fmt.Fprintf(w, "✗ %s: %v\n", c.Name, c.Err); err != nil {
				return err
			}
		} else {
			if _, err := fmt.Fprintf(w, "✓ %s\n", c.Name); err != nil {
				return err
			}
		}
	}
	if ok {
		_, err := fmt.Fprintln(w, "Asagiri est prêt.")
		return err
	}
	return fmt.Errorf("doctor: au moins un contrôle a échoué")
}
