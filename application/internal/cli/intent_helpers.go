package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/LaProgrammerie/asagiri/application/internal/env"
	"github.com/LaProgrammerie/asagiri/application/internal/intent"
	"github.com/LaProgrammerie/asagiri/application/internal/source"
	srcnotion "github.com/LaProgrammerie/asagiri/application/internal/source/notion"
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
	fmt.Fprintf(w, "%s [y/N]: ", msg)
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

func (c *appContext) snapshot() (intent.StateSnapshot, error) {
	return intent.BuildSnapshot(c.RepoRoot, c.Config, c.Store)
}

func (c *appContext) syncPrimitive(ctx context.Context, args []string, force bool) error {
	reg := newSourceRegistry(c, notionClientFromConfig(c))
	if len(args) == 0 {
		return fmt.Errorf("sync: source requise")
	}
	srcName := args[0]
	src, err := reg.byName(srcName)
	if err != nil {
		return err
	}
	ref := source.SourceRef{}
	dest := source.LocalSpecPath{Root: c.Config.Sources.Notion.ImportPath}
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--page":
			if i+1 < len(args) {
				ref.URL = args[i+1]
				i++
			}
		case "--feature":
			if i+1 < len(args) {
				dest.Feature = args[i+1]
				ref.Name = dest.Feature
				i++
			}
		}
	}
	opts := source.SyncOptions{Force: force, Interactive: isInteractive()}
	_, err = src.Sync(ctx, ref, dest, opts)
	return err
}

func notionClientFromConfig(c *appContext) *srcnotion.Client {
	token := c.Config.NotionToken()
	if token == "" {
		return nil
	}
	return &srcnotion.Client{Token: token}
}
