package bootstrap

import (
	"github.com/LaProgrammerie/asagiri/application/internal/config"
	"github.com/LaProgrammerie/asagiri/application/internal/onboarding"
)

// DoctorOptions configures doctor behaviour.
type DoctorOptions struct {
	Full bool
}

// DoctorWithOptions runs environment checks with optional onboarding extensions.
func DoctorWithOptions(startDir string, opts DoctorOptions) ([]DoctorCheck, error) {
	checks, err := Doctor(startDir)
	if err != nil {
		return checks, err
	}
	if !opts.Full {
		return checks, nil
	}

	repoRoot, gitErr := GitRoot(startDir)
	if gitErr != nil {
		return checks, nil
	}
	cfgPath := config.ConfigPath(repoRoot)
	cfg, cfgErr := config.Load(cfgPath, repoRoot)
	if cfgErr != nil {
		cfg = nil
	}
	for _, c := range onboarding.RunDoctorChecks(repoRoot, cfg, onboarding.DoctorOpts{Full: true, SkipExec: true}) {
		checks = append(checks, onboardingCheckToBootstrap(c))
	}
	return checks, nil
}

func onboardingCheckToBootstrap(c onboarding.Check) DoctorCheck {
	var err error
	if c.Status == onboarding.StatusFail {
		if c.Message != "" {
			err = &doctorMessageError{msg: c.Message}
		} else {
			err = &doctorMessageError{msg: "check failed"}
		}
	} else if c.Status == onboarding.StatusWarn {
		if c.Message != "" {
			err = &doctorWarnError{msg: c.Message}
		}
	}
	return DoctorCheck{Name: c.ID, Err: err}
}

type doctorMessageError struct{ msg string }

func (e *doctorMessageError) Error() string { return e.msg }

type doctorWarnError struct{ msg string }

func (e *doctorWarnError) Error() string { return "warn: " + e.msg }
