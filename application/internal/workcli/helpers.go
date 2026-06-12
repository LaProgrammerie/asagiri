package workcli

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
)

func isInteractive() bool {
	if env.YesEnabled() {
		return false
	}
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func confirmPrompt(w io.Writer, r io.Reader, msg string) (bool, error) {
	_, _ = fmt.Fprintf(w, "%s [y/N]: ", msg)
	sc := bufio.NewScanner(r)
	if !sc.Scan() {
		return false, sc.Err()
	}
	ans := strings.ToLower(strings.TrimSpace(sc.Text()))
	return ans == "y" || ans == "yes", nil
}

func requireConfirm(opts intent.WorkOptions, msg string) error {
	if opts.Yes || !opts.Interactive {
		if !opts.Yes && !opts.Interactive {
			return &intent.ConfirmationRequiredError{Message: msg + " (mode non interactif: utilisez --yes)"}
		}
		return nil
	}
	ok, err := confirmPrompt(os.Stderr, os.Stdin, msg)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("annulé par l'utilisateur")
	}
	return nil
}
