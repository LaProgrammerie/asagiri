package checks

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const typeTests = "tests"

// TestsRunner links investigation candidates to related tests and go test results.
type TestsRunner struct{}

func (TestsRunner) Type() string { return typeTests }

func (TestsRunner) Run(ctx context.Context, scope Scope, deps Dependencies) (CheckResult, error) {
	start := time.Now()
	findings := make([]Finding, 0)
	evidence := make([]Evidence, 0)

	modDir := filepath.Join(scope.RepoRoot, "application")
	if st, err := os.Stat(filepath.Join(modDir, "go.mod")); err != nil || st.IsDir() {
		if st, err := os.Stat(filepath.Join(scope.RepoRoot, "go.mod")); err != nil || st.IsDir() {
			return CheckResult{
				ID:         checkID(typeTests, scope.TrustID),
				Name:       "Tests",
				Type:       typeTests,
				Status:     statusSkipped,
				Confidence: 0,
				Findings: []Finding{{
					Severity: "info",
					Category: "implementation.test",
					Message:  "no Go module in repository",
				}},
				Duration: time.Since(start),
			}, nil
		}
		modDir = scope.RepoRoot
	}

	inv, err := deps.Investigate(ctx, scope.RepoRoot, scope.Flow, scope.Task, deps.Config)
	candidates := inv.CandidateFiles
	if err != nil {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "implementation.test",
			Message:  fmt.Sprintf("investigation for test scope: %v", err),
		})
	} else if len(candidates) == 0 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "implementation.test",
			Message:  "no investigation candidates for test scope",
		})
	}

	related := deps.RelatedTests(candidates)
	if len(related) == 0 && len(candidates) > 0 {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "regression.test",
			Message:  "no related *_test.go files for investigation candidates",
		})
	}
	evidence = append(evidence, Evidence{
		Kind:    "tests",
		Source:  modDir,
		Summary: fmt.Sprintf("%d related test files", len(related)),
	})

	failed, testErr := deps.ParseFailedTests(ctx, scope.RepoRoot)
	if testErr != nil {
		findings = append(findings, Finding{
			Severity: "warning",
			Category: "regression.test",
			Message:  fmt.Sprintf("go test: %v", testErr),
		})
	}
	for _, f := range failed {
		findings = append(findings, Finding{
			Severity: "error",
			Category: "regression.test",
			Message:  "failing test: " + f,
		})
	}

	status := statusFromFindings(findings)
	if len(failed) > 0 {
		status = statusFailed
	}
	conf := confidenceFromStatus(status)
	if len(candidates) == 0 && conf > testsNoCandidatesCap {
		conf = testsNoCandidatesCap
		if status == statusPassed {
			status = statusWarn
		}
	}

	return CheckResult{
		ID:         checkID(typeTests, scope.TrustID),
		Name:       "Tests",
		Type:       typeTests,
		Status:     status,
		Confidence: roundConfidence(conf),
		Findings:   findings,
		Evidence:   evidence,
		Duration:   time.Since(start),
	}, nil
}
