package postgres

import (
    "context"
    "time"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/dathan/go-project-template/internal/model"
)

// UserRepository defines operations for persisting and retrieving users
// from PostgreSQL. It hides the underlying SQL and provides methods
// friendly to higher layers like services and handlers.
type UserRepository interface {
    GetByProviderID(ctx context.Context, provider, providerID string) (*model.User, error)
    CreateOrUpdate(ctx context.Context, user *model.User) (*model.User, error)
}

// userRepo is a concrete implementation of UserRepository that uses
// pgxpool.Pool for PostgreSQL access. The struct is unexported to
// discourage direct instantiation; use NewUserRepository to construct.
type userRepo struct {
    db *pgxpool.Pool
}

// NewUserRepository returns a new userRepo bound to the given database pool.
func NewUserRepository(db *pgxpool.Pool) UserRepository {
    return &userRepo{db: db}
}

// GetByProviderID returns a user matching the provider and providerID,
// or nil if no such user exists. An error is returned on database failures.
func (r *userRepo) GetByProviderID(ctx context.Context, provider, providerID string) (*model.User, error) {
    const query = `SELECT id, provider, provider_id, email, name, created_at, updated_at
                   FROM users WHERE provider=$1 AND provider_id=$2 LIMIT 1`
    row := r.db.QueryRow(ctx, query, provider, providerID)
    var u model.User
    err := row.Scan(&u.ID, &u.Provider, &u.ProviderID, &u.Email, &u.Name, &u.CreatedAt, &u.UpdatedAt)
    if err != nil {
        if err == pgx.ErrNoRows {
            return nil, nil
        }
        return nil, err
    }
    return &u, nil
}

// CreateOrUpdate inserts a new user or updates an existing user. The
// combination of provider and providerID must be unique in the users
// table. When updating, the existing ID is preserved. On insert, a new
// UUID is generated. The returned user reflects the persisted state.
func (r *userRepo) CreateOrUpdate(ctx context.Context, user *model.User) (*model.User, error) {
    tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
    if err != nil {
        return nil, err
    }
    defer func() { _ = tx.Rollback(ctx) }()
    existing, err := r.GetByProviderID(ctx, user.Provider, user.ProviderID)
    if err != nil {
        return nil, err
    }
    now := time.Now()
    if existing == nil {
        // Insert new user
        id := uuid.NewString()
        const insertQuery = `INSERT INTO users
            (id, provider, provider_id, email, name, created_at, updated_at)
            VALUES ($1, $2, $3, $4, $5, $6, $7)`
        _, err = tx.Exec(ctx, insertQuery,
            id, user.Provider, user.ProviderID, user.Email, user.Name, now, now)
        if err != nil {
            return nil, err
        }
        user.ID = id
        user.CreatedAt = now
        user.UpdatedAt = now
    } else {
        // Update existing user
        const updateQuery = `UPDATE users
            SET email=$1, name=$2, updated_at=$3
            WHERE provider=$4 AND provider_id=$5`
        _, err = tx.Exec(ctx, updateQuery,
            user.Email, user.Name, now, user.Provider, user.ProviderID)
        if err != nil {
            return nil, err
        }
        user.ID = existing.ID
        user.CreatedAt = existing.CreatedAt
        user.UpdatedAt = now
    }
    if err := tx.Commit(ctx); err != nil {
        return nil, err
    }
    return user, nil
}