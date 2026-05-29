package product

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const defaultPrototypeURL = "http://localhost:5173"

// PrototypeRunOptions configures local prototype dev-server launch.
type PrototypeRunOptions struct {
	Product string
	DryRun  bool
	Timeout time.Duration // 0 = start in background; >0 = wait then cancel (tests / smoke)
}

// PrototypeRunResult summarizes a prototype run.
type PrototypeRunResult struct {
	Product string
	Dir     string
	URL     string
	Command string
	PID     int
}

type prototypeCommandRunner interface {
	StartDevServer(ctx context.Context, dir string) (*exec.Cmd, error)
}

type npmDevRunner struct{}

func (npmDevRunner) StartDevServer(ctx context.Context, dir string) (*exec.Cmd, error) {
	cmd := exec.CommandContext(ctx, "npm", "run", "dev")
	cmd.Dir = dir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("npm run dev: %w", err)
	}
	return cmd, nil
}

// RunPrototype launches npm run dev in the product prototype directory.
func (s *Service) RunPrototype(opts PrototypeRunOptions) (PrototypeRunResult, error) {
	return s.runPrototypeWithRunner(opts, npmDevRunner{})
}

func (s *Service) runPrototypeWithRunner(opts PrototypeRunOptions, runner prototypeCommandRunner) (PrototypeRunResult, error) {
	product := Slug(opts.Product)
	protoDir := filepath.Join(s.repo.productRoot(product), "prototype")
	pkgPath := filepath.Join(protoDir, "package.json")
	if _, err := os.Stat(pkgPath); err != nil {
		return PrototypeRunResult{}, fmt.Errorf("prototype not found for %q (expected %s): %w", product, pkgPath, err)
	}

	result := PrototypeRunResult{
		Product: product,
		Dir:     protoDir,
		URL:     defaultPrototypeURL,
		Command: "npm run dev",
	}
	if opts.DryRun {
		return result, nil
	}

	ctx := context.Background()
	var cancel context.CancelFunc
	if opts.Timeout > 0 {
		ctx, cancel = context.WithTimeout(ctx, opts.Timeout)
		defer cancel()
	}

	cmd, err := runner.StartDevServer(ctx, protoDir)
	if err != nil {
		return PrototypeRunResult{}, err
	}
	result.PID = cmd.Process.Pid

	if opts.Timeout > 0 {
		waitErr := cmd.Wait()
		if ctx.Err() == context.DeadlineExceeded {
			return result, nil
		}
		if waitErr != nil && !strings.Contains(waitErr.Error(), "signal: killed") {
			return result, fmt.Errorf("prototype dev server exited: %w", waitErr)
		}
		return result, nil
	}

	go func() {
		_ = cmd.Wait()
	}()
	return result, nil
}
