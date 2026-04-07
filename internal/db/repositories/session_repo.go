package repositories

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/dathan/go-project-template/internal/db/models"
)

// SessionRepo is the exported GORM-backed implementation of SessionRepository.
type SessionRepo struct{ db *gorm.DB }

func NewSessionRepo(db *gorm.DB) *SessionRepo { return &SessionRepo{db: db} }

func (r *SessionRepo) Create(ctx context.Context, s *models.Session) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return r.db.WithContext(ctx).Create(s).Error
}

func (r *SessionRepo) GetByToken(ctx context.Context, token string) (*models.Session, error) {
	var s models.Session
	if err := r.db.WithContext(ctx).First(&s, "token = ?", token).Error; err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *SessionRepo) DeleteByToken(ctx context.Context, token string) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "token = ?", token).Error
}

func (r *SessionRepo) DeleteExpired(ctx context.Context) error {
	return r.db.WithContext(ctx).
		Where("expires_at < NOW()").
		Delete(&models.Session{}).Error
}

func (r *SessionRepo) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&models.Session{}, "user_id = ?", userID).Error
}
