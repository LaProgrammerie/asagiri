package api

import (
	"context"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// DefaultSocketPath returns the default Unix socket path for a repo.
func DefaultSocketPath(repoRoot string) string {
	return filepath.Join(repoRoot, runtime.DefaultRelDir, "runtime.sock")
}

// ServeUnix starts the same REST handler on a Unix domain socket (spec-my-A §24.18).
func ServeUnix(ctx context.Context, opts Options) error {
	if opts.SocketPath == "" {
		opts.SocketPath = DefaultSocketPath(opts.RepoRoot)
	}
	if err := os.MkdirAll(filepath.Dir(opts.SocketPath), 0o755); err != nil {
		return err
	}
	_ = os.Remove(opts.SocketPath)

	store, err := runtime.Open(opts.RepoRoot)
	if err != nil {
		return err
	}
	defer func() { _ = store.Close() }()

	token := opts.Token
	if token == "" {
		token, err = LoadToken(opts.RepoRoot)
		if err != nil {
			return err
		}
	}

	ln, err := net.Listen("unix", opts.SocketPath)
	if err != nil {
		return err
	}
	srv := &http.Server{
		Handler:           authMiddleware(token, NewServer(store).Handler()),
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
		_ = ln.Close()
		_ = os.Remove(opts.SocketPath)
	}()
	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}
