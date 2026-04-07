package repositories

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dathan/go-project-template/internal/db/models"
)

const queryByID = "id = ?"

// UserRepo is the exported GORM-backed implementation of UserRepository.
type UserRepo struct{ db *gorm.DB }

func NewUserRepo(db *gorm.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *models.User) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(u).Error
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, queryByID, id).Error; err != nil {
		return nil, fmt.Errorf("user %s: %w", id, err)
	}
	return &u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).First(&u, "email = ?", email).Error; err != nil {
		return nil, fmt.Errorf("user email %s: %w", email, err)
	}
	return &u, nil
}

func (r *UserRepo) GetByProvider(ctx context.Context, provider, providerID string) (*models.User, error) {
	var u models.User
	if err := r.db.WithContext(ctx).
		First(&u, "provider = ? AND provider_id = ?", provider, providerID).Error; err != nil {
		return nil, fmt.Errorf("user provider=%s id=%s: %w", provider, providerID, err)
	}
	return &u, nil
}

func (r *UserRepo) List(ctx context.Context, opts models.ListOptions) ([]*models.User, int64, error) {
	var users []*models.User
	var total int64

	q := r.db.WithContext(ctx).Model(&models.User{})
	q.Count(&total)

	order := "created_at DESC"
	if opts.Order != "" {
		order = opts.Order
	}
	limit := 20
	if opts.Limit > 0 {
		limit = opts.Limit
	}

	if err := q.Order(order).Limit(limit).Offset(opts.Offset).Find(&users).Error; err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (r *UserRepo) Update(ctx context.Context, u *models.User) error {
	return r.db.WithContext(ctx).Save(u).Error
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.User{}, queryByID, id).Error
}

func (r *UserRepo) SetPaidAt(ctx context.Context, id uuid.UUID, t time.Time) error {
	return r.db.WithContext(ctx).
		Model(&models.User{}).
		Where(queryByID, id).
		Update("paid_at", t).Error
}
