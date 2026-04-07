// Package server wires the HTTP server together without leaking any framework
// types into application or business logic code.
package server

import (
	"context"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"time"

	"github.com/dathan/go-project-template/internal/auth"
	"github.com/dathan/go-project-template/internal/config"
	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/pkg"
)

// Server wraps http.Server with graceful-shutdown support.
type Server struct {
	httpServer *http.Server
}

// New builds a Server from its dependencies.
func New(
	cfg *config.Config,
	store db.Store,
	jwtSvc *auth.JWTService,
	sessionSvc *auth.SessionService,
	oauthRegistry *auth.Registry,
	agent *pkg.Agent,
	frontendAssets fs.FS,
) *Server {
	handler := buildMux(cfg, store, jwtSvc, sessionSvc, oauthRegistry, agent, frontendAssets)
	return &Server{
		httpServer: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
			Handler:      handler,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 30 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
	}
}

// ListenAndServe starts the server. It blocks until the context is cancelled,
// then performs a graceful shutdown.
//
// We explicitly create a tcp4 listener (AF_INET) instead of relying on
// http.Server.ListenAndServe, which on macOS resolves "0.0.0.0:port" to an
// AF_INET6 dual-stack socket. When another process already holds the AF_INET
// wildcard (e.g. an SSH port-forward), that process silently intercepts all
// IPv4 connections while the server appears to start normally. Using tcp4
// makes the bind fail fast with EADDRINUSE in that scenario instead.
func (s *Server) ListenAndServe(ctx context.Context) error {
	ln, err := net.Listen("tcp4", s.httpServer.Addr)
	if err != nil {
		return fmt.Errorf("listening on %s: %w", s.httpServer.Addr, err)
	}

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.Serve(ln); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		return s.httpServer.Shutdown(shutdownCtx)
	}
}

func (s *Server) Addr() string { return s.httpServer.Addr }
