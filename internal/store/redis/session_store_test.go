package redisstore

import (
    "context"
    "os"
    "testing"
    "time"

    "github.com/go-redis/redis/v8"
)

// TestSessionStore verifies that session data can be stored and retrieved
// from Redis. It requires a Redis server reachable at TEST_REDIS_ADDR.
func TestSessionStore(t *testing.T) {
    addr := os.Getenv("TEST_REDIS_ADDR")
    if addr == "" {
        t.Skip("TEST_REDIS_ADDR not set; skipping integration test")
    }
    client := redis.NewClient(&redis.Options{Addr: addr})
    defer client.Close()
    store := NewSessionStore(client)
    ctx := context.Background()
    sid := "test-session"
    data := SessionData{UserID: "user123", Provider: "google", TokenData: map[string]string{"access_token": "abc"}}
    if err := store.Set(ctx, sid, data, time.Minute); err != nil {
        t.Fatalf("Set: %v", err)
    }
    got, err := store.Get(ctx, sid)
    if err != nil {
        t.Fatalf("Get: %v", err)
    }
    if got == nil || got.UserID != data.UserID || got.Provider != data.Provider {
        t.Fatalf("unexpected session data: %v", got)
    }
    if err := store.Delete(ctx, sid); err != nil {
        t.Fatalf("Delete: %v", err)
    }
    gone, err := store.Get(ctx, sid)
    if err != nil {
        t.Fatalf("Get after delete: %v", err)
    }
    if gone != nil {
        t.Fatalf("expected nil after delete, got %v", gone)
    }
}