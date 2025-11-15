package http

import (
    "encoding/json"
    "net/http"
    "time"

    "github.com/dathan/go-project-template/internal/service"
)

// AuthHandler wires HTTP requests to the AuthService. It is responsible
// for extracting parameters from the request, invoking service methods
// and writing appropriate responses. This handler does not perform any
// business logic itself.
type AuthHandler struct {
    svc *service.AuthService
}

// NewAuthHandler returns a new AuthHandler.
func NewAuthHandler(svc *service.AuthService) *AuthHandler {
    return &AuthHandler{svc: svc}
}

// Login responds with the provider-specific login URL. The frontend
// should read this URL and redirect the user to the provider. The
// provider value must be supplied via the "provider" query parameter.
// Example: GET /api/v1/auth/login?provider=google
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
    provider := r.URL.Query().Get("provider")
    if provider == "" {
        http.Error(w, "missing provider", http.StatusBadRequest)
        return
    }
    url, err := h.svc.GetLoginURL(provider)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }
    resp := map[string]string{"url": url}
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(resp)
}

// Callback handles the OAuth callback. It exchanges the code for a token,
// retrieves user info and creates a session. The provider and code
// parameters should be supplied via query parameters. On success, a
// session cookie is written and the user is redirected to the root.
// Example: GET /api/v1/auth/callback?provider=google&code=...
func (h *AuthHandler) Callback(w http.ResponseWriter, r *http.Request) {
    ctx := r.Context()
    provider := r.URL.Query().Get("provider")
    code := r.URL.Query().Get("code")
    if provider == "" || code == "" {
        http.Error(w, "missing provider or code", http.StatusBadRequest)
        return
    }
    sessionID, err := h.svc.HandleCallback(ctx, provider, code)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    // set session cookie
    cookie := &http.Cookie{
        Name:     "session_id",
        Value:    sessionID,
        Path:     "/",
        HttpOnly: true,
        Secure:   r.TLS != nil,
        Expires:  time.Now().Add(h.svc.SessionTTL()),
    }
    http.SetCookie(w, cookie)
    // redirect to root or a success page
    http.Redirect(w, r, "/", http.StatusSeeOther)
}

// SessionTTL exposes the configured session lifetime. It is used for
// setting cookie expiration from within the handler. The value is
// returned from the AuthService.
// Note: SessionTTL is defined on AuthService in the service layer.