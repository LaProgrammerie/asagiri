package reportsink

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/trust/safeid"
)

const runtimeDir = ".asagiri"
const reportsRoot = "reports"

// ErrRuntimeAbsent is returned when .asagiri is missing.
var ErrRuntimeAbsent = fmt.Errorf("%s absent — lancez asa init", runtimeDir)

// RequireInitialized ensures .asagiri exists before writing report artefacts.
func RequireInitialized(repoRoot string) error {
	repoRoot = strings.TrimSpace(repoRoot)
	if repoRoot == "" {
		return fmt.Errorf("reports: repo root required")
	}
	p := filepath.Join(repoRoot, runtimeDir)
	st, err := os.Stat(p)
	if err != nil {
		if os.IsNotExist(err) {
			return ErrRuntimeAbsent
		}
		return fmt.Errorf("reports: stat %s: %w", runtimeDir, err)
	}
	if !st.IsDir() {
		return fmt.Errorf("reports: %s n'est pas un répertoire", runtimeDir)
	}
	return nil
}

// SaveTrustTask writes a work trust task report JSON atomically.
func SaveTrustTask(repoRoot, taskID string, report any) (string, error) {
	return SaveTrustTaskWithOptions(repoRoot, taskID, report, DefaultSaveOptions())
}

// SaveTrustTaskWithOptions writes a task report with optional history archival.
func SaveTrustTaskWithOptions(repoRoot, taskID string, report any, opts SaveOptions) (string, error) {
	rel, err := trustTaskRel(taskID)
	if err != nil {
		return "", err
	}
	return save(repoRoot, rel, report, "trust task", opts)
}

// SaveTrustFeature writes a work trust feature report JSON atomically.
func SaveTrustFeature(repoRoot, feature string, report any) (string, error) {
	return SaveTrustFeatureWithOptions(repoRoot, feature, report, DefaultSaveOptions())
}

// SaveTrustFeatureWithOptions writes a feature report with optional history archival.
func SaveTrustFeatureWithOptions(repoRoot, feature string, report any, opts SaveOptions) (string, error) {
	rel, err := trustFeatureRel(feature)
	if err != nil {
		return "", err
	}
	return save(repoRoot, rel, report, "trust feature", opts)
}

// SaveTrustRun writes a work trust run report JSON atomically.
func SaveTrustRun(repoRoot, runID string, report any) (string, error) {
	return SaveTrustRunWithOptions(repoRoot, runID, report, DefaultSaveOptions())
}

// SaveTrustRunWithOptions writes a run report with optional history archival.
func SaveTrustRunWithOptions(repoRoot, runID string, report any, opts SaveOptions) (string, error) {
	rel, err := trustRunRel(runID)
	if err != nil {
		return "", err
	}
	return save(repoRoot, rel, report, "trust run", opts)
}

// SaveDoctor writes the latest doctor report JSON atomically.
func SaveDoctor(repoRoot string, report any) (string, error) {
	return SaveDoctorWithOptions(repoRoot, report, DefaultSaveOptions())
}

// SaveDoctorWithOptions writes doctor latest with optional history archival.
func SaveDoctorWithOptions(repoRoot string, report any, opts SaveOptions) (string, error) {
	return save(repoRoot, doctorLatestRel(), report, "doctor", opts)
}

func save(repoRoot, relUnderRuntime string, report any, label string, opts SaveOptions) (string, error) {
	if err := RequireInitialized(repoRoot); err != nil {
		return "", err
	}
	abs := filepath.Join(repoRoot, runtimeDir, relUnderRuntime)
	if err := archiveBeforeOverwrite(repoRoot, relUnderRuntime, abs, opts); err != nil {
		return "", err
	}
	if err := writeJSONAtomic(abs, report); err != nil {
		return "", fmt.Errorf("save %s report: %w", label, err)
	}
	return relRepo(repoRoot, abs), nil
}

func trustTaskRel(taskID string) (string, error) {
	if err := validateSegment(taskID); err != nil {
		return "", err
	}
	return filepath.Join(reportsRoot, "trust", "tasks", taskID+".json"), nil
}

func trustFeatureRel(feature string) (string, error) {
	if err := validateSegment(feature); err != nil {
		return "", err
	}
	return filepath.Join(reportsRoot, "trust", "features", feature+".json"), nil
}

func trustRunRel(runID string) (string, error) {
	if err := validateSegment(runID); err != nil {
		return "", err
	}
	return filepath.Join(reportsRoot, "trust", "runs", runID+".json"), nil
}

func doctorLatestRel() string {
	return filepath.Join(reportsRoot, "doctor", "latest.json")
}

func validateSegment(seg string) error {
	if err := safeid.Validate(seg); err != nil {
		return fmt.Errorf("invalid report id: %w", err)
	}
	return nil
}

func writeJSONAtomic(absPath string, v any) error {
	body, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal json: %w", err)
	}
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("mkdir: %w", err)
	}
	f, err := os.CreateTemp(dir, ".report-*.tmp")
	if err != nil {
		return fmt.Errorf("temp file: %w", err)
	}
	tmpPath := f.Name()
	cleanup := true
	defer func() {
		if cleanup {
			_ = os.Remove(tmpPath)
		}
	}()
	if _, err := f.Write(body); err != nil {
		_ = f.Close()
		return fmt.Errorf("write temp: %w", err)
	}
	if err := f.Chmod(0o644); err != nil {
		_ = f.Close()
		return fmt.Errorf("chmod temp: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close temp: %w", err)
	}
	if err := os.Rename(tmpPath, absPath); err != nil {
		return fmt.Errorf("rename report: %w", err)
	}
	cleanup = false
	return nil
}

func relRepo(repoRoot, absPath string) string {
	if rel, err := filepath.Rel(repoRoot, absPath); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(absPath)
}

// TrustTaskAbs returns the absolute path for a saved task report (tests).
func TrustTaskAbs(repoRoot, taskID string) (string, error) {
	rel, err := trustTaskRel(taskID)
	if err != nil {
		return "", err
	}
	return filepath.Join(repoRoot, runtimeDir, rel), nil
}

// DoctorLatestAbs returns the absolute path for the latest doctor report (tests).
func DoctorLatestAbs(repoRoot string) string {
	return filepath.Join(repoRoot, runtimeDir, doctorLatestRel())
}
