package runtime

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

// StartDaemonDetached starts `asa daemon run` in the background (spec-my-A §24.3).
func StartDaemonDetached(repoRoot string) (int, error) {
	st, err := StartDaemon(repoRoot)
	if err != nil {
		return 0, err
	}
	exe, err := os.Executable()
	if err != nil {
		return st.PID, err
	}
	cmd := exec.Command(exe, "daemon", "run")
	cmd.Dir = repoRoot
	cmd.Stdout = nil
	cmd.Stderr = nil
	if err := cmd.Start(); err != nil {
		return 0, fmt.Errorf("daemon detach: %w", err)
	}
	pid := cmd.Process.Pid
	_ = os.WriteFile(filepath.Join(repoRoot, DefaultRelDir, pidFileName), []byte(strconv.Itoa(pid)), 0o644)
	return pid, nil
}
