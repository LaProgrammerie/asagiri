package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/LaProgrammerie/asagiri/application/internal/runtime"
)

// Options configures the local runtime HTTP API (spec-my-A §24.18).
type Options struct {
	RepoRoot   string
	Port       int
	SocketPath string
	Token      string
}

// Serve starts a JSON REST server bound to 127.0.0.1 only.
func Serve(ctx context.Context, opts Options) error {
	if opts.Port <= 0 {
		opts.Port = 8765
	}
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

	srv := &http.Server{
		Addr:              fmt.Sprintf("127.0.0.1:%d", opts.Port),
		Handler:           authMiddleware(token, NewServer(store).Handler()),
		ReadHeaderTimeout: 5 * time.Second,
		BaseContext: func(net.Listener) context.Context {
			return ctx
		},
	}

	ln, err := net.Listen("tcp", srv.Addr)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdown)
	}()

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

// Addr returns the listen address for a port.
func Addr(port int) string {
	if port <= 0 {
		port = 8765
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}
