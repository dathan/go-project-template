package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dathan/go-project-template/internal/db/models"
)

// PaymentRepo is the exported GORM-backed implementation of PaymentRepository.
type PaymentRepo struct{ db *gorm.DB }

func NewPaymentRepo(db *gorm.DB) *PaymentRepo { return &PaymentRepo{db: db} }

func (r *PaymentRepo) Create(ctx context.Context, p *models.Payment) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *PaymentRepo) GetByStripeID(ctx context.Context, stripePaymentID string) (*models.Payment, error) {
	var p models.Payment
	if err := r.db.WithContext(ctx).First(&p, "stripe_payment_id = ?", stripePaymentID).Error; err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *PaymentRepo) ListByUser(ctx context.Context, userID uuid.UUID, opts models.ListOptions) ([]*models.Payment, error) {
	var payments []*models.Payment
	order := "created_at DESC"
	if opts.Order != "" {
		order = opts.Order
	}
	limit := 20
	if opts.Limit > 0 {
		limit = opts.Limit
	}
	if err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order(order).Limit(limit).Offset(opts.Offset).
		Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *PaymentRepo) UpdateStatus(ctx context.Context, stripePaymentID, status string) error {
	return r.db.WithContext(ctx).
		Model(&models.Payment{}).
		Where("stripe_payment_id = ?", stripePaymentID).
		Update("status", status).Error
}
