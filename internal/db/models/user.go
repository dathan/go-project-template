package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// User represents an authenticated user. Provider + ProviderID are the OAuth
// identity. Email is the canonical unique key for merging accounts.
type User struct {
	ID         uuid.UUID  `gorm:"type:uuid;primaryKey"`
	Email      string     `gorm:"uniqueIndex;not null"`
	Name       string
	AvatarURL  string
	Provider   string     `gorm:"not null"` // google | github | slack | linkedin
	ProviderID string     `gorm:"not null"`
	Role       string     `gorm:"default:'user'"`
	PaidAt     *time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (u *User) IsAdmin() bool { return u.Role == RoleAdmin }
func (u *User) IsPaid() bool  { return u.PaidAt != nil }
