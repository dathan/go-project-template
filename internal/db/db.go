package db

import (
	"context"
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/dathan/go-project-template/internal/config"
	"github.com/dathan/go-project-template/internal/db/models"
	"github.com/dathan/go-project-template/internal/db/repositories"
)

// GORMStore implements Store using GORM + Postgres.
type GORMStore struct {
	db *gorm.DB
}

// New opens a Postgres connection via GORM and auto-migrates models.
func New(cfg config.DatabaseConfig) (*GORMStore, error) {
	gormCfg := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	}

	db, err := gorm.Open(postgres.Open(cfg.DSN()), gormCfg)
	if err != nil {
		return nil, fmt.Errorf("opening postgres: %w", err)
	}

	if err := db.AutoMigrate(
		&models.User{},
		&models.Session{},
		&models.Payment{},
	); err != nil {
		return nil, fmt.Errorf("automigrate: %w", err)
	}

	return &GORMStore{db: db}, nil
}

// DB returns the underlying *gorm.DB for callers that need direct access
// (e.g. running raw queries beyond the Store interface).
func (s *GORMStore) DB() *gorm.DB { return s.db }

func (s *GORMStore) Users() UserRepository {
	return repositories.NewUserRepo(s.db)
}

func (s *GORMStore) Sessions() SessionRepository {
	return repositories.NewSessionRepo(s.db)
}

func (s *GORMStore) Payments() PaymentRepository {
	return repositories.NewPaymentRepo(s.db)
}

func (s *GORMStore) Exec(ctx context.Context, sql string, args ...any) error {
	return s.db.WithContext(ctx).Exec(sql, args...).Error
}

func (s *GORMStore) QueryInto(ctx context.Context, dest any, sql string, args ...any) error {
	return s.db.WithContext(ctx).Raw(sql, args...).Scan(dest).Error
}

func (s *GORMStore) Transaction(ctx context.Context, fn func(Store) error) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&GORMStore{db: tx})
	})
}

func (s *GORMStore) Close() error {
	sqlDB, err := s.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
