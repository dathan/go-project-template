// Package db defines the Store interface — the single point of contact between
// application code and the database.  Business logic depends on this interface,
// not on GORM or any concrete driver.
package db

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/dathan/go-project-template/internal/db/models"
)

// UserRepository defines all user persistence operations.
type UserRepository interface {
	Create(ctx context.Context, u *models.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*models.User, error)
	GetByEmail(ctx context.Context, email string) (*models.User, error)
	GetByProvider(ctx context.Context, provider, providerID string) (*models.User, error)
	List(ctx context.Context, opts models.ListOptions) ([]*models.User, int64, error)
	Update(ctx context.Context, u *models.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	SetPaidAt(ctx context.Context, id uuid.UUID, t time.Time) error
}

// SessionRepository manages server-side sessions.
type SessionRepository interface {
	Create(ctx context.Context, s *models.Session) error
	GetByToken(ctx context.Context, token string) (*models.Session, error)
	DeleteByToken(ctx context.Context, token string) error
	DeleteExpired(ctx context.Context) error
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
}

// PaymentRepository manages Stripe payment records.
type PaymentRepository interface {
	Create(ctx context.Context, p *models.Payment) error
	GetByStripeID(ctx context.Context, stripePaymentID string) (*models.Payment, error)
	ListByUser(ctx context.Context, userID uuid.UUID, opts models.ListOptions) ([]*models.Payment, error)
	UpdateStatus(ctx context.Context, stripePaymentID, status string) error
}

// Store is the top-level data access interface.  All repositories are accessed
// through it.  Raw SQL is always available as an escape hatch.
type Store interface {
	Users()    UserRepository
	Sessions() SessionRepository
	Payments() PaymentRepository

	// Exec runs arbitrary SQL that returns no rows (INSERT, UPDATE, DELETE).
	Exec(ctx context.Context, sql string, args ...any) error
	// QueryInto runs a raw SELECT and scans the result into dest (must be a pointer to a slice or struct).
	QueryInto(ctx context.Context, dest any, sql string, args ...any) error
	// Transaction runs fn inside a DB transaction; rolls back on error.
	Transaction(ctx context.Context, fn func(Store) error) error

	Close() error
}
