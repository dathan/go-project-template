package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dathan/go-project-template/internal/auth"
	"github.com/dathan/go-project-template/internal/config"
	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/server"
	"github.com/dathan/go-project-template/pkg"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("loading config", "err", err)
		os.Exit(1)
	}

	store, err := db.New(cfg.Database)
	if err != nil {
		slog.Error("connecting to database", "err", err)
		os.Exit(1)
	}
	defer store.Close()

	jwtSvc := auth.NewJWTService(cfg.Auth.JWTSecret, cfg.Auth.SessionDuration)
	sessionSvc := auth.NewSessionService(store, cfg.Auth.SessionDuration)
	oauthRegistry := auth.NewRegistry(cfg.Auth.OAuth)

	agent, err := pkg.NewAgent(cfg.Agent.Provider)
	if err != nil {
		slog.Warn("agent init failed, agent endpoints will error", "err", err)
		agent = nil
	}

	srv := server.New(cfg, store, jwtSvc, sessionSvc, oauthRegistry, agent, nil)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	slog.Info("server starting", "addr", srv.Addr())
	if err := srv.ListenAndServe(ctx); err != nil {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
