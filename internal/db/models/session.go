package models

import (
	"time"

	"github.com/google/uuid"
)

// Session is a server-side browser session backed by a signed cookie token.
type Session struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"not null;index"`
	CreatedAt time.Time
}

func (s *Session) IsExpired() bool { return time.Now().After(s.ExpiresAt) }
