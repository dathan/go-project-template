package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/dathan/go-project-template/internal/auth"
	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/db/models"
	"github.com/dathan/go-project-template/internal/server/middleware"
)

// AuthHandler wires together OAuth, JWT, sessions, and the user store.
type AuthHandler struct {
	registry   *auth.Registry
	jwtSvc     *auth.JWTService
	sessionSvc *auth.SessionService
	store      db.Store
	// stateStore is an in-process CSRF state map (replace with Redis for multi-instance).
	stateStore map[string]string
}

func NewAuthHandler(
	registry *auth.Registry,
	jwtSvc *auth.JWTService,
	sessionSvc *auth.SessionService,
	store db.Store,
) *AuthHandler {
	return &AuthHandler{
		registry:   registry,
		jwtSvc:     jwtSvc,
		sessionSvc: sessionSvc,
		store:      store,
		stateStore: make(map[string]string),
	}
}

// Redirect starts the OAuth flow for the given provider.
// GET /auth/{provider}
func (h *AuthHandler) Redirect(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	p, err := h.registry.Get(provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	url, state, err := p.AuthCodeURL()
	if err != nil {
		http.Error(w, "failed to generate oauth url", http.StatusInternalServerError)
		return
	}

	// Persist state for CSRF validation (scoped per-request; keyed by state token)
	h.stateStore[state] = provider

	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   300,
	})
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Callback handles the provider redirect after user authorizes.
// GET /auth/{provider}/callback
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
	provider := r.PathValue("provider")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	// Validate CSRF state
	stateCookie, err := r.Cookie("oauth_state")
	if err != nil || stateCookie.Value != state {
		http.Error(w, "invalid oauth state", http.StatusBadRequest)
		return
	}
	http.SetCookie(w, &http.Cookie{Name: "oauth_state", MaxAge: -1, Path: "/"})

	p, err := h.registry.Get(provider)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	info, err := p.Exchange(r.Context(), code)
	if err != nil {
		http.Error(w, fmt.Sprintf("oauth exchange failed: %v", err), http.StatusInternalServerError)
		return
	}

	user, err := h.upsertUser(r, info)
	if err != nil {
		http.Error(w, "failed to create/update user", http.StatusInternalServerError)
		return
	}

	// Create server-side session (browser) and sign a JWT (API clients).
	token, err := h.sessionSvc.Create(r.Context(), user.ID)
	if err != nil {
		http.Error(w, "failed to create session", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     auth.CookieName(),
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	// Also issue a JWT for SPA / API clients.
	jwt, err := h.jwtSvc.Sign(user.ID, user.Role)
	if err != nil {
		http.Error(w, "failed to sign jwt", http.StatusInternalServerError)
		return
	}

	// Redirect to frontend with JWT in fragment (#token=...) — never in query string.
	http.Redirect(w, r, fmt.Sprintf("/#token=%s", jwt), http.StatusTemporaryRedirect)
}

// Logout deletes the session and clears the cookie.
// POST /auth/logout
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.CookieName())
	if err == nil {
		_ = h.sessionSvc.Delete(r.Context(), cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{Name: auth.CookieName(), MaxAge: -1, Path: "/"})
	writeJSON(w, http.StatusOK, map[string]string{"message": "logged out"})
}

// Me returns the currently authenticated user.
// GET /api/v1/me
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.UserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) upsertUser(r *http.Request, info *auth.UserInfo) (*models.User, error) {
	ctx := r.Context()
	existing, err := h.store.Users().GetByProvider(ctx, info.Provider, info.ProviderID)
	if err == nil {
		// Update mutable fields on each login
		existing.Name = info.Name
		existing.AvatarURL = info.AvatarURL
		existing.Email = info.Email
		return existing, h.store.Users().Update(ctx, existing)
	}

	// New user
	user := &models.User{
		Email:      info.Email,
		Name:       info.Name,
		AvatarURL:  info.AvatarURL,
		Provider:   info.Provider,
		ProviderID: info.ProviderID,
		Role:       models.RoleUser,
	}
	return user, h.store.Users().Create(ctx, user)
}
