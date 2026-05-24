package investigation

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"
)

// RunCommand runs argv with timeout; never uses shell.
func RunCommand(ctx context.Context, timeout time.Duration, name string, args ...string) ([]byte, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("investigation: timeout requis")
	}
	cctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(cctx, name, args...)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()
	if err != nil {
		if errors.Is(cctx.Err(), context.DeadlineExceeded) {
			return buf.Bytes(), fmt.Errorf("investigation: timeout %s: %w", name, cctx.Err())
		}
		return buf.Bytes(), fmt.Errorf("investigation: %s: %w (%s)", name, err, buf.String())
	}
	return buf.Bytes(), nil
}
