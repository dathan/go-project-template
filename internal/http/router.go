package http

import (
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/go-chi/chi/v5/middleware"
)

// NewRouter constructs a chi.Router and registers all HTTP handlers. It
// configures basic middleware such as logging and recovery. Additional
// middleware (authentication, metrics, etc.) can be added here.
func NewRouter(authHandler *AuthHandler) http.Handler {
    r := chi.NewRouter()
    // Basic middleware
    r.Use(middleware.Logger)
    r.Use(middleware.Recoverer)

    // Health endpoint
    r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    })

    // API routes
    r.Route("/api/v1/auth", func(r chi.Router) {
        r.Get("/login", authHandler.Login)
        r.Get("/callback", authHandler.Callback)
    })
    return r
}