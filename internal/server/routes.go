package server

import (
	"io/fs"
	"net/http"

	"github.com/dathan/go-project-template/internal/auth"
	"github.com/dathan/go-project-template/internal/config"
	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/server/handlers"
	"github.com/dathan/go-project-template/internal/server/middleware"
	"github.com/dathan/go-project-template/pkg"
)

// buildMux registers all routes and returns the root http.Handler.
func buildMux(
	cfg *config.Config,
	store db.Store,
	jwtSvc *auth.JWTService,
	sessionSvc *auth.SessionService,
	oauthRegistry *auth.Registry,
	agent *pkg.Agent,
	frontendAssets fs.FS,
) http.Handler {
	mux := http.NewServeMux()

	// ── Instantiate handlers ──────────────────────────────────────────────────
	authH := handlers.NewAuthHandler(oauthRegistry, jwtSvc, sessionSvc, store)
	adminH := handlers.NewAdminHandler(store, jwtSvc)
	payH := handlers.NewPaymentHandler(store, cfg.Stripe.SecretKey, cfg.Stripe.WebhookSecret)
	agentH := handlers.NewAgentHandler(agent)

	// ── Middleware factories ──────────────────────────────────────────────────
	requireAuth := middleware.Auth(jwtSvc, sessionSvc, store)

	// ── Public routes ─────────────────────────────────────────────────────────
	mux.HandleFunc("GET /healthz", handlers.Health)

	mux.HandleFunc("GET /auth/{provider}", authH.Redirect)
	mux.HandleFunc("GET /auth/{provider}/callback", authH.Callback)
	mux.HandleFunc("POST /auth/logout", authH.Logout)

	// Stripe webhook — raw body, no auth (verified by signature)
	mux.HandleFunc("POST /api/v1/webhooks/stripe", payH.Webhook)

	// ── Authenticated routes ──────────────────────────────────────────────────
	mux.Handle("GET /api/v1/me",
		middleware.Chain(http.HandlerFunc(authH.Me), requireAuth))

	mux.Handle("POST /api/v1/payments/intent",
		middleware.Chain(http.HandlerFunc(payH.CreateIntent), requireAuth))

	mux.Handle("POST /api/v1/agent/prompt",
		middleware.Chain(http.HandlerFunc(agentH.Prompt), requireAuth))

	mux.Handle("GET /api/v1/agent/stream",
		middleware.Chain(http.HandlerFunc(agentH.Stream), requireAuth))

	// ── Admin routes ──────────────────────────────────────────────────────────
	mux.Handle("GET /api/v1/admin/users",
		middleware.Chain(http.HandlerFunc(adminH.ListUsers), requireAuth, middleware.Admin))

	mux.Handle("POST /api/v1/admin/users/{id}/assume",
		middleware.Chain(http.HandlerFunc(adminH.AssumeUser), requireAuth, middleware.Admin))

	// ── Frontend (SPA catch-all, must be last) ────────────────────────────────
	if frontendAssets != nil {
		mux.Handle("/", frontendHandler(frontendAssets))
	}

	// ── Apply global middleware ───────────────────────────────────────────────
	return middleware.Chain(
		mux,
		middleware.Logging,
		middleware.CORS("*"),
	)
}
