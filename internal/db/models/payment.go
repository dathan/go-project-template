package models

import (
	"time"

	"github.com/google/uuid"
)

const (
	PaymentStatusPending   = "pending"
	PaymentStatusSucceeded = "succeeded"
	PaymentStatusFailed    = "failed"
)

// Payment records a Stripe payment intent linked to a user.
type Payment struct {
	ID              uuid.UUID `gorm:"type:uuid;primaryKey"`
	UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
	StripePaymentID string    `gorm:"uniqueIndex;not null"`
	Amount          int64     // in cents
	Currency        string    `gorm:"default:'usd'"`
	Status          string    `gorm:"default:'pending'"`
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
