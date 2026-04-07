package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/dathan/go-project-template/internal/db"
	"github.com/dathan/go-project-template/internal/db/models"
)

const sessionCookieName = "session"

// SessionService manages server-side sessions stored in Postgres.
type SessionService struct {
	store    db.Store
	duration time.Duration
}

func NewSessionService(store db.Store, duration time.Duration) *SessionService {
	return &SessionService{store: store, duration: duration}
}

// Create generates a random token, persists the session, and returns the token.
func (s *SessionService) Create(ctx context.Context, userID uuid.UUID) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating session token: %w", err)
	}
	token := hex.EncodeToString(b)

	sess := &models.Session{
		UserID:    userID,
		Token:     token,
		ExpiresAt: time.Now().Add(s.duration),
	}
	if err := s.store.Sessions().Create(ctx, sess); err != nil {
		return "", fmt.Errorf("storing session: %w", err)
	}
	return token, nil
}

// Validate returns the session if the token is valid and not expired.
func (s *SessionService) Validate(ctx context.Context, token string) (*models.Session, error) {
	sess, err := s.store.Sessions().GetByToken(ctx, token)
	if err != nil {
		return nil, err
	}
	if sess.IsExpired() {
		_ = s.store.Sessions().DeleteByToken(ctx, token)
		return nil, fmt.Errorf("session expired")
	}
	return sess, nil
}

// Delete removes a session (logout).
func (s *SessionService) Delete(ctx context.Context, token string) error {
	return s.store.Sessions().DeleteByToken(ctx, token)
}

// CookieName is the HTTP cookie name used for sessions.
func CookieName() string { return sessionCookieName }
