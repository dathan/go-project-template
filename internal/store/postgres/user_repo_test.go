package postgres

import (
    "context"
    "os"
    "testing"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/dathan/go-project-template/internal/model"
)

// TestCreateAndGet verifies that a user can be inserted and retrieved by
// provider and provider ID. This test requires a PostgreSQL database to be
// available at TEST_DATABASE_URL. If the variable is not set the test is
// skipped.
func TestCreateAndGet(t *testing.T) {
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL not set; skipping integration test")
    }
    ctx := context.Background()
    pool, err := pgxpool.New(ctx, dsn)
    if err != nil {
        t.Fatalf("connect: %v", err)
    }
    defer pool.Close()
    repo := NewUserRepository(pool)
    user := &model.User{
        Provider:   "google",
        ProviderID: uuid.NewString(),
        Email:      "test@example.com",
        Name:       "Test User",
    }
    // Insert
    saved, err := repo.CreateOrUpdate(ctx, user)
    if err != nil {
        t.Fatalf("CreateOrUpdate: %v", err)
    }
    if saved.ID == "" {
        t.Fatalf("expected ID set")
    }
    // Get
    fetched, err := repo.GetByProviderID(ctx, user.Provider, user.ProviderID)
    if err != nil {
        t.Fatalf("GetByProviderID: %v", err)
    }
    if fetched == nil {
        t.Fatalf("expected user, got nil")
    }
    if fetched.Email != user.Email || fetched.Name != user.Name {
        t.Fatalf("expected %v, got %v", user, fetched)
    }
}