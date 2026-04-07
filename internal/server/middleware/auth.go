package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/google/uuid"

	"github.com/dathan/go-project-template/internal/auth"
	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/db/models"
)

type contextKey string

const (
	ContextKeyUser    contextKey = "user"
	ContextKeyClaims  contextKey = "claims"
)

// UserFromContext retrieves the authenticated user from the request context.
func UserFromContext(ctx context.Context) *models.User {
	u, _ := ctx.Value(ContextKeyUser).(*models.User)
	return u
}

// ClaimsFromContext retrieves the JWT claims from the request context.
func ClaimsFromContext(ctx context.Context) *auth.Claims {
	c, _ := ctx.Value(ContextKeyClaims).(*auth.Claims)
	return c
}

// Auth validates either a Bearer JWT or a session cookie, then loads the user
// from the database and injects it into the request context.
func Auth(jwtSvc *auth.JWTService, sessionSvc *auth.SessionService, store db.Store) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user, claims, err := resolveIdentity(r, jwtSvc, sessionSvc, store)
			if err != nil || user == nil {
				http.Error(w, "unauthorized", http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), ContextKeyUser, user)
			ctx = context.WithValue(ctx, ContextKeyClaims, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Admin gates a handler to admin-role users only.
func Admin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := UserFromContext(r.Context())
		if user == nil || !user.IsAdmin() {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// resolveIdentity tries Bearer JWT first, then session cookie.
func resolveIdentity(
	r *http.Request,
	jwtSvc *auth.JWTService,
	sessionSvc *auth.SessionService,
	store db.Store,
) (*models.User, *auth.Claims, error) {
	// 1. Bearer token
	if bearer := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer "); bearer != r.Header.Get("Authorization") {
		claims, err := jwtSvc.Verify(bearer)
		if err != nil {
			return nil, nil, err
		}
		user, err := store.Users().GetByID(r.Context(), claims.UserID)
		if err != nil {
			return nil, nil, err
		}
		return user, claims, nil
	}

	// 2. Session cookie
	cookie, err := r.Cookie(auth.CookieName())
	if err != nil {
		return nil, nil, err
	}
	sess, err := sessionSvc.Validate(r.Context(), cookie.Value)
	if err != nil {
		return nil, nil, err
	}
	user, err := store.Users().GetByID(r.Context(), sess.UserID)
	if err != nil {
		return nil, nil, err
	}
	// Synthesize minimal claims for session-based auth
	claims := &auth.Claims{
		UserID: user.ID,
		Role:   user.Role,
	}
	return user, claims, nil
}

// MustGetUserID is a helper for handlers to extract the user ID without nil checks.
func MustGetUserID(ctx context.Context) uuid.UUID {
	u := UserFromContext(ctx)
	if u == nil {
		return uuid.Nil
	}
	return u.ID
}
